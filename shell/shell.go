package shell

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"taskmaster/process"
)

func Run(manager *process.Manager, sigChan chan os.Signal) {
	reader := bufio.NewReader(os.Stdin)
	var input string
	for {
		
		fmt.Print("taskmaster> ")
	
		
		input, _ = reader.ReadString('\n')
		
		
		input = strings.TrimSpace(input)
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
			if len(parts) < 2 {
				fmt.Println("Kullanım: stop <program_adı>")
			} else {
				programName := parts[1]
				manager.StopProgram(programName)
				fmt.Printf("'%s' durduruldu.\n", programName)
			}
		case "restart":
			if len(parts) < 2 {
				fmt.Println("Kullanım: restart <program_adı>")
			} else {
				programName := parts[1]
				manager.RestartProgram(programName)
				fmt.Printf("'%s' yeniden başlatıldı.\n", programName)
			}
		case "reload":
			sigChan <- syscall.SIGHUP
			fmt.Println("sigChan tetiklendi.")
		case "exit":
			return
		default:
			fmt.Println("Bilinmeyen komut:", input)
		}
	}
}
