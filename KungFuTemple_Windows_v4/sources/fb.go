package main

import "math"

// Framebuffer logiciel 32 bits (0xRRGGBB).
type FB struct {
	W, H int
	Pix  []uint32
}

func NewFB(w, h int) *FB { return &FB{w, h, make([]uint32, w*h)} }

func (f *FB) Clear(c uint32) {
	for i := range f.Pix {
		f.Pix[i] = c
	}
}

func ch(c uint32) (int, int, int) { return int(c >> 16 & 255), int(c >> 8 & 255), int(c & 255) }
func rgb(r, g, b int) uint32 {
	if r < 0 {
		r = 0
	} else if r > 255 {
		r = 255
	}
	if g < 0 {
		g = 0
	} else if g > 255 {
		g = 255
	}
	if b < 0 {
		b = 0
	} else if b > 255 {
		b = 255
	}
	return uint32(r)<<16 | uint32(g)<<8 | uint32(b)
}

func blend(d, s uint32, a float64) uint32 {
	if a >= 1 {
		return s
	}
	if a <= 0 {
		return d
	}
	ia := int(a * 256)
	dr, dg, db := ch(d)
	sr, sg, sb := ch(s)
	return rgb((dr*(256-ia)+sr*ia)>>8, (dg*(256-ia)+sg*ia)>>8, (db*(256-ia)+sb*ia)>>8)
}

func addc(d, s uint32, a float64) uint32 {
	dr, dg, db := ch(d)
	sr, sg, sb := ch(s)
	return rgb(dr+int(float64(sr)*a), dg+int(float64(sg)*a), db+int(float64(sb)*a))
}

func shade(c uint32, k float64) uint32 {
	r, g, b := ch(c)
	return rgb(int(float64(r)*k), int(float64(g)*k), int(float64(b)*k))
}

func lighten(c uint32, k float64) uint32 { return blend(c, 0xffffff, k) }
func mix(a, b uint32, t float64) uint32  { return blend(a, b, t) }

func (f *FB) Rect(x, y, w, h int, c uint32) {
	x0, y0, x1, y1 := max(x, 0), max(y, 0), min(x+w, f.W), min(y+h, f.H)
	for yy := y0; yy < y1; yy++ {
		row := f.Pix[yy*f.W : yy*f.W+f.W]
		for xx := x0; xx < x1; xx++ {
			row[xx] = c
		}
	}
}

func (f *FB) RectA(x, y, w, h int, c uint32, a float64) {
	x0, y0, x1, y1 := max(x, 0), max(y, 0), min(x+w, f.W), min(y+h, f.H)
	for yy := y0; yy < y1; yy++ {
		row := f.Pix[yy*f.W : yy*f.W+f.W]
		for xx := x0; xx < x1; xx++ {
			row[xx] = blend(row[xx], c, a)
		}
	}
}

func (f *FB) VGrad(x, y, w, h int, c1, c2 uint32) {
	for i := 0; i < h; i++ {
		t := 0.0
		if h > 1 {
			t = float64(i) / float64(h-1)
		}
		f.Rect(x, y+i, w, 1, mix(c1, c2, t))
	}
}

func (f *FB) HGrad(x, y, w, h int, c1, c2 uint32) {
	for i := 0; i < w; i++ {
		t := 0.0
		if w > 1 {
			t = float64(i) / float64(w-1)
		}
		f.Rect(x+i, y, 1, h, mix(c1, c2, t))
	}
}

// Capsule anticrénelée (segment épais à bouts ronds).
func (f *FB) Capsule(x1, y1, x2, y2, r float64, c uint32, alpha float64) {
	if alpha <= 0 || r <= 0 {
		return
	}
	minx := int(math.Floor(math.Min(x1, x2) - r - 1))
	maxx := int(math.Ceil(math.Max(x1, x2) + r + 1))
	miny := int(math.Floor(math.Min(y1, y2) - r - 1))
	maxy := int(math.Ceil(math.Max(y1, y2) + r + 1))
	minx, miny = max(minx, 0), max(miny, 0)
	maxx, maxy = min(maxx, f.W-1), min(maxy, f.H-1)
	dx, dy := x2-x1, y2-y1
	ll := dx*dx + dy*dy
	for y := miny; y <= maxy; y++ {
		py := float64(y) + .5
		row := f.Pix[y*f.W : y*f.W+f.W]
		for x := minx; x <= maxx; x++ {
			px := float64(x) + .5
			t := 0.0
			if ll > 0 {
				t = ((px-x1)*dx + (py-y1)*dy) / ll
				if t < 0 {
					t = 0
				} else if t > 1 {
					t = 1
				}
			}
			ex, ey := px-(x1+t*dx), py-(y1+t*dy)
			d := math.Sqrt(ex*ex + ey*ey)
			cov := r - d + .5
			if cov <= 0 {
				continue
			}
			if cov > 1 {
				cov = 1
			}
			row[x] = blend(row[x], c, cov*alpha)
		}
	}
}

func (f *FB) Circle(cx, cy, r float64, c uint32, a float64) { f.Capsule(cx, cy, cx, cy, r, c, a) }

