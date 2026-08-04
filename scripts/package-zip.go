//go:build ignore

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go run scripts/package-zip.go INPUT_DIR OUTPUT.zip")
		os.Exit(2)
	}
	root, err := filepath.Abs(os.Args[1])
	checkZip(err)
	output, err := filepath.Abs(os.Args[2])
	checkZip(err)
	var paths []string
	checkZip(filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != root && !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	}))
	sort.Strings(paths)
	f, err := os.Create(output)
	checkZip(err)
	zw := zip.NewWriter(f)
	epoch := time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, path := range paths {
		rel, relErr := filepath.Rel(root, path)
		checkZip(relErr)
		info, statErr := os.Stat(path)
		checkZip(statErr)
		header, headerErr := zip.FileInfoHeader(info)
		checkZip(headerErr)
		header.Name = filepath.ToSlash(rel)
		header.Method = zip.Deflate
		header.SetModTime(epoch)
		if strings.HasSuffix(strings.ToLower(header.Name), ".exe") {
			header.SetMode(0o755)
		} else {
			header.SetMode(0o644)
		}
		out, createErr := zw.CreateHeader(header)
		checkZip(createErr)
		in, openErr := os.Open(path)
		checkZip(openErr)
		_, copyErr := io.Copy(out, in)
		closeErr := in.Close()
		checkZip(copyErr)
		checkZip(closeErr)
	}
	checkZip(zw.Close())
	checkZip(f.Close())
}

func checkZip(err error) {
	if err != nil {
		panic(err)
	}
}
