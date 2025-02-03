package icyproxy

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Source struct {
	URL             string `json:"url"`
	MetadataJsonURL string `json:"metadataJsonUrl"`
	MetadataFormat  string `json:"metadataFormat"`
}

func ParseSources(r io.Reader) (map[string]Source, error) {
	res := map[string]Source{}
	if err := json.NewDecoder(r).Decode(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func ParseSourcesFromFile(filename string) (map[string]Source, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	defer fd.Close()

	res, err := ParseSources(fd)
	if err != nil {
		return nil, fmt.Errorf("error parsing sources: %w", err)
	}

	return res, nil
}
