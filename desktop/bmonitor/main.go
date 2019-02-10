package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	loginPost    = "http://%s/index/login.cgi"
	indexPage    = "http://%s/"
	maintainPage = "http://%s/html/management/maintenance.asp"
	statsPage    = "http://%s/html/status/waninforefresh.asp"
	rebootPost   = "http://%s/html/management/reboot.cgi?RequestFile=/html/management/maintenance.asp"
	rsaModulus   = "BEB90F8AF5D8A7C7DA8CA74AC43E1EE8A48E6860C0D46A5D690BEA082E3A74E1571F2C58E94EE339862A49A811A31BB4A48F41B3BCDFD054C3443BB610B5418B3CBAFAE7936E1BE2AFD2E0DF865A6E59C2B8DF1E8D5702567D0A9650CB07A43DE39020969DF0997FCA587D9A8AE4627CF18477EC06765DF3AA8FB459DD4C9AF3"
	rsaExponent  = "10001"
)

var csrfParamReg = regexp.MustCompile("var csrf_param = \"(\\w+)\";")
var csrfTokenReg = regexp.MustCompile("var csrf_token = \"(\\w+)\";")
var statsReg = regexp.MustCompile("WanStatistics = {'uprate' : '(\\d+)' , 'downrate' : '(\\d+)' , 'upvolume' : '\\d+' , 'downvolume' : '\\d+' , 'liveTime' : '\\d+'};\nwanIPDNS = {'dataip': '(.+)' , 'datadns': '.+', .+};")

var client = &http.Client{}

//needs to be sent in every request
//changes on every request
type tokens struct {
	csrfParam string
	csrfToken string
}

func main() {
	host := flag.String("host", "192.168.1.1", "B593s-22's ip adress")
	password := flag.String("pass", "", "web gui admin password")
	flag.Parse()
	if *password == "" {
		flag.Usage()
		return
	}
	username := "admin"
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create cookie jar. Error: %v\n", err)
		return
	}
	client.Jar = jar
	root, err := url.Parse(fmt.Sprintf(indexPage, *host))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse B593s-22's ip adress. Error: %v\n", err)
		flag.Usage()
		return
	}
	jar.SetCookies(root, []*http.Cookie{&http.Cookie{Name: "Language", Value: "en_us", Path: "/", Expires: time.Now().AddDate(1, 1, 1)}})
	var t tokens
	err = getTokens(fmt.Sprintf(indexPage, *host), &t, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get tokens from index page. Error: %v\n", err)
		return
	}
	pub, err := rsaPublicKey(rsaModulus, rsaExponent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create rsa publick key. Error: %v\n", err)
		return
	}
	passSHA := getSHA(username, *password, &t)
	passEnc, err := rsaEncrypt(passSHA, pub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not rsa encrypt password sha. Error: %v\n", err)
		return
	}
	err = login(username, passEnc, fmt.Sprintf(loginPost, *host), &t, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not login to B593s-22 as admin. Error: %v\n", err)
		return
	}
	stats, err := getStats(fmt.Sprintf(statsPage, *host))
	if err != nil {
		return
	}
	//fmt.Printf("downrate: %.2f mbps\nuprate: %.2f mbps\nip: %s\n", stats.downrate/1024/1024, stats.uprate/1024/1024, stats.dataip)
	// Initialize GTK without parsing any command line arguments.
	gtk.Init(nil)

	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.SetIconFromFile("/usr/share/icons/breeze/categories/32/applications-internet.svg")
	// Create a new label widget to show in the window.
	statl, err := gtk.LabelNew(fmt.Sprintf("D: %.2f Mbps\nU: %.2f Mbps\nIP: %s\n", stats.downrate/1024/1024, stats.uprate/1024/1024, stats.dataip))
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}

	// Add the label to the window.
	win.SetTitle(fmt.Sprintf("%.2f mbps", stats.downrate/1024/1024))
	win.Add(statl)

	// Set the default window size.
	win.SetDefaultSize(150, 100)

	go func() {
		for {
			stats, err := getStats(fmt.Sprintf(statsPage, *host))
			if err != nil {
				log.Fatal("getStats() failed:", err)
			}
			s := fmt.Sprintf("%.2f/%.2f", stats.downrate/1024/1024, stats.uprate/1024/1024)
			_, err = glib.IdleAdd(win.SetTitle, s)
			if err != nil {
				log.Fatal("IdleAdd() failed:", err)
			}
			s = fmt.Sprintf("D: %.2f Mbps\nU: %.2f Mbps\nIP: %s\n", stats.downrate/1024/1024, stats.uprate/1024/1024, stats.dataip)
			_, err = glib.IdleAdd(statl.SetLabel, s)
			if err != nil {
				log.Fatal("IdleAdd() failed:", err)
			}
			time.Sleep(time.Second)
		}
	}()

	// Recursively show all widgets contained in this window.
	win.ShowAll()

	// Begin executing the GTK main loop.  This blocks until
	// gtk.MainQuit() is run.
	gtk.Main()
}

