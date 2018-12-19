// watch clone
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
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
		out, err := cmd.StdoutPipe()
		errExit(err)
		err = cmd.Start()
		errExit(err)
		b := make([]byte, 1)
		buff := bytes.NewBufferString("")
		tput("clear")
		mu.Lock()
		for i, j := 0, 0; ; i++ {
			_, err := out.Read(b)
			if err == io.EOF {
				break
			} else {
				errExit(err)
			}
			if string(b) == "\n" {
				j++
				i = 0
			}
			if i >= w {
				j++
				i = 0
			}
			if j >= h {
				err := cmd.Process.Kill()
				errExit(err)
				break
			}
			_, err = buff.Write(b)
			errExit(err)
		}
		mu.Unlock()
		fmt.Print(buff.String())
		buff.Reset()
		_, err = cmd.Process.Wait()
		errExit(err)
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

func errExit(err error) {
	if err == nil {
		return
	}
	tput("rmcup")
	fmt.Fprintf(os.Stderr, "Error :%v", err)
	os.Exit(1)
}
