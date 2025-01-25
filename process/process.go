package process

import (
	"log"
	"os/exec"
	"sync"
)

type ProcessManager struct {
	programs map[string]*exec.Cmd
	mu       sync.Mutex
}

func NewProcessManager(cfg *Config) *ProcessManager {
	pm := &ProcessManager{programs: make(map[string]*exec.Cmd)}

	// Başlatılması gereken programları yükle
	for _, prog := range cfg.Programs {
		if prog.AutoStart {
			pm.StartProgram(prog)
		}
	}

	return pm
}

func (pm *ProcessManager) StartProgram(prog Program) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	cmd := exec.Command(prog.Cmd)
	cmd.Dir = prog.WorkingDir
	cmd.Env = prog.Env

	err := cmd.Start()
	if err != nil {
		log.Printf("Program başlatılamadı: %s, hata: %v\n", prog.Name, err)
		return
	}

	pm.programs[prog.Name] = cmd
	log.Printf("Program başlatıldı: %s\n", prog.Name)
}

func (pm *ProcessManager) StopProgram(name string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	cmd, ok := pm.programs[name]
	if ok && cmd.Process != nil {
		cmd.Process.Kill()
		delete(pm.programs, name)
		log.Printf("Program durduruldu: %s\n", name)
	}
}

func (pm *ProcessManager) StopAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for name, cmd := range pm.programs {
		if cmd.Process != nil {
			cmd.Process.Kill()
			log.Printf("Program durduruldu: %s\n", name)
		}
	}
	pm.programs = make(map[string]*exec.Cmd)
}
   