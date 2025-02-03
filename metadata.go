package icyproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"
)

func fetchMetadata(ctx context.Context, u *url.URL) (map[string]any, error) {
	req := (&http.Request{Method: http.MethodGet, URL: u}).WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	defer res.Body.Close()

	var md map[string]any
	if err := json.NewDecoder(res.Body).Decode(&md); err != nil {
		return nil, fmt.Errorf("error decoding response body: %w", err)
	}

	return md, nil
}

func fetchAndRenderMetadata(ctx context.Context, u *url.URL, tmpl *template.Template) (string, error) {
	md, err := fetchMetadata(ctx, u)
	if err != nil {
		return "", fmt.Errorf("error fetching metadata: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, md); err != nil {
		return "", fmt.Errorf("error rendering metadata template: %w", err)
	}

	return buf.String(), nil
}

func StartMetadataFetcher(ctx context.Context, reader *IcyReader, metadataJsonUrl *url.URL, metadataTmpl *template.Template) {
	go func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("metadata fetcher routine for URL %s panicked: %v", metadataJsonUrl, err)
			}

			for {
				title, err := fetchAndRenderMetadata(ctx, metadataJsonUrl, metadataTmpl)
				if errors.Is(err, context.Canceled) {
					break
				} else if err != nil {
					log.Printf("error fetching metadata from %s: %s", metadataJsonUrl, err)
				}

				reader.SetTitle(title)

				select {
				case <-time.After(10 * time.Second):
				case <-ctx.Done():
				}
			}
		}()
	}(ctx)
}