// Anneau anticrénelé.
func (f *FB) Ring(cx, cy, r1, r2 float64, c uint32, a float64) {
	minx, maxx := max(int(cx-r2-1), 0), min(int(cx+r2+1), f.W-1)
	miny, maxy := max(int(cy-r2-1), 0), min(int(cy+r2+1), f.H-1)
	for y := miny; y <= maxy; y++ {
		for x := minx; x <= maxx; x++ {
			d := math.Hypot(float64(x)+.5-cx, float64(y)+.5-cy)
			cov := math.Min(r2-d+.5, d-r1+.5)
			if cov <= 0 {
				continue
			}
			if cov > 1 {
				cov = 1
			}
			i := y*f.W + x
			f.Pix[i] = blend(f.Pix[i], c, cov*a)
		}
	}
}

func (f *FB) Ellipse(cx, cy, rx, ry float64, c uint32, a float64) {
	if rx <= 0 || ry <= 0 {
		return
	}
	minx, maxx := max(int(cx-rx-1), 0), min(int(cx+rx+1), f.W-1)
	miny, maxy := max(int(cy-ry-1), 0), min(int(cy+ry+1), f.H-1)
	for y := miny; y <= maxy; y++ {
		for x := minx; x <= maxx; x++ {
			nx, ny := (float64(x)+.5-cx)/rx, (float64(y)+.5-cy)/ry
			d := math.Sqrt(nx*nx + ny*ny)
			cov := (1 - d) * math.Min(rx, ry)
			if cov <= 0 {
				continue
			}
			if cov > 1 {
				cov = 1
			}
			i := y*f.W + x
			f.Pix[i] = blend(f.Pix[i], c, cov*a)
		}
	}
}

// Halo lumineux additif.
func (f *FB) Glow(cx, cy, r float64, c uint32, k float64) {
	minx, maxx := max(int(cx-r), 0), min(int(cx+r), f.W-1)
	miny, maxy := max(int(cy-r), 0), min(int(cy+r), f.H-1)
	r2 := r * r
	for y := miny; y <= maxy; y++ {
		dy := float64(y) - cy
		row := f.Pix[y*f.W : y*f.W+f.W]
		for x := minx; x <= maxx; x++ {
			dx := float64(x) - cx
			d2 := dx*dx + dy*dy
			if d2 >= r2 {
				continue
			}
			t := 1 - math.Sqrt(d2)/r
			row[x] = addc(row[x], c, t*t*k)
		}
	}
}

// Polygone plein (remplissage par balayage, pair-impair).
func (f *FB) Poly(xs, ys []float64, c uint32, a float64) {
	n := len(xs)
	if n < 3 {
		return
	}
	miny, maxy := ys[0], ys[0]
	for _, y := range ys {
		miny, maxy = math.Min(miny, y), math.Max(maxy, y)
	}
	y0, y1 := max(int(miny), 0), min(int(maxy), f.H-1)
	var xsInt []float64
	for y := y0; y <= y1; y++ {
		py := float64(y) + .5
		xsInt = xsInt[:0]
		for i := 0; i < n; i++ {
			j := (i + 1) % n
			if (ys[i] <= py && ys[j] > py) || (ys[j] <= py && ys[i] > py) {
				xsInt = append(xsInt, xs[i]+(py-ys[i])/(ys[j]-ys[i])*(xs[j]-xs[i]))
			}
		}
		for i := 1; i < len(xsInt); i++ {
			for j := i; j > 0 && xsInt[j] < xsInt[j-1]; j-- {
				xsInt[j], xsInt[j-1] = xsInt[j-1], xsInt[j]
			}
		}
		for i := 0; i+1 < len(xsInt); i += 2 {
			xa, xb := max(int(xsInt[i]+.5), 0), min(int(xsInt[i+1]+.5), f.W)
			row := f.Pix[y*f.W : y*f.W+f.W]
			for x := xa; x < xb; x++ {
				row[x] = blend(row[x], c, a)
			}
		}
	}
}

// Agrandissement entier au plus proche.
func (f *FB) UpscaleFrom(src *FB, k int) {
	for y := 0; y < src.H; y++ {
		srow := src.Pix[y*src.W : y*src.W+src.W]
		for ky := 0; ky < k; ky++ {
			dy := y*k + ky
			if dy >= f.H {
				break
			}
			drow := f.Pix[dy*f.W : dy*f.W+f.W]
			for x, c := range srow {
				for kx := 0; kx < k; kx++ {
					if dx := x*k + kx; dx < f.W {
						drow[dx] = c
					}
				}
			}
		}
	}
}

// Filtre CRT : lignes de balayage et léger vignettage.
func (f *FB) CRT(strong bool) {
	fac := [4]float64{1, .96, .86, .55}
	if !strong {
		fac = [4]float64{1, 1, .93, .78}
	}
	col := make([]float64, f.W)
	for x := range col {
		vx := math.Abs(float64(x)/float64(f.W)*2 - 1)
		col[x] = 1 - .18*vx*vx*vx*vx
	}
	ck := make([]uint32, f.W)
	for y := 0; y < f.H; y++ {
		vy := math.Abs(float64(y)/float64(f.H)*2 - 1)
		rk := fac[y%4] * (1 - .18*vy*vy*vy*vy)
		for x := range ck {
			ck[x] = uint32(rk * col[x] * 256)
		}
		row := f.Pix[y*f.W : y*f.W+f.W]
		for x, c := range row {
			k := ck[x]
			row[x] = ((c>>16&255)*k>>8)<<16 | ((c>>8&255)*k>>8)<<8 | (c&255)*k>>8
		}
	}
}
