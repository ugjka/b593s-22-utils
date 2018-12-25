// Simple free implementation
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

// TODO: https://golang.org/pkg/syscall/#Sysinfo_t

func main() {
	b, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	arr := bytes.Split(b, []byte("\n"))
	fmt.Printf("%s\n", bytes.Join(arr[:3], []byte("\n")))
}