func getSHA(username, password string, t *tokens) string {
	passB64 := base64.StdEncoding.EncodeToString([]byte(password))
	combined := username + passB64 + t.csrfParam + t.csrfToken
	return fmt.Sprintf("%x", sha256.Sum256([]byte(combined)))
}

func findSubmatch(r *regexp.Regexp, in string) (string, error) {
	arr := r.FindStringSubmatch(in)
	if len(arr) < 2 {
		return "", fmt.Errorf("no regexp matches")
	}
	return arr[1], nil
}

func rsaModulusToBigInt(m string) (b big.Int, err error) {
	_, err = fmt.Sscanf(m, "%X", &b)
	if err != nil {
		return
	}
	return
}

func rsaExponentToInt(e string) (i int, err error) {
	_, err = fmt.Sscanf(e, "%X", &i)
	if err != nil {
		return
	}
	return
}

func rsaPublicKey(mod, exp string) (*rsa.PublicKey, error) {
	m, err := rsaModulusToBigInt(mod)
	if err != nil {
		return nil, err
	}
	e, err := rsaExponentToInt(exp)
	if err != nil {
		return nil, err
	}
	pub := new(rsa.PublicKey)
	pub.N = &m
	pub.E = e
	return pub, nil
}

func rsaEncrypt(in string, pub *rsa.PublicKey) (string, error) {
	b, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(in))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func getTokens(link string, t *tokens, debug bool) error {
	resp, err := client.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = findTokens(b, t)
	if err != nil {
		return err
	}
	if debug {
		fmt.Println(string(b))
	}
	return nil
}

func findTokens(b []byte, t *tokens) error {
	s, err := findSubmatch(csrfParamReg, string(b))
	if err != nil {
		return err
	}
	t.csrfParam = s
	s, err = findSubmatch(csrfTokenReg, string(b))
	if err != nil {
		return err
	}
	t.csrfToken = s
	return nil
}

func login(username, rsaPassword, link string, t *tokens, debug bool) error {
	v := url.Values{}
	v.Add("Username", username)
	v.Add("Password", rsaPassword)
	v.Add("csrf_param", t.csrfParam)
	v.Add("csrf_token", t.csrfToken)
	resp, err := client.PostForm(link, v)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if debug {
		fmt.Println(string(b))
	}
	return nil
}

type stats struct {
	uprate   float64
	downrate float64
	dataip   string
}

func getStats(link string) (s *stats, err error) {
	resp, err := client.Get(link)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	matches := statsReg.FindStringSubmatch(string(data))
	if len(matches) < 4 {
		return nil, fmt.Errorf("no matches for stats regex")
	}
	s = &stats{
		dataip: matches[3],
	}
	fmt.Sscan(matches[1], &s.uprate)
	fmt.Sscan(matches[2], &s.downrate)
	return s, nil
}
