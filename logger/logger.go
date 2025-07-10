package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EnsureLogFileExists sadece bir kere çalıştırılır, dosya/dizin yoksa oluşturur
func EnsureLogFileExists(logPath string) error {
	if logPath == "" {
		return nil
	}

	// Dizini oluştur (eğer yoksa)
	logDir := filepath.Dir(logPath)
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return err
	}

	// Dosya yoksa boş dosya oluştur
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		file, err := os.Create(logPath)
		if err != nil {
			return err
		}
		file.Close()
	}

	return nil
}

// LogToFile sadece dosyaya yazar, oluşturma yapmaz
func LogToFile(filePath, message string) {
	if filePath == "" {
		return
	}

	// Sadece dosyayı aç (append mode)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Log dosyası açılamadı (%s): %v\n", filePath, err)
		return
	}
	defer file.Close()

	timestamp := time.Now().Format("2006/01/02 15:04:05")
	logMessage := fmt.Sprintf("%s [TASKMASTER] %s\n", timestamp, message)

	_, err = file.WriteString(logMessage)
	if err != nil {
		fmt.Printf("Log yazma hatası (%s): %v\n", filePath, err)
	}
}

// LogInfo logs informational messages to stdout file
func LogInfo(stdoutPath, message string) {
	LogToFile(stdoutPath, "[INFO] "+message)
}

// LogError logs error messages to stderr file
func LogError(stderrPath, message string) {
	LogToFile(stderrPath, "[ERROR] "+message)
}

// LogWarning logs warning messages to stderr file
func LogWarning(stderrPath, message string) {
	LogToFile(stderrPath, "[WARNING] "+message)
}

// CreateLogFile sadece process başlangıcında çağrılır
func CreateLogFile(logPath string) (*os.File, error) {
	if logPath == "" {
		return nil, nil
	}

	// Önce dosyanın var olduğundan emin ol
	err := EnsureLogFileExists(logPath)
	if err != nil {
		return nil, err
	}

	// Dosyayı aç
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// CreateStdoutFile creates stdout log file
func CreateStdoutFile(stdoutPath string) (*os.File, error) {
	return CreateLogFile(stdoutPath)
}

// CreateStderrFile creates stderr log file
func CreateStderrFile(stderrPath string) (*os.File, error) {
	return CreateLogFile(stderrPath)
}
