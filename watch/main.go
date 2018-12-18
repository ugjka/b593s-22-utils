// watch clone
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	d := flag.Int64("d", 1000, "delay in milliseconds")
	flag.Parse()
	a := flag.Args()
	if len(a) < 1 {
		os.Stderr.WriteString("nothing given!!!\n")
		return
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	resize := make(chan os.Signal, 5)
	signal.Notify(resize, syscall.SIGWINCH)

	h, w, err := geometry()
	if err == nil {
		go resizeHandler(&h, &w, resize)
	} else {
		h = 20
		w = 80
	}
	go interruptHandler(stop)
	tput("smcup")
	cmd := &exec.Cmd{}
	for {
		cmd = exec.Command(a[0])
		if len(a) > 1 {
			cmd.Args = a
		}
		b, _ := cmd.CombinedOutput()
		arr := strings.Split(string(b), "\n")
		//Clear screen and move cursor home
		tput("clear")
		for i := 0; i < h && i < len(arr)-1; i++ {
			text := arr[i]
			if len(text) > w {
				text = text[:w]
			}
			if i != h-1 {
				text += "\n"
			}
			fmt.Print(text)
		}
		time.Sleep(time.Millisecond * time.Duration(*d))
	}
}

func geometry() (h, w int, err error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	b, err := cmd.Output()
	if err != nil {
		return
	}
	_, err = fmt.Sscanf(string(b), "%d %d", &h, &w)
	return
}

func resizeHandler(h, w *int, resize <-chan os.Signal) {
	for {
		<-resize
		*h, *w, _ = geometry()
	}
}

func interruptHandler(stop <-chan os.Signal) {
	<-stop
	tput("rmcup")
	os.Exit(0)
}

func tput(args ...string) {
	cmd := exec.Command("tput", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
