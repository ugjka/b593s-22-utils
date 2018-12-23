// Show the signal level of your b593s-22 in your system status tray
package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/getlantern/systray"
)

func main() {
	systray.Run(onready, nil)
}

func onready() {
	go quitter()
	//Get level images
	sigs, err := getSigImgs()
	if err != nil {
		log.Fatal(err)
	}
	//Get no signal image
	sigh, err := getSigh()
	if err != nil {
		log.Fatal(err)
	}
	systray.SetIcon(sigh)
	for {
		time.Sleep(time.Second)
		level, err := getSignal()
		if err != nil {
			systray.SetIcon(sigh)
			continue
		}
		systray.SetIcon(sigs[level/10])
	}
}

func getSignal() (sig int, err error) {
	const post = "http://192.168.1.1/index/getStatusByAjax.cgi?rid=%d"
	rand.Seed(time.Now().UnixNano())
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
	return s.SIG, nil
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

func getSigImgs() (sigs map[int][]byte, err error) {
	const sigTmpl = "http://192.168.1.1/res/signal_%d.gif"
	sigs = make(map[int][]byte)
	for i := 0; i <= 5; i++ {
		resp, err := http.DefaultClient.Get(fmt.Sprintf(sigTmpl, i))
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()
		sigs[i] = b
	}
	return sigs, nil
}

func getSigh() (sigh []byte, err error) {
	const sighURL = "http://192.168.1.1/res/sigh.png"
	resp, err := http.DefaultClient.Get(sighURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	sigh, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return sigh, nil
}
