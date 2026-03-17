// An experimental command line tool for visualizing the graph of wiki pages.
// The output of this command is a digraph for input with GraphViz tools.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gnoack/ukuleleweb"
	"github.com/peterbourgon/diskv/v3"
)

type PageInfo struct {
	OutgoingLinks []string `json:"outgoingLinks"`
	Size          int      `json:"size"`
}

var formatters = map[string]func(io.Writer, map[string]PageInfo){
	"dot":  writeDigraphDot,
	"json": writeDigraphJson,
}

func writeDigraphDot(w io.Writer, links map[string]PageInfo) {
	fmt.Fprintln(w, "digraph G {")
	fmt.Fprintln(w, "\toverlap = false;")
	fmt.Fprintln(w, "\tnode [color=red];")

	fmt.Fprintln(w)
	for pn := range links {
		fmt.Fprintf(w, "\t%v [color=black shape=box];\n", pn)
	}

	fmt.Fprintln(w)
	for pn, info := range links {
		for _, ogPn := range info.OutgoingLinks {
			fmt.Fprintf(w, "\t%v -> %v;\n", pn, ogPn)
		}
	}
	fmt.Fprintln(w, "}")
}

func writeDigraphJson(w io.Writer, links map[string]PageInfo) {
	buf, err := json.Marshal(links)
	if err != nil {
		log.Fatal("Failed to marshal JSON")
	}
	_, err = w.Write(buf)
	if err != nil {
		log.Fatal("Failed to write JSON")
	}
}

func runViz(args []string) {
	fs := flag.NewFlagSet("uku viz", flag.ExitOnError)
	fs.Usage = func() {
		o := fs.Output()
		fmt.Fprintf(o, "Usage: uku viz -store_dir=/path/to/wiki | neato -Tsvg > out.svg\n\n")
		fmt.Fprintln(o, "Flags:")
		fs.PrintDefaults()
	}

	storeDir := fs.String("store_dir", "", "Store directory")
	outFormat := fs.String("out.format", "dot", `output format ("dot" or "json")`)

	fs.Parse(args)

	if *storeDir == "" {
		fmt.Fprintln(fs.Output(), "Needs --store_dir")
		fs.Usage()
		os.Exit(1)
	}

	write, ok := formatters[*outFormat]
	if !ok {
		var keys []string
		for k := range formatters {
			keys = append(keys, k)
		}
		fmt.Fprintf(fs.Output(), "Wrong --out.format, need one of %q\n", keys)
		fs.Usage()
		os.Exit(1)
	}

	d := diskv.New(diskv.Options{
		BasePath:     *storeDir,
		CacheSizeMax: 1024 * 1024, // 1MB
	})

	links := make(map[string]PageInfo)
	for pn := range d.Keys(nil) {
		md := d.ReadString(pn)
		info := PageInfo{Size: len(md), OutgoingLinks: []string{}}
		for ogPn := range ukuleleweb.OutgoingLinks(md) {
			info.OutgoingLinks = append(info.OutgoingLinks, ogPn)
		}
		links[pn] = info
	}
	write(os.Stdout, links)
}
