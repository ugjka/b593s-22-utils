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
	"syscall"
	"time"
)

func main() {
	d := flag.Int64("d", 500, "delay in milliseconds")
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

	h, w := geometry()
	err := tput("smcup")
	if err != nil {
		print("\033[?1049h")
	}
	cmd := &exec.Cmd{}
	b := make([]byte, h*w)

	for {
		select {
		case <-resize:
			h, w = geometry()
			b = make([]byte, h*w)
		case <-stop:
			exit()
		default:
		}
		cmd = exec.Command(a[0])
		if len(a) > 1 {
			cmd.Args = a
		}

		out, err := cmd.StdoutPipe()
		errExit(err)
		err = cmd.Start()
		errExit(err)
		n, err := out.Read(b)
		errExit(err)
		err = kill(cmd)
		errExit(err)

		i, j, t := 0, 0, 0
		for _, v := range b[:n] {
			if v == '\n' || i >= w {
				j++
				i = 0
			}
			if j >= h {
				break
			}
			i++
			t++
		}

		err = tput("clear")
		if err != nil {
			print("\033[2J\033[H")
		}
		fmt.Printf("%s", b[:t])
		time.Sleep(time.Millisecond * time.Duration(*d))
	}
}

func geometry() (h, w int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	b, err := cmd.Output()
	if err != nil {
		return 20, 80
	}
	_, err = fmt.Sscanf(string(b), "%d %d", &h, &w)
	if err != nil {
		return 20, 80
	}
	return
}

func exit() {
	err := tput("rmcup")
	if err != nil {
		print("\033[?1049l")
	}
	os.Exit(0)
}

func tput(args ...string) error {
	cmd := exec.Command("tput", args...)
	b, err := cmd.Output()
	if err == nil {
		io.Copy(os.Stdout, bytes.NewReader(b))
	}
	return err
}

func errExit(err error) {
	if err == nil {
		return
	}
	terr := tput("rmcup")
	if terr != nil {
		print("\033[?1049l")
	}
	fmt.Fprintf(os.Stderr, "Error :%v", err)
	os.Exit(1)
}

func kill(c *exec.Cmd) (err error) {
	if c.Process == nil {
		return
	}
	err = c.Process.Kill()
	if err != nil {
		return
	}
	_, err = c.Process.Wait()
	if err != nil {
		return
	}
	return
}
