// Package dither maps a full-colour dashboard onto the small palette a colour
// e-ink panel can actually show. Server-rendered dashboards are drawn for LCDs,
// so smooth gradients arrive as banding once the panel quantises them. Error
// diffusion trades that banding for fine noise, which reads better on paper.
package dither

import (
	"errors"
	"image"
	"image/color"
)

// Palette is an ordered set of displayable colours.
type Palette []color.NRGBA

// NewPalette converts #rgb or #rrggbb strings into a palette.
func NewPalette(values []string, parse func(string) (uint8, uint8, uint8, error)) (Palette, error) {
	if len(values) == 0 {
		return nil, errors.New("palette is empty")
	}
	p := make(Palette, 0, len(values))
	for _, value := range values {
		r, g, b, err := parse(value)
		if err != nil {
			return nil, err
		}
		p = append(p, color.NRGBA{R: r, G: g, B: b, A: 255})
	}
	return p, nil
}

// nearest returns the palette entry closest to the requested colour using a
// luminance-weighted distance, which tracks perceived error better than a plain
// RGB distance on a paper-white panel.
func (p Palette) nearest(r, g, b int32) color.NRGBA {
	best := p[0]
	bestDistance := int32(1 << 30)
	for _, candidate := range p {
		dr := r - int32(candidate.R)
		dg := g - int32(candidate.G)
		db := b - int32(candidate.B)
		distance := 2*dr*dr + 4*dg*dg + 3*db*db
		if distance < bestDistance {
			bestDistance = distance
			best = candidate
		}
	}
	return best
}

// Needed reports whether an image contains more distinct colours than the
// palette can represent. Dashboards that are already palette-clean are left
// untouched so text stays crisp.
func Needed(src image.Image, p Palette) bool {
	if len(p) == 0 {
		return false
	}
	allowed := make(map[color.NRGBA]struct{}, len(p))
	for _, c := range p {
		allowed[c] = struct{}{}
	}
	b := src.Bounds()
	// Sampling keeps this cheap on a 1620x2160 panel; a dashboard that needs
	// dithering has far more than one offending pixel.
	const step = 4
	for y := b.Min.Y; y < b.Max.Y; y += step {
		for x := b.Min.X; x < b.Max.X; x += step {
			c := color.NRGBAModel.Convert(src.At(x, y)).(color.NRGBA)
			c.A = 255
			if _, ok := allowed[c]; !ok {
				return true
			}
		}
	}
	return false
}

// Apply runs Floyd-Steinberg error diffusion over src and returns a new image
// drawn only from p. Alpha is preserved from the source.
func Apply(src image.Image, p Palette) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	if len(p) == 0 {
		return dst
	}
	width, height := b.Dx(), b.Dy()
	if width == 0 || height == 0 {
		return dst
	}

	// Two rolling error rows are enough for the Floyd-Steinberg kernel and keep
	// allocation proportional to width rather than to the whole image.
	current := make([]int32, (width+2)*3)
	next := make([]int32, (width+2)*3)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			src := color.NRGBAModel.Convert(src.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
			i := (x + 1) * 3
			r := clamp(int32(src.R) + current[i])
			g := clamp(int32(src.G) + current[i+1])
			bl := clamp(int32(src.B) + current[i+2])

			chosen := p.nearest(r, g, bl)
			dst.SetNRGBA(b.Min.X+x, b.Min.Y+y, color.NRGBA{R: chosen.R, G: chosen.G, B: chosen.B, A: src.A})

			er := r - int32(chosen.R)
			eg := g - int32(chosen.G)
			eb := bl - int32(chosen.B)

			spread(current, i+3, er, eg, eb, 7)
			spread(next, i-3, er, eg, eb, 3)
			spread(next, i, er, eg, eb, 5)
			spread(next, i+3, er, eg, eb, 1)
		}
		current, next = next, current
		for i := range next {
			next[i] = 0
		}
	}
	return dst
}

func spread(row []int32, index int, er, eg, eb, weight int32) {
	if index < 0 || index+2 >= len(row) {
		return
	}
	row[index] += er * weight / 16
	row[index+1] += eg * weight / 16
	row[index+2] += eb * weight / 16
}

func clamp(v int32) int32 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
