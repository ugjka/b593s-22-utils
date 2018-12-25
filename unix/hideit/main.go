// Simple daemonizer
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Nothing was given...")
		return
	}
	cmd := &exec.Cmd{}
	if len(args) == 2 {
		cmd = exec.Command(args[1])
	} else {
		cmd = exec.Command(args[1], args[2:]...)
	}
	cmd.Start()
	fmt.Println(cmd.Process.Pid)
}
