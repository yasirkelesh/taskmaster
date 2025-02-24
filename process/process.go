package process

import (
	"fmt"
	"os/exec"
	"taskmaster/config"
)

type Manager struct {
	config   config.Config
	processes map[string][]*Process
}

type Process struct {
	cmd    *exec.Cmd
	config config.Program
	state  string // "running", "stopped", "failed" gibi durumlar
}

func NewManager(cfg config.Config) *Manager {
	return &Manager{
		config:    cfg,
		processes: make(map[string][]*Process),
	}
}

func (m *Manager) Start() {
	for name, prog := range m.config.Programs {
		if prog.AutoStart {
			for i := 0; i < prog.NumProcs; i++ {
				p := &Process{
					cmd:    exec.Command("sh", "-c", prog.Command),
					config: prog,
					state:  "stopped",
				}
				err := p.cmd.Start()
				if err != nil {
					fmt.Printf("Süreç başlatılamadı (%s): %v\n", name, err)
					p.state = "failed"
				} else {
					p.state = "running"
					// Sürecin durumunu izlemek için goroutine başlat
					go func(p *Process) {
						p.cmd.Wait()
						p.state = "stopped"
					}(p)
				}
				m.processes[name] = append(m.processes[name], p)
			}
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
	// TODO: Yeni yapılandırmaya göre süreçleri güncelle
	m.config = newCfg
}

// Yeni ek: Süreç durumlarını döndüren bir yöntem
func (m *Manager) GetStatus() map[string][]string {
	status := make(map[string][]string)
	for name, procs := range m.processes {
		for _, p := range procs {
			status[name] = append(status[name], p.state)
		}
	}
	return status
}