package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:9988", "listen address")
	token := flag.String("token", "local-test", "required test token")
	flag.Parse()
	var sequence atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok\n")) })
	mux.HandleFunc("/api/display", func(w http.ResponseWriter, r *http.Request) { serveDisplay(w, r, *token, sequence.Add(1)) })
	mux.HandleFunc("/api/display/current", func(w http.ResponseWriter, r *http.Request) {
		n := sequence.Load()
		if n < 1 {
			n = 1
		}
		serveDisplay(w, r, *token, n)
	})
	mux.HandleFunc("/image/", func(w http.ResponseWriter, r *http.Request) {
		n, _ := strconv.ParseInt(r.URL.Query().Get("n"), 10, 64)
		landscape := n%2 == 0
		serveImage(w, n, landscape)
	})
	log.Printf("TRMNL mock listening at http://%s", *listen)
	server := &http.Server{Addr: *listen, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func serveDisplay(w http.ResponseWriter, r *http.Request, token string, n int64) {
	if r.Header.Get("access-token") != token {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	host := r.Host
	payload := map[string]any{"status": 0, "image_url": "http://" + host + "/image/test.png?n=" + strconv.FormatInt(n, 10), "filename": fmt.Sprintf("mock-%d.png", n), "image_name": fmt.Sprintf("mock-%d", n), "refresh_rate": 60, "special_function": "sleep", "image_url_timeout": 10}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func serveImage(w http.ResponseWriter, n int64, landscape bool) {
	wid, hei := 1200, 1600
	if landscape {
		wid, hei = 1800, 1000
	}
	img := image.NewNRGBA(image.Rect(0, 0, wid, hei))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.NRGBA{245, 243, 235, 255}}, image.Point{}, draw.Src)
	colors := []color.NRGBA{{0, 0, 0, 255}, {177, 54, 44, 255}, {225, 148, 33, 255}, {31, 104, 76, 255}, {40, 83, 150, 255}, {108, 65, 143, 255}}
	bar := wid / len(colors)
	for i, c := range colors {
		draw.Draw(img, image.Rect(i*bar, 0, (i+1)*bar, hei/3), &image.Uniform{c}, image.Point{}, draw.Src)
	}
	for i := 0; i < 12; i++ {
		shade := uint8(i * 255 / 11)
		x0 := i * wid / 12
		x1 := (i + 1) * wid / 12
		draw.Draw(img, image.Rect(x0, hei/3, x1, 2*hei/3), &image.Uniform{color.NRGBA{shade, shade, shade, 255}}, image.Point{}, draw.Src)
	}
	border := max(8, wid/100)
	draw.Draw(img, image.Rect(0, 0, wid, border), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(0, hei-border, wid, hei), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(0, 0, border, hei), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(wid-border, 0, wid, hei), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Mock-Sequence", strconv.FormatInt(n, 10))
	_ = png.Encode(w, img)
}

func init() { log.SetOutput(os.Stderr); log.SetFlags(log.LstdFlags | log.LUTC) }
