//go:generate go run gen/main.go
// Show the signal level of your b593s-22 in your system status tray
package main

import (
	"encoding/xml"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/getlantern/systray"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	systray.Run(onready, nil)
}

func onready() {
	go quitter()
	//Get no signal image
	systray.SetIcon(nosig)
	prev := math.MaxInt64
	for {
		time.Sleep(time.Second)
		level, err := getSignal()
		if err != nil {
			systray.SetIcon(nosig)
			continue
		}
		if level == prev {
			continue
		}
		systray.SetIcon(sigs[level])
		prev = level
	}
}

func getSignal() (sig int, err error) {
	const post = "http://192.168.1.1/index/getStatusByAjax.cgi?rid=%d"
	n := rand.Intn(1000)
	resp, err := http.DefaultClient.PostForm(fmt.Sprintf(post, n), nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var s Status
	err = xml.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		return 0, err
	}
	return s.SIG / 10, nil
}

//Status represents signal status
type Status struct {
	SIG int
}

func quitter() {
	m := systray.AddMenuItem("Quit", "Quit the whole app")
	<-m.ClickedCh
	systray.Quit()
}
