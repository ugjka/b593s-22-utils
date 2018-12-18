// watch clone
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
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
	mu := &sync.RWMutex{}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	resize := make(chan os.Signal, 5)
	signal.Notify(resize, syscall.SIGWINCH)

	h, w, err := geometry()
	if err == nil {
		go resizeHandler(&h, &w, resize, mu)
	} else {
		h = 20
		w = 80
	}
	go interruptHandler(stop, mu)
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
		mu.RLock()
		for i := 0; i < h && i < len(arr)-1; i++ {
			text := arr[i]
			if len(text) > w {
				text = text[:w]
			}
			if i+1 != h {
				fmt.Println(text)
				continue
			}
			fmt.Print(text)
		}
		mu.RUnlock()
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

func resizeHandler(h, w *int, resize <-chan os.Signal, mu *sync.RWMutex) {
	for {
		<-resize
		mu.Lock()
		*h, *w, _ = geometry()
		mu.Unlock()
	}
}

func interruptHandler(stop <-chan os.Signal, mu *sync.RWMutex) {
	<-stop
	mu.Lock()
	tput("rmcup")
	os.Exit(0)
}

func tput(args ...string) {
	cmd := exec.Command("tput", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
