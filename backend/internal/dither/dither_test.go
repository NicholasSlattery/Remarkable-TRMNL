package dither

import (
	"image"
	"image/color"
	"testing"

	"trmnl-remarkable/backend/internal/config"
)

func testPalette(t *testing.T) Palette {
	t.Helper()
	p, err := NewPalette(config.DefaultDitherPalette, config.ParseHexColor)
	if err != nil {
		t.Fatalf("default palette is invalid: %v", err)
	}
	return p
}

func TestApplyEmitsOnlyPaletteColours(t *testing.T) {
	p := testPalette(t)
	src := image.NewNRGBA(image.Rect(0, 0, 64, 48))
	for y := 0; y < 48; y++ {
		for x := 0; x < 64; x++ {
			// A diagonal gradient is exactly what bands on a colour panel.
			shade := uint8((x*4 + y*2) % 256)
			src.SetNRGBA(x, y, color.NRGBA{R: shade, G: uint8(255 - int(shade)), B: uint8((x * 3) % 256), A: 255})
		}
	}
	out := Apply(src, p)
	allowed := make(map[color.NRGBA]struct{}, len(p))
	for _, c := range p {
		allowed[c] = struct{}{}
	}
	for y := out.Bounds().Min.Y; y < out.Bounds().Max.Y; y++ {
		for x := out.Bounds().Min.X; x < out.Bounds().Max.X; x++ {
			c := out.NRGBAAt(x, y)
			if _, ok := allowed[color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255}]; !ok {
				t.Fatalf("pixel %d,%d is %v, which is outside the palette", x, y, c)
			}
		}
	}
}

func TestApplyPreservesAverageBrightness(t *testing.T) {
	p := testPalette(t)
	src := image.NewNRGBA(image.Rect(0, 0, 80, 80))
	// Mid grey has no exact palette match, so only diffusion can approximate it.
	for y := 0; y < 80; y++ {
		for x := 0; x < 80; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}
	out := Apply(src, p)
	var total int
	for y := 0; y < 80; y++ {
		for x := 0; x < 80; x++ {
			c := out.NRGBAAt(x, y)
			total += (int(c.R) + int(c.G) + int(c.B)) / 3
		}
	}
	mean := total / (80 * 80)
	if mean < 96 || mean > 160 {
		t.Fatalf("dithered mid grey averaged %d; error diffusion is not converging", mean)
	}
}

func TestNeededSkipsPaletteCleanImages(t *testing.T) {
	p := testPalette(t)
	clean := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			clean.SetNRGBA(x, y, p[(x+y)%len(p)])
		}
	}
	if Needed(clean, p) {
		t.Fatal("a palette-clean dashboard was marked as needing dithering")
	}

	gradient := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			gradient.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 8), G: uint8(y * 8), B: 90, A: 255})
		}
	}
	if !Needed(gradient, p) {
		t.Fatal("a gradient was not marked as needing dithering")
	}
}

func TestApplyPreservesAlpha(t *testing.T) {
	p := testPalette(t)
	src := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	src.SetNRGBA(1, 1, color.NRGBA{R: 200, G: 40, B: 40, A: 128})
	out := Apply(src, p)
	if got := out.NRGBAAt(1, 1).A; got != 128 {
		t.Fatalf("alpha = %d, want 128", got)
	}
}

func TestNewPaletteRejectsBadColours(t *testing.T) {
	if _, err := NewPalette([]string{"#zzz"}, config.ParseHexColor); err == nil {
		t.Fatal("an invalid colour was accepted")
	}
	if _, err := NewPalette(nil, config.ParseHexColor); err == nil {
		t.Fatal("an empty palette was accepted")
	}
	p, err := NewPalette([]string{"#fff", "#000000"}, config.ParseHexColor)
	if err != nil {
		t.Fatalf("valid palette rejected: %v", err)
	}
	if len(p) != 2 || p[0].R != 255 || p[1].R != 0 {
		t.Fatalf("palette = %v", p)
	}
}
