package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"taskmaster/config"
	"taskmaster/mcpserver"
	"taskmaster/process"
	"taskmaster/shell"

	"github.com/gin-gonic/gin"
)

func main() {
	// Yapılandırmayı yükle

	r := gin.Default()
	apiKey := "taskmaster-secret-key"
	r.Use(mcpserver.AuthMiddleware(apiKey))

	// GET /jobs uç noktası

	mcpinputchan := make(chan string)
	r.GET("/jobs", func(c *gin.Context) {
		jobs, err := mcpserver.GetJobStatus(mcpinputchan)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "success", "data": jobs})
	})

	// Sunucuyu başlat
	go r.Run(":8080")

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		fmt.Printf("Configuration failed to load: %v\n", err)
		os.Exit(1)
	}

	// Süreç yöneticisini başlat
	manager := process.NewManager(cfg)
	manager.Start()

	// SIGHUP sinyalini dinle (yapılandırma yenileme)
	sigChan := make(chan os.Signal, 1)
	go func() {
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
	shell.Run(manager, sigChan)

	// Programın kapanmasını bekle
	manager.Stop()
	fmt.Println("Taskmaster kapatıldı.")
}
