package process

import (
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
				p := &Process{cmd: exec.Command("sh", "-c", prog.Command), config: prog}
				p.cmd.Start()
				m.processes[name] = append(m.processes[name], p)
			}
		}
	}
}

func (m *Manager) Stop() {
	for _, procs := range m.processes {
		for _, p := range procs {
			p.cmd.Process.Kill()
		}
	}
}

func (m *Manager) UpdateConfig(newCfg config.Config) {
	// Yeni yapılandırmaya göre süreçleri güncelle
	m.config = newCfg
	// TODO: Değişenleri ekle/kaldır, değişmeyenleri koru
}