// Barebones uptime util
package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/hako/durafmt"
)

func main() {
	sys := &syscall.Sysinfo_t{}
	err := syscall.Sysinfo(sys)
	errExit(err)
	dur := time.Duration(sys.Uptime) * time.Second
	humandur := durafmt.Parse(dur)
	fmt.Printf("up %s\n", humandur)
}

func errExit(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
