package mcpserver

import (
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

type Job struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// API anahtarı doğrulama middleware'i
func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-API-Key") != apiKey {
			c.JSON(401, gin.H{"status": "error", "message": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// taskmasterctl status komutunu çalıştır ve parse et
func GetJobStatus(mcpinput chan string) ([]Job, error) {
	// Gerçek uygulamada taskmasterd ile iletişime geç
	// Şimdilik taskmasterctl status simüle ediliyor
	cmd := exec.Command("taskmasterctl", "status")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	mcpinput <- string("status")

	// Örnek çıktı: "ping: running\nwebserver: stopped"
	lines := strings.Split(string(output), "\n")
	jobs := []Job{}
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			jobs = append(jobs, Job{ID: parts[0], Status: parts[1]})
		}
	}
	return jobs, nil
}
