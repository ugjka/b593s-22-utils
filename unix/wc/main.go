// wc implementation
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
	if *l == false && *w == false && *c == false {
		*l, *w, *c = true, true, true
	}
	err := check(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight)
	if *l {
		fmt.Fprint(tw, "lines\t")
	}
	if *w {
		fmt.Fprint(tw, "words\t")
	}
	if *c {
		fmt.Fprint(tw, "characters\t")
	}
	fmt.Fprint(tw, " \n")

	if len(args) == 0 {
		args = append(args, "-")
	}
	tlines, twords, tchars := 0, 0, 0
	for _, v := range args {
		sc := &bufio.Scanner{}
		if v != "-" {
			f, err := os.Open(v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}
			defer f.Close()
			sc = bufio.NewScanner(f)
		} else {
			sc = bufio.NewScanner(os.Stdin)
		}
		lines, words, chars := 0, 0, 0
		for sc.Scan() {
			lines++
			words += len(strings.Fields(sc.Text()))
			chars += strings.Count(sc.Text(), "")
		}
		tlines += lines
		twords += words
		tchars += chars

		if *l {
			fmt.Fprintf(tw, "%d\t", lines)
		}
		if *w {
			fmt.Fprintf(tw, "%d\t", words)
		}
		if *c {
			fmt.Fprintf(tw, "%d\t", chars)
		}
		fmt.Fprintf(tw, " %s\n", v)
	}
	if len(args) > 1 {
		if *l {
			fmt.Fprintf(tw, "%d\t", tlines)
		}
		if *w {
			fmt.Fprintf(tw, "%d\t", twords)
		}
		if *c {
			fmt.Fprintf(tw, "%d\t", tchars)
		}
		fmt.Fprint(tw, " total\n")
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
