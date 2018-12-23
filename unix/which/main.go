// which clone
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	a := os.Args
	if len(a) < 2 {
		fmt.Fprintf(os.Stderr, "No name given!\n")
		return
	}
	path, err := exec.LookPath(a[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s does not exist!\n", a[1])
		return
	}
	fmt.Println(path)
}
