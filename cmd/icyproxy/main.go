package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"bustany.org/icyproxy"
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s SOURCES\n", os.Args[0])
}

type Source struct {
	URL string `json:"url"`
}

func parseSources(r io.Reader) (map[string]Source, error) {
	res := map[string]Source{}
	if err := json.NewDecoder(r).Decode(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func parseSourcesFromFile(filename string) (map[string]Source, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	defer fd.Close()

	res, err := parseSources(fd)
	if err != nil {
		return nil, fmt.Errorf("error parsing sources: %w", err)
	}

	return res, nil
}

func makeRequestHandler(s Source) (http.Handler, error) {
	u, err := url.Parse(s.URL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamRequest := http.Request{
			Method: http.MethodGet,
			URL:    u,
		}

		res, err := http.DefaultClient.Do(&upstreamRequest)
		if err != nil {
			http.Error(w, "error contacting upstream", http.StatusBadGateway)
			return
		}

		defer res.Body.Close()

		for k, v := range res.Header {
			w.Header()[k] = v
		}

		wantIcy := r.Header.Get("Icy-Metadata") == "1"
		canIcy := res.Header.Get("icy-metaint") != ""

		var source io.Reader = res.Body

		if wantIcy && !canIcy {
			w.Header().Add("icy-metaint", strconv.Itoa(icyproxy.IcyReaderInterval))
			reader := &icyproxy.IcyReader{AudioData: res.Body}
			reader.SetTitle("test title")
			source = reader
		}

		io.Copy(w, source)
	}), nil
}

func makeHandler(sources map[string]Source) (http.Handler, error) {
	res := http.NewServeMux()

	for sourceId, source := range sources {
		sourceHandler, err := makeRequestHandler(source)
		if err != nil {
			return nil, fmt.Errorf("error setting up handler for source %s: %w", sourceId, err)
		}

		res.Handle("GET /"+sourceId, sourceHandler)
	}

	return res, nil
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

	sources, err := parseSourcesFromFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("Error parsing sources: %s", err)
	}

	handler, err := makeHandler(sources)
	if err != nil {
		log.Fatalf("Error setting up proxy handlers: %s", err)
	}

	log.Printf("Starting server on %s", *listenAddr)

	if err := http.ListenAndServe(*listenAddr, loggingHandler(handler)); err != nil {
		log.Fatalf("Error starting HTTP server: %s", err)
	}
}
