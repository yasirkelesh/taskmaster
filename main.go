package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"taskmaster/config"
	"taskmaster/process"
	"taskmaster/shell"
)

func main() {
	// 1. Yapılandırma dosyasını yükle
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Yapılandırma dosyası yüklenemedi: %v", err)
	}

	// 2. Süreç yöneticisini başlat
	pm := process.NewProcessManager(cfg)

	// 3. Kontrol kabuğunu başlat
	go func() {
		shell.Start(pm)
	}()

	// 4. Sinyalleri yakala (ör. SIGHUP, SIGINT)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for sig := range sigs {
		switch sig {
		case syscall.SIGHUP:
			log.Println("Yapılandırma dosyası yeniden yükleniyor...")
			cfg, err := config.Load("config.yaml")
			if err == nil {
				pm.Reload(cfg)
			}
		case syscall.SIGINT, syscall.SIGTERM:
			log.Println("Program durduruluyor...")
			pm.StopAll()
			os.Exit(0)
		}
	}
}
