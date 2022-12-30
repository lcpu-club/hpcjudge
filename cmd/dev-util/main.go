package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/satori/uuid"
)

func main() {

	switch os.Args[1] {
	case "gen-uuid-v4":
		fmt.Println(uuid.NewV4())
	case "multi-run":
		wg := &sync.WaitGroup{}
		for _, arg := range os.Args[2:] {
			wg.Add(1)
			go func(arg string) {
				cmd := exec.Command("sh", "-c", arg)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				fmt.Println("\r\n******** Running", arg, "********")
				err := cmd.Run()
				wg.Done()
				if err != nil {
					fmt.Println(err)
				}
			}(arg)
		}
		wg.Wait()
	}
}
