package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

const path = "/var/locks/launcher"

func main() {
	a := os.Args
	switch {
	case len(a) < 2:
		return
	case a[1] == "-D":
		exec.Command(a[0], "start").Start()
		return
	case a[1] != "start":
		return
	}

	// Check the lock
	_, err := os.Open(path)
	if err == nil {
		return
	}

	os.Setenv("PATH", "/sbin:/bin:/xbin:/usr/bin:/vendor/bin:/system/sbin:/system/bin:/system/xbin:/app/bin:/mnt/usb1_1/bin")

	// Create a lock
	ioutil.WriteFile(path, []byte("started"), 0644)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Programs that exit
	oneshots := make([]*exec.Cmd, 1)
	oneshots[0] = exec.Command("settings.sh")

	// Programs that do not exit
	daemons := make([]*exec.Cmd, 1)
	daemons[0] = exec.Command("redditbot")

	for _, v := range oneshots {
		v.Run()
	}

	for _, v := range daemons {
		v.Start()
	}

	<-stop
	for _, v := range daemons {
		if v.Process != nil {
			v.Process.Kill()
		}
	}
	os.Remove(path)
}
