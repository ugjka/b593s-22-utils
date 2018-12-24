// which clone
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	all := flag.Bool("a", false, "print all matches in PATH")
	sym := flag.Bool("s", false, "resolve symlink")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		errExit(fmt.Errorf("no argument given"))
	}
	path := os.Getenv("PATH")
	if path == "" {
		errExit(fmt.Errorf("no path variable found"))
	}
	dirs := strings.Split(path, ":")
	for _, v := range dirs {
		v = v + "/" + args[0]
		if isExist(v) && isExec(v) {
			if *sym {
				v = resolve(v)
			}
			fmt.Println(v)
			if !*all {
				break
			}
		}
	}
}

func errExit(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func isExist(path string) bool {
	_, err := os.Lstat(path)
	if err == nil {
		return true
	}
	return false
}

func isExec(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}

func resolve(path string) string {
	to, err := os.Readlink(path)
	if err != nil {
		return path
	}
	return fmt.Sprintf("%s -> %s", path, to)
}
