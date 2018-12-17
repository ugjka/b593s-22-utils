package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	b, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	arr := bytes.Split(b, []byte("\n"))
	fmt.Printf("%s\n", bytes.Join(arr[:3], []byte("\n")))
}
