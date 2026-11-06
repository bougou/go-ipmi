//go:build ignore

package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	inputPath  = "data/enterprise-numbers.txt"
	outputPath = "data/enterprise-names.tsv.gz"
)

func main() {
	if err := generate(inputPath, outputPath); err != nil {
		log.Fatal(err)
	}
}

func generate(input, output string) error {
	src, err := os.Open(input)
	if err != nil {
		return err
	}
	defer src.Close()

	tmp, err := os.Create(output + ".tmp")
	if err != nil {
		return err
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	zw, err := gzip.NewWriterLevel(tmp, gzip.BestCompression)
	if err != nil {
		return err
	}
	if err := writeNames(zw, src); err != nil {
		zw.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), output)
}

func writeNames(dst io.Writer, src io.Reader) error {
	scanner := bufio.NewScanner(src)
	seen := make(map[uint32]struct{})
	var pending uint32
	havePending := false
	count := 0

	for scanner.Scan() {
		line := scanner.Text()
		if havePending {
			havePending = false
			if len(line) < 3 || line[0] != ' ' || line[1] != ' ' || line[2] == ' ' {
				continue
			}
			name := line[2:]
			if strings.ContainsRune(name, '\t') {
				return fmt.Errorf("PEN %d organization contains a tab", pending)
			}
			if _, exists := seen[pending]; exists {
				return fmt.Errorf("duplicate PEN %d", pending)
			}
			if _, err := fmt.Fprintf(dst, "%d\t%s\n", pending, name); err != nil {
				return err
			}
			seen[pending] = struct{}{}
			count++
			continue
		}

		if line == "" || line[0] < '0' || line[0] > '9' {
			continue
		}
		end := 0
		for end < len(line) && line[end] >= '0' && line[end] <= '9' {
			end++
		}
		id, err := strconv.ParseUint(line[:end], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid PEN %q: %w", line[:end], err)
		}
		pending = uint32(id)
		havePending = true
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("no PEN entries found")
	}
	return nil
}
