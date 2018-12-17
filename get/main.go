package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/jfbus/httprs"

	"github.com/cheggaaa/pb"
)

func main() {
	cfg := &tls.Config{InsecureSkipVerify: true}
	http.DefaultClient.Transport = &http.Transport{TLSClientConfig: cfg}

	if len(os.Args) < 2 {
		os.Stderr.WriteString("Nothing given!!!\n")
		return
	}
	_url, err := url.Parse(os.Args[1])
	if err != nil {
		os.Stderr.WriteString("Invalid URL!!!\n")
		return
	}

	resp, err := http.Get(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get: %v\n", err)
		return
	}
	defer resp.Body.Close()

	_len := resp.ContentLength
	rs := httprs.NewHttpReadSeeker(resp, http.DefaultClient)
	defer rs.Close()

	filename := getFileName(_url, resp.Header)
	info, exist := os.Stat(filename)

	out, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open or create file: %v\n", err)
		return
	}
	defer out.Close()

	if exist == nil {
		currLen := info.Size()
		if _len == currLen {
			fmt.Fprintf(os.Stderr, "File already downloaded!!!\n")
			return
		}
		err := seek(currLen, out, rs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not seek: %v\n", err)
			return
		}
		_len = _len - currLen
	}

	err = copyWithPB(out, rs, _len)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not copy contents: %v\n", err)
		return
	}

	fmt.Println("Success!!!")
}

func getFileName(_url *url.URL, h http.Header) (filename string) {
	f := h.Get("Content-Disposition")
	var filenameReg = regexp.MustCompile("filename=\"(.+)\"")
	arr := filenameReg.FindStringSubmatch(f)
	if len(arr) > 1 {
		return arr[1]
	}
	arr = strings.Split(_url.Path, "/")
	return arr[len(arr)-1]
}

func copyWithPB(dst io.Writer, src io.Reader, len int64) (err error) {
	bar := pb.New64(len).SetUnits(pb.U_BYTES)
	bar.Start()
	psrc := bar.NewProxyReader(src)
	_, err = io.Copy(dst, psrc)
	return err
}

func seek(seekTo int64, seekers ...io.Seeker) (err error) {
	for _, s := range seekers {
		_, err = s.Seek(seekTo, 0)
		if err != nil {
			return err
		}
	}
	return nil
}
