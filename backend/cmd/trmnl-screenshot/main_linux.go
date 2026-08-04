// trmnl-screenshot captures the Xochitl QImage exposed by framebuffer-spy.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 4 {
		fatal("usage: trmnl-screenshot PID 0xADDRESS OUTPUT.png")
	}
	pid := number(os.Args[1], 10)
	address := number64(os.Args[2], 0)
	output := os.Args[3]
	const width, height, stride = 1620, 2160, 6528
	f, err := os.Open(fmt.Sprintf("/proc/%d/mem", pid))
	if err != nil {
		fatal(err.Error())
	}
	defer f.Close()
	raw := make([]byte, stride*height)
	n, err := f.ReadAt(raw, address)
	if err != nil {
		fatal(fmt.Sprintf("read framebuffer: %v (%d bytes)", err, n))
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := y*stride + x*4
			img.SetNRGBA(x, y, color.NRGBA{R: raw[i+2], G: raw[i+1], B: raw[i], A: 255})
		}
	}
	tmp := output + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		fatal(err.Error())
	}
	if err := png.Encode(out, img); err != nil {
		out.Close()
		fatal(err.Error())
	}
	if err := out.Sync(); err != nil {
		out.Close()
		fatal(err.Error())
	}
	if err := out.Close(); err != nil {
		fatal(err.Error())
	}
	if err := os.Rename(tmp, output); err != nil {
		fatal(err.Error())
	}
	fmt.Printf("captured=%s width=%d height=%d\n", output, width, height)
}
func number(s string, base int) int {
	n, e := strconv.ParseInt(strings.TrimSpace(s), base, 32)
	if e != nil {
		fatal(e.Error())
	}
	return int(n)
}
func number64(s string, base int) int64 {
	n, e := strconv.ParseUint(strings.TrimSpace(s), base, 64)
	if e != nil {
		fatal(e.Error())
	}
	return int64(n)
}
func fatal(s string) { fmt.Fprintln(os.Stderr, s); os.Exit(2) }
