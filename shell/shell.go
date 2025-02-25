package shell

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"taskmaster/process"
)

func Run(manager *process.Manager) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("taskmaster> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		// Komutları parçalara ayır
		parts := strings.Fields(input)
		command := ""
		if len(parts) > 0 {
			command = parts[0]
		}
		switch command {
		case "status":
			status := manager.GetStatus()
			if len(status) == 0 {
				fmt.Println("Hiçbir süreç tanımlı değil.")
			} else {
				fmt.Println("Program\t\tStatus")
				fmt.Println("---------------------")
				for name, states := range status {
					for i, state := range states {
						fmt.Printf("%s[%d]\t\t%s\n", name, i, state)
					}
				}
			}
		case "start":
			if len(parts) < 2 {
				fmt.Println("Kullanım: start <program_adı>")
			} else {
				programName := parts[1]
				err := manager.StartProgram(programName)
				if err != nil {
					fmt.Printf("Hata: %v\n", err)
				} else {
					fmt.Printf("'%s' başlatıldı.\n", programName)
				}
			}
		case "stop":
			fmt.Println("TODO: Süreci durdur")
		case "restart":
			fmt.Println("TODO: Süreci yeniden başlat")
		case "reload":
			fmt.Println("TODO: Yapılandırmayı yenile")
		case "exit":
			return
		default:
			fmt.Println("Bilinmeyen komut:", input)
		}
	}
}