package process

import (
	"fmt"
	"os/exec"
	"syscall"
	"sync/atomic"
	"taskmaster/config"
)

type Process struct {
	id 	  	  int64
	cmd       *exec.Cmd
	config    config.Program
	state     string // "running", "stopped", "failed"
	cancelCh  chan struct{}
}

type Manager struct {
	config    config.Config
	processes map[string][]*Process
	processIDCounter int64
}

// removeProcess removes a specific process from the slice of processes for a given program name
func (m *Manager) removeProcess(name string, proc *Process) {
	procs := m.processes[name]
	for i, p := range procs {
		if p == proc {
			m.processes[name] = append(procs[:i], procs[i+1:]...)
			p.cmd.Process.Kill()
			break
		}
	}
}



func NewManager(cfg config.Config) *Manager {
	return &Manager{
		config:    cfg,
		processes: make(map[string][]*Process),
	}
}

func (m *Manager) RestartProgram(name string) {
    // Önce yapılandırmayı kontrol et
    prog, exists := m.config.Programs[name]
    if !exists {
        fmt.Printf("'%s' adında bir program yok\n", name)
        return
    }

    // Mevcut süreçlerin bir kopyasını al
    var procs []*Process
    if existingProcs, ok := m.processes[name]; ok {
        procs = make([]*Process, len(existingProcs))
        copy(procs, existingProcs)
    }

    // Restart bayrağı ile özel bir stop işlemi yapalım
    // Bu sayede auto-restart tetiklenmeyecek
    for _, p := range procs {
        if p != nil && p.state == "running" {
            // Sürecin auto-restart işleminin atlamasını sağlayalım
            if p.cancelCh != nil {
                close(p.cancelCh)  // Mevcut izleme goroutine'ini kapat
                p.cancelCh = nil   // Kapatıldığını işaretle
            }
            
            // Süreci öldür
            if p.cmd != nil && p.cmd.Process != nil {
                err := p.cmd.Process.Kill()
                if err != nil {
                    fmt.Printf("Süreç durdurulurken hata: %v\n", err)
                }
            }
            
            // Süreci süreç listesinden kaldır
            m.removeProcess(name, p)
        }
    }
    
    // Tüm süreçleri temizle
    m.processes[name] = nil
    
    // Şimdi yeni süreçleri başlat
    m.processes[name] = make([]*Process, 0, prog.NumProcs)
    for i := 0; i < prog.NumProcs; i++ {
        p := m.startProcess(name, prog)
        m.processes[name] = append(m.processes[name], p)
    }
    
    fmt.Printf("'%s' programı yeniden başlatıldı\n", name)
}

func (m *Manager) Start() {
	for name, prog := range m.config.Programs {
		if prog.AutoStart {
			for i := 0; i < prog.NumProcs; i++ {
				p := m.startProcess(name, prog)
				m.processes[name] = append(m.processes[name], p)
			}
		}
	}
}

// Yeni ek: Belirli bir programı başlat
func (m *Manager) StartProgram(name string) error {
	prog, exists := m.config.Programs[name]
	if (!exists) {
		return fmt.Errorf("program '%s' yapılandırmada tanımlı değil", name)
	}
	// Zaten çalışan süreç sayısını kontrol et
	currentProcs := len(m.processes[name])
	if currentProcs >= prog.NumProcs {
		return fmt.Errorf("'%s' zaten maksimum süreç sayısında çalışıyor", name)
	}

	// Eksik süreçleri başlat
	for i := currentProcs; i < prog.NumProcs; i++ {
		p := m.startProcess(name, prog)
		m.processes[name] = append(m.processes[name], p)
	}
	return nil
}

