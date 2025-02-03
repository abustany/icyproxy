package icyproxy

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func makeRequestHandler(s Source) (http.Handler, error) {
	u, err := url.Parse(s.URL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	metadataJsonUrl, err := url.Parse(s.MetadataJsonURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing metadata JSON URL: %w", err)
	}

	metadataTmpl, err := template.New("").Parse(s.MetadataFormat)
	if err != nil {
		return nil, fmt.Errorf("error parsing metadata format: %w", err)
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
			w.Header().Add("icy-metaint", strconv.Itoa(IcyReaderInterval))
			reader := &IcyReader{AudioData: res.Body}
			StartMetadataFetcher(r.Context(), reader, metadataJsonUrl, metadataTmpl)
			source = reader
		}

		io.Copy(w, source)
	}), nil
}

func MakeHandler(sources map[string]Source) (http.Handler, error) {
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
