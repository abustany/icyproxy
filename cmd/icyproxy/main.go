package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"bustany.org/icyproxy"
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s SOURCES\n", os.Args[0])
}

func loggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %+v", r.Method, r.URL.Path, r.Header)
		h.ServeHTTP(w, r)
	})
}

func main() {
	flag.Usage = usage
	listenAddr := flag.String("listen", ":8080", "HTTP listen address")
	flag.Parse()

	if flag.NArg() != 1 {
		log.Println("Missing required argument: sources")
		usage()
		os.Exit(1)
	}

	sources, err := icyproxy.ParseSourcesFromFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("Error parsing sources: %s", err)
	}

	handler, err := icyproxy.MakeHandler(sources)
	if err != nil {
		log.Fatalf("Error setting up proxy handlers: %s", err)
	}

	log.Printf("Starting server on %s", *listenAddr)

	if err := http.ListenAndServe(*listenAddr, loggingHandler(handler)); err != nil {
		log.Fatalf("Error starting HTTP server: %s", err)
	}
}
