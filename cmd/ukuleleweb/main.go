package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gnoack/ukuleleweb"
	"github.com/peterbourgon/diskv/v3"
)

var (
	listenNet  = flag.String("net", "tcp", "HTTP listen network (i.e. 'tcp', 'unix')")
	listenAddr = flag.String("addr", "localhost:8080", "HTTP listen address")
	storeDir   = flag.String("store_dir", "", "Store directory")
	mainPage   = flag.String("main_page", "MainPage", "The default page to use as the main page")
)

func main() {
	flag.Parse()

	if *storeDir == "" {
		fmt.Fprintln(flag.CommandLine.Output(), "Needs --store_dir")
		flag.Usage()
		return
	}

	d := diskv.New(diskv.Options{
		BasePath:     *storeDir,
		CacheSizeMax: 1024 * 1024, // 1MB
	})
	ukuleleweb.AddRoutes(http.DefaultServeMux, *mainPage, d)

	s := http.Server{}
	l, err := net.Listen(*listenNet, *listenAddr)
	if err != nil {
		log.Fatalf("Could not listen on net %q address %q: %v", *listenNet, *listenAddr, err)
	}

	restrictAccess(*storeDir)

	fmt.Printf("Listening on %s!%s\n", *listenNet, *listenAddr)
	err = s.Serve(l)
	if err != nil {
		log.Printf("http.ListenAndServe: %v", err)
	}
}
