package process

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"taskmaster/config"
	"taskmaster/logger"
	"time"
)

type Process struct {
	id       int64
	cmd      *exec.Cmd
	config   config.Program
	state    string // "running", "stopped", "failed"
	cancelCh chan struct{}
}

type Manager struct {
	config           config.Config
	processes        map[string][]*Process
	mutex            sync.RWMutex
	processIDCounter int64
}

// removeProcess removes a specific process from the slice of processes for a given program name
func (m *Manager) removeProcess(name string, proc *Process) {

	if proc.cmd != nil && proc.cmd.Process != nil {
		err := proc.cmd.Process.Kill()
		if err != nil {
			fmt.Printf("Süreç durdurulurken hata: %v\n", err)
			fmt.Print("taskmaster> ")
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	procs := m.processes[name]

	for i, p := range procs {
		if p == proc {
			m.processes[name] = append(procs[:i], procs[i+1:]...)
			//p.cmd.Process.Kill()
			break
		}
	}
}

func (m *Manager) removeProcessMap(name string, proc *Process) {

	m.mutex.Lock()
	defer m.mutex.Unlock()
	procs := m.processes[name]
	for i, p := range procs {
		if p == proc {
			m.processes[name] = append(procs[:i], procs[i+1:]...)
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
		fmt.Print("taskmaster> ")
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
				close(p.cancelCh) // Mevcut izleme goroutine'ini kapat
				p.cancelCh = nil  // Kapatıldığını işaretle
			}

			// Süreci öldür

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
	fmt.Print("taskmaster> ")
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
	if !exists {
		return fmt.Errorf("program '%s' yapılandırmada tanımlı değil", name)
	}
	// Zaten çalışan süreç sayısını kontrol et

	m.mutex.Lock()
	defer m.mutex.Unlock()
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
		id:       newID,
		cmd:      exec.Command("sh", "-c", prog.Command),
		config:   prog,
		state:    "stopped",
		cancelCh: make(chan struct{}),
	}

	// Stdout ve Stderr pipe'ları oluştur
	if prog.Stdout != "" {
		stdoutPipe, err := p.cmd.StdoutPipe()
		if err == nil {
			go m.logPipeOutput(stdoutPipe, prog.Stdout, name, "STDOUT")
		}
	}

	if prog.Stderr != "" {
		stderrPipe, err := p.cmd.StderrPipe()
		if err == nil {
			go m.logPipeOutput(stderrPipe, prog.Stderr, name, "STDERR")
		}
	}

	// Working directory ayarla
	if prog.WorkingDir != "" {
		p.cmd.Dir = prog.WorkingDir
	}

	// Environment variables ayarla
	if len(prog.Env) > 0 {
		env := os.Environ()
		for key, value := range prog.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		p.cmd.Env = env
	}

	err := p.cmd.Start()
	if err != nil {
		fmt.Printf("Süreç başlatılamadı (%s ID:%d): %v\n", name, p.id, err)
		fmt.Print("taskmaster> ")
		logger.LogError(prog.Stderr, fmt.Sprintf("Süreç başlatılamadı (%s ID:%d): %v", name, p.id, err))
		p.state = "failed"
		return p
	}
	p.state = "running"

	fmt.Printf("Süreç başlatıldı: %s [ID:%d] (PID:%d)\n", name, p.id, p.cmd.Process.Pid)
	fmt.Print("taskmaster> ")
	logger.LogInfo(prog.Stdout, fmt.Sprintf("Süreç başlatıldı: %s [ID:%d] (PID:%d)", name, p.id, p.cmd.Process.Pid))

	go m.monitorProcess(name, p)
	return p
}

// Yeni fonksiyon: Pipe çıktısını logla
func (m *Manager) logPipeOutput(pipe io.ReadCloser, logFile, programName, outputType string) {
	defer pipe.Close()

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			timestamp := time.Now().Format("2006/01/02 15:04:05")
			logMessage := fmt.Sprintf("%s [%s-%s] %s\n", timestamp, programName, outputType, line)

			// Dosyaya yaz
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				file.WriteString(logMessage)
				file.Close()
			}
		}
	}
}

