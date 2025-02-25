package process

import (
	"fmt"
	"os/exec"
	"syscall"
	"taskmaster/config"
)

type Manager struct {
	config    config.Config
	processes map[string][]*Process
}

type Process struct {
	cmd    *exec.Cmd
	config config.Program
	state  string // "running", "stopped", "failed"
}

func NewManager(cfg config.Config) *Manager {
	return &Manager{
		config:    cfg,
		processes: make(map[string][]*Process),
	}
}

func (m *Manager) RestartProgram(name string) {
	m.StopProgram(name)
	prog, exists := m.config.Programs[name]
	if !exists {
		fmt.Printf("'%s' adında bir program yok\n", name)
		return
	}

	for i := 0; i < prog.NumProcs; i++ {
		p := m.startProcess(name, prog)
		m.processes[name] = append(m.processes[name], p)
	}
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

func (m *Manager) startProcess(name string, prog config.Program) *Process {
	p := &Process{
		cmd:    exec.Command("sh", "-c", prog.Command),
		config: prog,
		state:  "stopped",
	}
	err := p.cmd.Start()
	if err != nil {
		fmt.Printf("Süreç başlatılamadı (%s): %v\n", name, err)
		p.state = "failed"
		return p
	}
	p.state = "running"

	// Sürecin durumunu izle ve autorestart uygula
	go func(p *Process) {
		err := p.cmd.Wait()
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
			fmt.Printf("%s yeniden başlatılıyor (always politikası)\n", name)
			m.startProcess(name, p.config)
		case "never":
			fmt.Printf("%s bitti, yeniden başlatılmayacak (never politikası)\n", name)
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
				m.startProcess(name, p.config)
			} else {
				fmt.Printf("%s beklenen çıkış (%d), yeniden başlatılmadı\n", name, exitCode)
			}
		}
	}(p)
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

// Yeni ek: Belirli bir programı başlat
func (m *Manager) StartProgram(name string) error {
	prog, exists := m.config.Programs[name]
	if !exists {
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
