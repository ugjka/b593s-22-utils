package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func main() {
	l := flag.Bool("l", false, "count lines")
	w := flag.Bool("w", false, "count words")
	c := flag.Bool("c", false, "count chars")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "error: no files given!\n")
		os.Exit(1)
	}
	if *l == false && *w == false && *c == false {
		*l = true
		*w = true
		*c = true
	}
	err := check(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	tw := tabwriter.NewWriter(os.Stdout, 4, 0, 2, ' ', tabwriter.AlignRight)
	for _, v := range args {
		lines, words, chars := 0, 0, 0
		f, err := os.Open(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			lines++
			words += len(strings.Fields(sc.Text()))
			chars += strings.Count(sc.Text(), "")
		}
		if *l {
			fmt.Fprintf(tw, "%d\t", lines)
		}
		if *w {
			fmt.Fprintf(tw, "%d\t", words)
		}
		if *c {
			fmt.Fprintf(tw, "%d\t", chars)
		}
		fmt.Fprintf(tw, "%s\t\n", v)
	}
	tw.Flush()
}

func check(paths []string) error {
	for _, v := range paths {
		info, err := os.Lstat(v)
		if err != nil {
			return fmt.Errorf("could not stat %s", v)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("not a regular file %s", v)
		}
		_, err = os.Open(v)
		if err != nil {
			return fmt.Errorf("could not open %s", v)
		}
	}
	return nil
}
