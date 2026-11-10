// Package iana looks up organization names in an embedded snapshot of the
// IANA Private Enterprise Numbers registry. Lookups work offline.
package iana

import (
	"bufio"
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

//go:generate go run ./internal/generate

//go:embed data/enterprise-names.tsv.gz
var compressedRegistry []byte

var registryNames = sync.OnceValue(func() map[uint32]string {
	r, err := gzip.NewReader(bytes.NewReader(compressedRegistry))
	if err != nil {
		panic(fmt.Errorf("iana: open embedded registry: %w", err))
	}
	defer r.Close()
	names, err := parseNames(r)
	if err != nil {
		panic(fmt.Errorf("iana: read embedded registry: %w", err))
	}
	return names
})

// Lookup returns the registered organization for pen, or ("", false) if the
// snapshot has no entry. The first call loads the full registry into memory.
// Concurrent calls share the same registry for the lifetime of the process.
func Lookup(pen uint32) (string, bool) {
	name, ok := registryNames()[pen]
	return name, ok
}

func parseNames(r io.Reader) (map[uint32]string, error) {
	names := make(map[uint32]string)
	scanner := bufio.NewScanner(r)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		idText, name, ok := strings.Cut(scanner.Text(), "\t")
		if !ok || name == "" {
			return nil, fmt.Errorf("invalid name record at line %d", lineNumber)
		}
		id, err := strconv.ParseUint(idText, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid PEN at line %d: %w", lineNumber, err)
		}
		pen := uint32(id)
		if _, exists := names[pen]; exists {
			return nil, fmt.Errorf("duplicate PEN %d at line %d", id, lineNumber)
		}
		names[pen] = name
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return names, nil
}
