package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"

	"github.com/gnoack/ukuleleweb"
	"github.com/peterbourgon/diskv/v3"
)

var (
	listenNet    = flag.String("net", "tcp", "HTTP listen network (i.e. 'tcp', 'unix')")
	listenAddr   = flag.String("addr", "localhost:8080", "HTTP listen address")
	storeDir     = flag.String("store_dir", "", "Store directory")
	mainPage     = flag.String("main_page", "MainPage", "The default page to use as the main page")
	templatePage = flag.String("template.page", "", "A glob for template files to override the page template")
	templateEdit = flag.String("template.edit", "", "A glob for template files to override the edit template")
)

func main() {
	flag.Parse()

	if *storeDir == "" {
		fmt.Fprintln(flag.CommandLine.Output(), "Needs --store_dir")
		flag.Usage()
		return
	}

	if *templatePage != "" {
		ukuleleweb.PageTmpl = template.Must(ukuleleweb.PageTmpl.ParseGlob(*templatePage))
	}

	if *templateEdit != "" {
		ukuleleweb.EditTmpl = template.Must(ukuleleweb.EditTmpl.ParseGlob(*templateEdit))
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