func (m *Manager) monitorProcess(name string, p *Process) {
	waitCh := make(chan error, 1)
	go func() {
		waitCh <- p.cmd.Wait()
	}()

	select {
	case <-p.cancelCh:
		fmt.Printf("%s için izleme goroutine'i konfigürasyon değişikliği nedeniyle sonlandırılıyor\n", name)
		fmt.Print("taskmaster> ")
		logger.LogInfo(p.config.Stdout, fmt.Sprintf("%s için izleme goroutine'i sonlandırıldı", name))
		return
	case err := <-waitCh:
		var exitCode int
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					exitCode = status.ExitStatus()
				}
				fmt.Printf("%s sonlandı, çıkış kodu: %d\n", name, exitCode)
				fmt.Print("taskmaster> ")

				// Çıkış koduna göre log seviyesi belirle
				if isErrorExitCode(exitCode) {
					logger.LogError(p.config.Stderr, fmt.Sprintf("%s hata ile sonlandı, çıkış kodu: %d", name, exitCode))
				} else {
					logger.LogInfo(p.config.Stdout, fmt.Sprintf("%s sonlandı, çıkış kodu: %d", name, exitCode))
				}
			} else {
				fmt.Printf("%s başarıyla tamamlandı (çıkış kodu 0)\n", name)
				fmt.Print("taskmaster> ")
				logger.LogInfo(p.config.Stdout, fmt.Sprintf("%s başarıyla tamamlandı (çıkış kodu 0)", name))
			}
		}

		p.state = "stopped"
		m.handleAutoRestart(name, p, exitCode)
	}
}

// Hata çıkış kodlarını belirleyen yardımcı fonksiyon
func isErrorExitCode(exitCode int) bool {
	errorCodes := []int{
		126, // Command found but not executable
		127, // Command not found
		128, // Invalid argument to exit
		130, // Process terminated by Ctrl+C
		143, // Process terminated by SIGTERM
		255, // Exit status out of range
	}

	for _, code := range errorCodes {
		if exitCode == code {
			return true
		}
	}

	// 1-125 arası kodlar genellikle program spesifik hatalar
	// 0 = başarılı, 1-125 = program hatası, 126+ = sistem hatası
	return exitCode > 0 && exitCode != 2 // 2 genellikle "usage error" ama kritik değil
}

func (m *Manager) handleAutoRestart(name string, p *Process, exitCode int) {
	autoRestartConfig := p.config.AutoRestart

	switch autoRestartConfig {
	case "always":
		fmt.Printf("%s yeniden başlatılıyor (always politikası)\n", name)
		fmt.Print("taskmaster> ")
		logger.LogInfo(p.config.Stdout, fmt.Sprintf("%s yeniden başlatılıyor (always politikası)", name))
		m.StartProgram(name)
	case "never":
		fmt.Printf("%s bitti, yeniden başlatılmayacak (never politikası)\n", name)
		fmt.Print("taskmaster> ")

		// Çıkış koduna göre log seviyesi belirle
		if isErrorExitCode(exitCode) {
			logger.LogError(p.config.Stderr, fmt.Sprintf("%s hata ile bitti, yeniden başlatılmayacak (never politikası), çıkış kodu: %d", name, exitCode))
		} else {
			logger.LogInfo(p.config.Stdout, fmt.Sprintf("%s normal şekilde bitti, yeniden başlatılmayacak (never politikası)", name))
		}
	case "unexpected":
		isExpected := false
		for _, code := range p.config.ExitCodes {
			if code == exitCode {
				isExpected = true
				break
			}
		}

		if !isExpected {
			fmt.Printf("%s beklenmeyen çıkış kodu ile sonlandı (%d), yeniden başlatılıyor\n", name, exitCode)
			fmt.Print("taskmaster> ")
			logger.LogWarning(p.config.Stderr, fmt.Sprintf("%s beklenmeyen çıkış kodu ile sonlandı (%d), yeniden başlatılıyor", name, exitCode))
			m.StartProgram(name)
		} else {
			fmt.Printf("%s beklenen çıkış kodu ile sonlandı (%d), yeniden başlatılmayacak\n", name, exitCode)
			fmt.Print("taskmaster> ")
			logger.LogInfo(p.config.Stdout, fmt.Sprintf("%s beklenen çıkış kodu ile sonlandı (%d), yeniden başlatılmayacak", name, exitCode))
		}
	}
	m.removeProcessMap(name, p)
}

func (m *Manager) StopProgram(name string) {
	procs, exists := m.processes[name]
	if !exists {
		fmt.Printf("'%s' adında bir program yok\n", name)
		fmt.Print("taskmaster> ")
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

// program kapnma işlemi
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
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	procs, exists := m.processes[name]

	if !exists || len(procs) == 0 {
		fmt.Printf("%s: Çalışan süreç yok\n", name)
		fmt.Print("taskmaster> ")
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
	fmt.Print("taskmaster> ")
}

func (m *Manager) Status() {
	fmt.Println("Program         Status             PID       ID")
	fmt.Println("--------------------------------------------")

	for name, procs := range m.processes {
		if len(procs) == 0 {
			fmt.Printf("%-15s No processes\n", name)
			fmt.Print("taskmaster> ")
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
		fmt.Print("taskmaster> ")
	}
}
