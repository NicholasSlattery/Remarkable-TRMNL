//go:build ignore

package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var include = []string{
	"dist/trmnl-remarkable-app",
	"device/install.sh",
	"device/uninstall.sh",
	"device/recover-stock.sh",
	"device/diagnostics.sh",
	"device/test-on-device.sh",
	"README.md",
	"LICENSE",
	"THIRD_PARTY_NOTICES.md",
	"docs/trmnl-setup.md",
	"docs/privacy.md",
	"docs/compatibility.md",
}

// device scripts live together in the repository, but existing installations
// expect them at the root of /home/root/trmnl-remarkable. Keep the archive
// names stable so a new installer can still recover or uninstall an older
// installation using the same absolute paths.
var archiveNames = map[string]string{
	"device/install.sh":        "install.sh",
	"device/uninstall.sh":      "uninstall.sh",
	"device/recover-stock.sh":  "recover-stock.sh",
	"device/diagnostics.sh":    "diagnostics.sh",
	"device/test-on-device.sh": "test-on-device.sh",
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: go run scripts/package-device.go ROOT OUTPUT.tar.gz [EXTRA_PATH...]")
		os.Exit(2)
	}
	// Version-specific documents (the validation record) are passed in so the
	// include list does not have to be edited for every release.
	include = append(include, os.Args[3:]...)
	root, err := filepath.Abs(os.Args[1])
	check(err)
	output, err := filepath.Abs(os.Args[2])
	check(err)
	check(os.MkdirAll(filepath.Dir(output), 0o755))
	f, err := os.Create(output)
	check(err)
	gz, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	check(err)
	gz.Name = ""
	gz.ModTime = time.Unix(0, 0).UTC()
	tw := tar.NewWriter(gz)
	var paths []string
	for _, item := range include {
		full := filepath.Join(root, filepath.FromSlash(item))
		info, statErr := os.Stat(full)
		check(statErr)
		if info.IsDir() {
			check(filepath.Walk(full, func(path string, _ os.FileInfo, walkErr error) error {
				if walkErr == nil {
					paths = append(paths, path)
				}
				return walkErr
			}))
		} else {
			paths = append(paths, full)
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		info, statErr := os.Stat(path)
		check(statErr)
		rel, relErr := filepath.Rel(root, path)
		check(relErr)
		name := filepath.ToSlash(rel)
		if archiveName, ok := archiveNames[name]; ok {
			name = archiveName
		}
		mode := int64(0o644)
		if info.IsDir() {
			mode = 0o755
			name += "/"
		} else if strings.HasSuffix(name, ".sh") || strings.HasSuffix(name, "/backend/entry") {
			mode = 0o755
		}
		h := &tar.Header{Name: name, Mode: mode, Size: info.Size(), ModTime: time.Unix(0, 0).UTC(), Uid: 0, Gid: 0, Uname: "root", Gname: "root"}
		if info.IsDir() {
			h.Typeflag = tar.TypeDir
			h.Size = 0
		} else {
			h.Typeflag = tar.TypeReg
		}
		check(tw.WriteHeader(h))
		if !info.IsDir() {
			in, openErr := os.Open(path)
			check(openErr)
			_, copyErr := io.Copy(tw, in)
			closeErr := in.Close()
			check(copyErr)
			check(closeErr)
		}
	}
	check(tw.Close())
	check(gz.Close())
	check(f.Close())
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
