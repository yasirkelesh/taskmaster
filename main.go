package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"taskmaster/config"
	"taskmaster/process"
	"taskmaster/shell"
)

func main() {
	// Yapılandırmayı yükle
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		fmt.Printf("Yapılandırma yüklenemedi: %v\n", err)
		os.Exit(1)
	}

	// Süreç yöneticisini başlat
	manager := process.NewManager(cfg)
	manager.Start()

	// SIGHUP sinyalini dinle (yapılandırma yenileme)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGHUP)
		for range sigChan {
			fmt.Println("SIGHUP alındı, yapılandırma yenileniyor...")
			newCfg, err := config.LoadConfig("config/config.yaml")
			if err == nil {
				manager.UpdateConfig(newCfg)
			}
		}
	}()

	// Kontrol kabuğunu başlat
	shell.Run(manager)

	// Programın kapanmasını bekle
	manager.Stop()
	fmt.Println("Taskmaster kapatıldı.")
}