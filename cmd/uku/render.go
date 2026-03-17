package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gnoack/ukuleleweb"
)

func runRender(args []string) {
	fs := flag.NewFlagSet("uku render", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: uku render [FILENAME...]\n\n")
		fmt.Fprintf(fs.Output(), "Render markdown files to HTML.\n")
	}
	fs.Parse(args)

	for _, fn := range fs.Args() {
		md, err := os.ReadFile(fn)
		if err != nil {
			log.Fatalf("ReadFile(%q): %v\n", fn, err)
			continue
		}
		html, err := ukuleleweb.RenderHTML(string(md))
		if err != nil {
			log.Fatalf("RenderHTML(ReadFile(%q)): %v", fn, err)
			continue
		}
		fmt.Print(html)
	}
}