func (m *Manager) startProcess(name string, prog config.Program) *Process {
	newID := atomic.AddInt64(&m.processIDCounter, 1)
	
	
	p := &Process{
		id:      newID,
		cmd:    exec.Command("sh", "-c", prog.Command),
		config: prog,
		state:  "stopped",
		cancelCh: make(chan struct{}),
	}
	err := p.cmd.Start()
	if err != nil {
		fmt.Printf("Süreç başlatılamadı (%s ID:%d): %v\n", name, p.id, err)
		p.state = "failed"
		return p
	}
	p.state = "running"

	fmt.Printf("Süreç başlatıldı: %s [ID:%d] (PID:%d)\n", name, p.id, p.cmd.Process.Pid)

	// Sürecin durumunu izle ve autorestart uygula
	go func(p *Process, cancelCh chan struct{}) {
		// İki kanalı birden dinle: cmd.Wait()'in tamamlanması ve iptal sinyali
		waitCh := make(chan error, 1)
		go func() {
			waitCh <- p.cmd.Wait()
		}()

		select {
		case <-cancelCh:
			// İzleme iptal edildi, reload nedeniyle
			fmt.Printf("%s için izleme goroutine'i konfigürasyon değişikliği nedeniyle sonlandırılıyor\n", name)
			return
		case err := <-waitCh:
			// Süreç sonlandı, normal işlemlere devam et
			p.state = "stopped"
			
			var exitCode int
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
						exitCode = status.ExitStatus()
					}
				}
			}

			switch p.config.AutoRestart {
			case "always":
				fmt.Printf("%s yeniden başlatılıyor (always politikası) pid %d\n", name, p.cmd.Process.Pid)
				m.removeProcess(name, p) // Eski süreci temizle
				m.StartProgram(name)			
			case "never":
				fmt.Printf("%s bitti, yeniden başlatılmayacak (never politikası)\n", name)
				if p.cmd.Process != nil {
					_ = p.cmd.Process.Kill() // Hata görmezden gelindi ama iyi bir pratikte işlenmelidir
				}
				m.removeProcess(name, p)
			case "unexpected":
				isExpected := false
				for _, code := range p.config.ExitCodes {
					if code == exitCode {
						isExpected = true
						break
					}
				}
				if !isExpected {
					fmt.Printf("%s beklenmedik çıkış (%d), yeniden başlatılıyor\n", name, exitCode)
					m.removeProcess(name, p) // Eski süreci temizle
					m.startProcess(name, p.config)
				} else {
					fmt.Printf("%s beklenen çıkış (%d), yeniden başlatılmadı\n", name, exitCode)
					if p.cmd.Process != nil {
						_ = p.cmd.Process.Kill()
					}
					m.removeProcess(name, p)
				}
			}
		}
	}(p, p.cancelCh)
	
	return p
}

func (m *Manager) StopProgram(name string) {
	procs, exists := m.processes[name]
	if !exists {
		fmt.Printf("'%s' adında bir program yok\n", name)
		return
	}

	for _, p := range procs {
		if p.state == "running" {
			p.cmd.Process.Kill()
			p.state = "stopped"
			m.processes[name] = nil
		}
	}
}

func (m *Manager) Stop() {
	for _, procs := range m.processes {
		for _, p := range procs {
			if p.state == "running" {
				p.cmd.Process.Kill()
				p.state = "stopped"
			}
		}
	}
}


func (m *Manager) UpdateConfig(newCfg config.Config) {
	m.config = newCfg
	// TODO: Yeni yapılandırmaya göre süreçleri güncelle
}

func (m *Manager) GetStatus() map[string][]string {
	status := make(map[string][]string)
	for name, procs := range m.processes {
		for _, p := range procs {
			status[name] = append(status[name], p.state)
		}
	}
	return status
}

func (m *Manager) StatusProgram(name string) {
    procs, exists := m.processes[name]
    if (!exists || len(procs) == 0) {
        fmt.Printf("%s: Çalışan süreç yok\n", name)
        return
    }
    
    for i, p := range procs {
        if p == nil {
            continue
        }
        var pid int
        if p.cmd != nil && p.cmd.Process != nil {
            pid = p.cmd.Process.Pid
        }
        fmt.Printf("%s[%d] (ID:%d, PID:%d): %s\n", name, i, p.id, pid, p.state)
    }
}

func (m *Manager) Status() {
    fmt.Println("Program         Status             PID       ID")
    fmt.Println("--------------------------------------------")
    
    for name, procs := range m.processes {
        if len(procs) == 0 {
            fmt.Printf("%-15s No processes\n", name)
            continue
        }
        
        for i, p := range procs {
            if p == nil {
                continue
            }
            var pid int
            if p.cmd != nil && p.cmd.Process != nil {
                pid = p.cmd.Process.Pid
            }
            fmt.Printf("%-15s[%d] %-10s     %-8d %-8d\n", name, i, p.state, pid, p.id)
        }
    }
}


