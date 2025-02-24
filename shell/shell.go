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

		switch input {
		case "status":
			fmt.Println("TODO: Durumları göster")
		case "start":
			fmt.Println("TODO: Süreci başlat")
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