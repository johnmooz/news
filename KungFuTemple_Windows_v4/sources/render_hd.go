package main

import (
	"math"
	"math/rand"
)

// Rendu HD : 1024x960, personnages en squelettes animés et ombrés.

const S = 4.0

type Pose [12]float64

const (
	pHX = iota
	pHY
	pTor
	pBT
	pBS
	pFT
	pFS
	pBU
	pBF
	pFU
	pFF
	pHead
)

func poseTarget(name string, ph float64) (Pose, bool) {
	var p Pose
	g := true
	guard := func() { p[pBU], p[pBF], p[pFU], p[pFF] = -.35, .95, .75, 2.45 }
	walk := func() {
		s, c := math.Sin(ph), math.Cos(ph)
		p[pTor] = .1
		p[pFT], p[pBT] = .42*s, -.42*s
		p[pFS] = math.Min(p[pFT], p[pFT]-.25-.3*c)
		p[pBS] = math.Min(p[pBT], p[pBT]-.25+.3*c)
	}
	crouchL := func() { p[pBT], p[pBS], p[pFT], p[pFS] = 1.0, -1.3, 1.45, -.7 }
	switch name {
	case "walk":
		walk()
		guard()
		p[pBU] += .15 * math.Sin(ph)
	case "kickprep":
		p = Pose{0, 0, -.2, -.05, 0, 1.45, .25, -.7, .5, .55, 2.3, 0}
	case "kick":
		p = Pose{0, 0, -.45, .05, -.02, 1.72, 1.78, -1.1, -.7, .45, 2.2, 0}
	case "punch":
		p = Pose{0, 0, .22, -.48, -.22, .48, .22, -.55, 1.3, 1.6, 1.56, 0}
	case "crouch":
		crouchL()
		p[pTor] = .6
		p[pBU], p[pBF], p[pFU], p[pFF] = .2, 1.6, 1.0, 2.4
	case "lowkick":
		p = Pose{0, 0, -.5, 1.0, -1.3, 1.52, 1.56, -.9, -.3, .7, 2.2, 0}
	case "lowpunch":
		crouchL()
		p[pTor] = .65
		p[pBU], p[pBF], p[pFU], p[pFF] = 0, 1.4, 1.72, 1.68
	case "jump":
		g = false
		p = Pose{0, 9, .1, 1.2, -.8, 1.6, -.2, 0, 0, 0, 0, 0}
		guard()
	case "airkick":
		g = false
		p = Pose{0, 10, -.35, 1.1, -.9, 1.62, 1.66, -1.2, -.8, .6, 2.4, 0}
	case "hurt":
		p = Pose{0, 0, -.45, -.3, -.2, .25, .1, 2.7, 3.2, 2.4, 2.9, -.3}
	case "grab":
		walk()
		p[pTor] = .3
		p[pBU], p[pBF], p[pFU], p[pFF] = 1.45, 1.6, 1.3, 1.45
	case "throwUp":
		walk()
		p[pTor] = -.15
		p[pBU], p[pBF], p[pFU], p[pFF] = -.5, .8, -2.5, -2.9
	case "throwFwd":
		p = Pose{0, 0, .25, -.3, -.1, .35, .1, -.8, .3, 1.5, 1.4, 0}
	case "cast":
		p = Pose{0, 0, .1, -.22, -.05, .3, .12, 1.4, 1.9, 1.6, 1.9, 0}
	case "lying":
		g = false
		p = Pose{0, 2.4, -1.5, 1.52, 1.56, 1.45, 1.6, -1.6, -1.6, .9, 1.3, -.2}
	default: // stand
		p = Pose{0, 0, .05, -.22, -.05, .3, .12, 0, 0, 0, 0, 0}
		guard()
		p[pFF] += .05 * math.Sin(float64(tick)*.08)
	}
	return p, g
}

type FigOpt struct {
	Scale, HeadK, Bulk, Alpha float64
	Flash, Robe, Shade        bool
}

type Joints struct {
	Hip, Neck, Head, BHand, FHand, FElb, FFoot [2]float64
	HR                                         float64
}

func animStep(a *Anim, t Pose, k float64) Pose {
	if !a.Init {
		a.Cur, a.Init = t, true
	}
	for i := range a.Cur {
		a.Cur[i] += (t[i] - a.Cur[i]) * k
	}
	return a.Cur
}

const ink = 0x160c0c

func drawFigure(f *FB, sx, sy, face float64, cur Pose, grounded bool, rot float64, pal Pal, o FigOpt) Joints {
	sc := o.Scale
	if o.HeadK == 0 {
		o.HeadK = 1
	}
	if o.Bulk == 0 {
		o.Bulk = 1
	}
	if o.Alpha == 0 {
		o.Alpha = 1
	}
	th, sh, tor, up, fo, hr := 7.2*sc, 7.2*sc, 10.5*sc, 5.4*sc, 5.2*sc, 3.4*sc*o.HeadK
	dirv := func(a float64) (float64, float64) { return math.Sin(a), -math.Cos(a) }
	p := cur
	hip := [2]float64{p[pHX] * sc, p[pHY] * sc}
	leg := func(t, s float64) (k, ft [2]float64) {
		dx, dy := dirv(t)
		k = [2]float64{hip[0] + dx*th, hip[1] + dy*th}
		dx, dy = dirv(s)
		ft = [2]float64{k[0] + dx*sh, k[1] + dy*sh}
		return
	}
	if grounded {
		hip[1] = 0
		_, bf := leg(p[pBT], p[pBS])
		_, ff := leg(p[pFT], p[pFS])
		hip[1] = -math.Min(bf[1], ff[1]) + 1.1*sc
	}
	bk, bfoot := leg(p[pBT], p[pBS])
	fk, ffoot := leg(p[pFT], p[pFS])
	td := [2]float64{math.Sin(p[pTor]), math.Cos(p[pTor])}
	neck := [2]float64{hip[0] + td[0]*tor, hip[1] + td[1]*tor}
	sho := [2]float64{hip[0] + td[0]*(tor-1.6*sc), hip[1] + td[1]*(tor-1.6*sc)}
	arm := func(u, fa float64) (e, h [2]float64) {
		dx, dy := dirv(u)
		e = [2]float64{sho[0] + dx*up, sho[1] + dy*up}
		dx, dy = dirv(fa)
		h = [2]float64{e[0] + dx*fo, e[1] + dy*fo}
		return
	}
	be, bh := arm(p[pBU], p[pBF])
	fe, fh := arm(p[pFU], p[pFF])
	ha := p[pTor] + p[pHead]
	hc := [2]float64{neck[0] + math.Sin(ha)*hr*1.05, neck[1] + math.Cos(ha)*hr*1.05}

	// pivot de rotation (roulade, chute)
	piv := [2]float64{0, hip[1]}
	cr, sr := math.Cos(rot), math.Sin(rot)
	tf := func(q [2]float64) (float64, float64) {
		x, y := q[0]-piv[0], q[1]-piv[1]
		x, y = x*cr+y*sr, -x*sr+y*cr
		x, y = x+piv[0], y+piv[1]
		return sx + face*x*S, sy - y*S
	}
	add := func(a [2]float64, dx, dy float64) [2]float64 { return [2]float64{a[0] + dx, a[1] + dy} }
	col := func(c uint32, dark bool) uint32 {
		if dark {
			c = shade(c, .7)
		}
		if o.Flash {
			c = lighten(c, .75)
		}
		return c
	}
	al := o.Alpha
	limb := func(a, b [2]float64, r float64, c uint32, dark bool) {
		x1, y1 := tf(a)
		x2, y2 := tf(b)
		f.Capsule(x1, y1, x2, y2, r*S+2.2, ink, al)
		cc := col(c, dark)
		f.Capsule(x1, y1, x2, y2, r*S, cc, al)
		f.Capsule(x1-r*1.1, y1-r*1.3, x2-r*1.1, y2-r*1.3, r*S*.38, lighten(cc, .4), al*.55)
	}
	ball := func(a [2]float64, r float64, c uint32, dark bool) {
		x, y := tf(a)
		f.Circle(x, y, r*S+2.2, ink, al)
		cc := col(c, dark)
		f.Circle(x, y, r*S, cc, al)
		f.Circle(x-r*1.4, y-r*1.4, r*S*.4, lighten(cc, .45), al*.5)
	}
	foot := func(ft [2]float64, shinA float64, dark bool) {
		dx, dy := math.Cos(shinA), math.Sin(shinA)
		toe := [2]float64{ft[0] + dx*2.6*sc, ft[1] + dy*2.6*sc}
		limb(add(ft, -dx*.4*sc, -dy*.4*sc), toe, 1.15*sc, pal.Shoe, dark)
	}
	hand := func(h [2]float64, dark bool) { ball(h, 1.65*sc, pal.Skin, dark) }

	// bras arrière
	limb(sho, be, 1.75*sc*o.Bulk, pal.Top, true)
	limb(be, bh, 1.5*sc*o.Bulk, pal.Skin, true)
	hand(bh, true)
	// jambe arrière
	limb(hip, bk, 2.3*sc*o.Bulk, pal.Pants, true)
	limb(bk, bfoot, 2.0*sc*o.Bulk, pal.Pants, true)
	foot(bfoot, p[pBS], true)
	// robe
	if o.Robe {
		n := [2]float64{td[1], -td[0]}
		pts := [][2]float64{add(hip, n[0]*3.8*sc+td[0]*2*sc, n[1]*3.8*sc+td[1]*2*sc), add(hip, -n[0]*3.8*sc+td[0]*2*sc, -n[1]*3.8*sc+td[1]*2*sc), {hip[0] - 6*sc, 1.5 * sc}, {hip[0] + 6.5*sc, 1.5 * sc}}
		xs, ys := make([]float64, 4), make([]float64, 4)
		for i, q := range pts {
			xs[i], ys[i] = tf(q)
		}
		f.Poly(xs, ys, col(pal.Top, false), al)
		ys2 := []float64{ys[2] - 6, ys[3] - 6, ys[3], ys[2]}
		xs2 := []float64{xs[2], xs[3], xs[3], xs[2]}
		f.Poly(xs2, ys2, col(pal.Belt, false), al)
	}
	// jambe avant
	limb(hip, fk, 2.3*sc*o.Bulk, pal.Pants, false)
	limb(fk, ffoot, 2.0*sc*o.Bulk, pal.Pants, false)
	foot(ffoot, p[pFS], false)
	// torse
	limb(add(hip, td[0]*.8*sc, td[1]*.8*sc), add(neck, -td[0]*1.2*sc, -td[1]*1.2*sc), 3.6*sc*o.Bulk, pal.Top, false)
	n := [2]float64{td[1], -td[0]}
	{
		chest := add(hip, td[0]*tor*.45, td[1]*tor*.45)
		lap := col(shade(pal.Top, .72), false)
		for _, sd := range []float64{1, -1} {
			x1, y1 := tf(add(neck, n[0]*sd*1.6*sc, n[1]*sd*1.6*sc))
			x2, y2 := tf(chest)
			f.Capsule(x1, y1, x2, y2, .5*S*sc, lap, al)
		}
		bc := add(hip, td[0]*1.7*sc, td[1]*1.7*sc)
		bw := 3.7 * sc * o.Bulk
		limb(add(bc, n[0]*bw, n[1]*bw), add(bc, -n[0]*bw, -n[1]*bw), 1.0*sc, pal.Belt, false)
		kn := add(bc, n[0]*1.4*sc, n[1]*1.4*sc)
		limb(kn, add(kn, .6*sc, -3*sc), .55*sc, pal.Belt, false)
	}
	// tête
	limb(add(neck, -td[0]*1.5*sc, -td[1]*1.5*sc), add(neck, td[0]*.5*sc, td[1]*.5*sc), 1.35*sc, pal.Skin, false)
	hx, hy := tf(hc)
	R := hr * S
	f.Circle(hx, hy, R+2.2, ink, al)
	skin := col(pal.Skin, false)
	if pal.Bald {
		f.Circle(hx, hy, R, skin, al)
		f.Circle(hx-face*R*.35, hy-R*.45, R*.35, lighten(skin, .5), al*.6)
	} else {
		f.Circle(hx, hy, R, col(pal.Hair, false), al)
		fx, fy := tf(add(hc, hr*.32, -hr*.22))
		f.Circle(fx, fy, R*.84, skin, al)
		f.Circle(fx+face*R*.15, fy-R*.3, R*.28, lighten(skin, .4), al*.5)
	}
	ex, ey := tf(add(hc, hr*.62, hr*.12))
	f.Circle(ex, ey, .45*S*sc*o.HeadK, ink, al)
	bx1, by1 := tf(add(hc, hr*.35, hr*.42))
	bx2, by2 := tf(add(hc, hr*.85, hr*.36))
	f.Capsule(bx1, by1, bx2, by2, .32*S*sc*o.HeadK, shade(col(pal.Hair, false), .6), al)
	if pal.HasBeard {
		bd := col(pal.Beard, false)
		cx, cy := tf(add(hc, hr*.5, -hr*.7))
		f.Circle(cx, cy, R*.55, bd, al)
		if pal.Beard == 0xf0f0f0 {
			lx, ly := tf(add(hc, hr*.6, -hr*1.9))
			f.Capsule(cx, cy, lx, ly, R*.35, bd, al)
		}
	}
	if pal.HasBand {
		bd := col(pal.Band, false)
		x1, y1 := tf(add(hc, -hr*.95, hr*.25))
		x2, y2 := tf(add(hc, hr*.95, hr*.35))
		f.Capsule(x1, y1, x2, y2, .6*S*sc, bd, al)
		w := math.Sin(float64(tick)*.25 + sx*.01)
		k := add(hc, -hr*.95, hr*.2)
		t1 := add(k, -3.6*sc, -.6*sc+w*1.2*sc)
		t2 := add(k, -3.1*sc, -2.2*sc+math.Sin(float64(tick)*.25+1)*1.1*sc)
		for _, t := range [][2]float64{t1, t2} {
			ax, ay := tf(k)
			tx, ty := tf(t)
			f.Capsule(ax, ay, tx, ty, .5*S*sc, bd, al)
		}
	}
	// bras avant
	limb(sho, fe, 1.8*sc*o.Bulk, pal.Top, false)
	limb(fe, fh, 1.55*sc*o.Bulk, pal.Skin, false)
	hand(fh, false)

	j := Joints{HR: R}
	j.Hip[0], j.Hip[1] = tf(hip)
	j.Neck[0], j.Neck[1] = tf(neck)
	j.Head[0], j.Head[1] = hx, hy
	j.BHand[0], j.BHand[1] = tf(bh)
	j.FHand[0], j.FHand[1] = tf(fh)
	j.FElb[0], j.FElb[1] = tf(fe)
	j.FFoot[0], j.FFoot[1] = tf(ffoot)
	return j
}

/* ---------- Thèmes et décor ---------- */

type Theme struct {
	SkyT, SkyB, Sun, Far, Near, Mist, WallT, WallB, Glow uint32
	Night                                                bool
	Amb                                                  int // 0 poussière 1 lucioles 2 pétales 3 braises
}

var themes = []Theme{{},
	{0x2a1f4a, 0xf08a50, 0xffd890, 0x7a4a6a, 0x3a2440, 0xf0b090, 0xeedcb8, 0xc8aa84, 0xffb050, false, 0},
	{0x050920, 0x1c2c5c, 0xf0f0ff, 0x1e2a50, 0x0c1230, 0x5068a0, 0xc0c4d8, 0x8a90b0, 0xffc060, true, 1},
	{0x4a8aa8, 0xd8f0e0, 0xfff8e0, 0x6a9a8a, 0x2a5a4a, 0xe8fff0, 0xe0ead4, 0xa8bea0, 0xffd080, false, 2},
	{0x140820, 0x4a2858, 0xd0c0ff, 0x3a2048, 0x1a0c24, 0x8060a0, 0xbcaccc, 0x806890, 0xff9070, true, 1},
	{0x300808, 0xc83818, 0xff7040, 0x6a1818, 0x2a0808, 0xff9060, 0xe4c4a0, 0xb08060, 0xff7030, false, 3},
}

const (
	UW     = 512
	PW     = 2048
	PY0    = 336
	PH     = 408
	keyCol = 0x123457
)

var (
	bgLevel = -1
	bgUnit  *FB
	pano    *FB
	amb     []*Part
)

func ridge(x float64, base float64, a [][3]float64) float64 {
	y := base
	for _, t := range a {
		y += t[0] * math.Sin(2*math.Pi*x*t[1]/PW+t[2])
	}
	return y
}

func buildPano(th Theme) *FB {
	p := NewFB(PW, PH)
	p.VGrad(0, 0, PW, PH, th.SkyT, th.SkyB)
	r := rand.New(rand.NewSource(7))
	if th.Night {
		for i := 0; i < 420; i++ {
			x, y := r.Intn(PW), r.Intn(PH*2/3)
			b := .3 + r.Float64()*.7
			p.Pix[y*PW+x] = addc(p.Pix[y*PW+x], 0xffffff, b)
			if b > .9 && x+1 < PW {
				p.Pix[y*PW+x+1] = addc(p.Pix[y*PW+x+1], 0xffffff, .4)
			}
		}
	}
	for _, sx := range []float64{560, 560 + PW} {
		p.Glow(sx, 120, 160, th.Sun, .5)
		p.Circle(sx, 120, 40, th.Sun, 1)
		if th.Night {
			p.Circle(sx+12, 112, 34, mix(th.Sun, th.SkyT, .15), 1)
		}
	}
	far := [][3]float64{{34, 3, 0}, {20, 7, 1}, {9, 17, 2}, {4, 41, .3}}
	near := [][3]float64{{28, 2, .5}, {16, 9, 2}, {7, 23, 0}, {3, 57, 1}}
	for x := 0; x < PW; x++ {
		yf := int(ridge(float64(x), 225, far))
		for y := max(yf, 0); y < PH; y++ {
			t := float64(y-yf) / float64(PH-yf+1)
			p.Pix[y*PW+x] = mix(th.Far, th.Mist, .35*(1-t))
		}
	}
	for y := 260; y < PH; y++ {
		a := .45 * math.Sin(float64(y-260)/float64(PH-260)*math.Pi)
		for x := 0; x < PW; x++ {
			p.Pix[y*PW+x] = blend(p.Pix[y*PW+x], th.Mist, a)
		}
	}
	for x := 0; x < PW; x++ {
		yn := int(ridge(float64(x), 305, near))
		for y := max(yn, 0); y < PH; y++ {
			p.Pix[y*PW+x] = th.Near
		}
	}
	// pagodes et pins en silhouette
	for _, px := range []float64{300, 1220, 1730} {
		base := ridge(px, 305, near)
		for t := 0; t < 4; t++ {
			w := 46.0 - float64(t)*9
			y := base - float64(t)*22
			p.Poly([]float64{px - w, px + w, px + w*.55, px - w*.55}, []float64{y, y, y - 9, y - 9}, th.Near, 1)
			p.Rect(int(px-w*.45), int(y-22), int(w*.9), 14, th.Near)
		}
		p.Rect(int(px-2), int(base-120), 4, 30, th.Near)
	}
	for i := 0; i < 70; i++ {
		x := r.Float64() * PW
		y := ridge(x, 305, near) + 4
		h := 18 + r.Float64()*20
		p.Poly([]float64{x - h*.35, x + h*.35, x}, []float64{y, y, y - h}, shade(th.Near, .85), 1)
	}
	return p
}

func tileRow(f *FB, y0, rows int, base, tile, hi uint32, cap uint32) {
	f.Rect(0, y0, UW, rows*14+8, base)
	for r := 0; r < rows; r++ {
		y := float64(y0 + r*14)
		off := float64((r % 2) * 16)
		for x := -32.0; x < UW+32; x += 32 {
			cx := x + off + 16
			f.Capsule(cx, y+2, cx, y+12, 12, shade(tile, .6), 1)
			f.Capsule(cx, y+1, cx, y+11, 10.5, tile, 1)
			f.Capsule(cx-5, y+2, cx-5, y+9, 2.5, hi, .55)
		}
	}
	if cap != 0 {
		y := float64(y0 + rows*14 + 2)
		for x := 16.0; x < UW+32; x += 32 {
			f.Circle(x, y, 7, shade(cap, .55), 1)
			f.Circle(x, y, 5.5, cap, 1)
			f.Circle(x-1.5, y-1.5, 2, lighten(cap, .5), .7)
		}
	}
}

func pillar(f *FB, cx int) {
	for i := -24; i <= 24; i++ {
		t := float64(i) / 24
		k := .45 + .75*math.Cos(t*math.Pi/2)
		if i < -8 {
			k *= .9
		}
		f.Rect(cx+i, 264, 1, 480, shade(0xb81c18, k))
	}
	f.Capsule(float64(cx-10), 270, float64(cx-10), 740, 3, 0xff8a70, .35)
	for _, y := range []int{336, 680} {
		for i := -27; i <= 27; i++ {
			t := float64(i) / 27
			f.Rect(cx+i, y, 1, 14, shade(0xe8b040, .55+.6*math.Cos(t*math.Pi/2)))
		}
	}
	f.VGrad(cx-34, 714, 68, 30, 0xa8a090, 0x585048)
	f.Rect(cx-34, 714, 68, 3, 0xd8d0c0)
}

func keyPattern(f *FB, y0 int, period int, s int, col uint32, offset int) {
	// motif grec : cellule 16x8 logique
	cell := [][4]int{{0, 0, 14, 1}, {13, 0, 1, 6}, {6, 5, 8, 1}, {6, 2, 1, 4}, {6, 2, 4, 1}, {0, 0, 1, 8}, {0, 7, 16, 1}}
	for x := -period - offset; x < f.W+period; x += period {
		for _, c := range cell {
			f.Rect(x+c[0]*s, y0+c[1]*s+s/2, c[2]*s, c[3]*s, shade(col, .45))
			f.Rect(x+c[0]*s, y0+c[1]*s, c[2]*s, c[3]*s, col)
		}
	}
}

func buildUnit(th Theme) *FB {
	f := NewFB(UW, 960)
	// toit
	tileRow(f, 192, 3, 0x0e3014, 0x2f8a3a, 0x9ff0a0, 0xe8b040)
	f.VGrad(0, 248, UW, 13, 0xd8402a, 0x801810)
	f.VGrad(0, 260, UW, 4, 0xffe080, 0xa87018)
	// frise peinte
	f.VGrad(0, 264, UW, 72, 0x1c4058, 0x0c1e2c)
	f.Rect(0, 268, UW, 2, 0xe8b040)
	f.Rect(0, 330, UW, 2, 0xe8b040)
	for _, px := range []int{64, 320} {
		f.VGrad(px, 278, 128, 46, 0x2a6a70, 0x184850)
		f.Ring(float64(px+64), 301, 14, 17, 0xe8b040, 1)
		f.Circle(float64(px+64), 301, 8, 0xd83020, 1)
		for i := 0; i < 2; i++ {
			sx := float64(px + 20 + i*88)
			f.Ring(sx, 301, 6, 8, 0xe8b040, 1)
		}
		f.Rect(px, 278, 128, 2, 0xe8b040)
		f.Rect(px, 322, 128, 2, 0xe8b040)
	}
	for _, cx := range []int{0, UW} {
		for i, w := range []int{60, 48, 36} {
			y := 272 + i*20
			f.VGrad(cx-w/2, y, w, 16, 0xf0c050, 0x8a5a18)
		}
	}
	// mur
	r := rand.New(rand.NewSource(3))
	for y := 336; y < 744; y++ {
		base := mix(th.WallT, th.WallB, float64(y-336)/408)
		for x := 0; x < UW; x++ {
			n := (r.Float64() - .5) * .06
			f.Pix[y*UW+x] = shade(base, 1+n)
		}
	}
	f.VGrad(0, 664, UW, 80, 0x6a361e, 0x3a1a10)
	for x := 0; x < UW; x += 32 {
		f.Rect(x, 668, 2, 76, 0x2a120a)
		f.Rect(x+2, 668, 1, 76, 0x8a4a2a)
	}
	f.Rect(0, 664, UW, 3, 0xc88a50)
	// fenêtre lune
	cx, cy := 256.0, 500.0
	f.Circle(cx, cy, 146, shade(th.WallB, .8), .35)
	for y := int(cy - 114); y <= int(cy+114); y++ {
		for x := int(cx - 114); x <= int(cx+114); x++ {
			dx, dy := float64(x)+.5-cx, float64(y)+.5-cy
			if dx*dx+dy*dy < 114*114 {
				lx, ly := (x-int(cx)+1024)%38, (y-int(cy)+1024)%38
				if lx < 3 || ly < 3 {
					f.Pix[y*UW+x] = 0x2a140c
				} else {
					f.Pix[y*UW+x] = keyCol
				}
			}
		}
	}
	f.Ring(cx, cy, 112, 132, 0x3a1a10, 1)
	f.Ring(cx, cy, 118, 125, 0xe8b040, 1)
	f.Ring(cx, cy, 121, 123, 0xfff0a0, .6)
	// lanternes
	for _, lx := range []float64{104, 408} {
		f.Capsule(lx, 336, lx, 380, 1.2, 0x2a1a10, 1)
		f.Rect(int(lx)-16, 376, 32, 10, 0xc89030)
		f.Ellipse(lx, 424, 32, 42, 0x5a0808, 1)
		f.Ellipse(lx, 424, 29, 39, 0xd02818, 1)
		f.Ellipse(lx, 424, 18, 39, 0xe84a28, .6)
		f.Ellipse(lx-8, 412, 7, 18, 0xffb080, .5)
		for _, k := range []float64{-20, -8, 8, 20} {
			f.Capsule(lx+k, 392, lx+k*1.05, 456, 1.1, 0x8a1010, .7)
		}
		f.Rect(int(lx)-16, 460, 32, 10, 0xc89030)
		f.Capsule(lx, 470, lx, 506, 2.5, 0xe8b040, 1)
		f.Capsule(lx, 490, lx, 512, 4, 0xd02818, 1)
	}
	pillar(f, 0)
	pillar(f, UW)
	// plancher
	for row := 0; row < 4; row++ {
		y := 744 + row*18
		f.VGrad(0, y, UW, 18, 0xd08a44, 0x9a5a28)
		f.Rect(0, y+16, UW, 2, 0x4a2410)
		for x := (row * 96) % 128; x < UW; x += 128 {
			f.Rect(x, y, 2, 16, 0x5a2e16)
		}
		for x := 0; x < UW; x += 7 {
			f.Rect(x, y+4+(x*13)%9, 5, 1, 0xb87038)
		}
	}
	f.RectA(0, 744, UW, 10, 0xffffff, .12)
	// frise grecque
	f.VGrad(0, 816, UW, 48, 0xb82414, 0x780c06)
	f.Rect(0, 816, UW, 3, 0xffd070)
	f.Rect(0, 861, UW, 3, 0x3a0804)
	keyPattern(f, 824, 64, 4, 0xf8c860, 0)
	// tuiles du bas
	tileRow(f, 864, 7, 0x0a2810, 0x24703a, 0x80d090, 0)
	return f
}

func ensureBG() {
	lv := level
	if state == "title" || state == "options" || state == "scores" {
		lv = 1
	}
	if lv == bgLevel {
		return
	}
	bgLevel = lv
	pano = buildPano(themes[lv])
	bgUnit = buildUnit(themes[lv])
	amb = nil
}

func drawBgHD(f *FB, camPx float64) {
	ensureBG()
	f.Rect(0, 0, FW, 192, 0)
	c := int(math.Floor(camPx))
	po := int(math.Floor(camPx*.35)) % PW
	for y := 192; y < FH; y++ {
		srow := bgUnit.Pix[y*UW : y*UW+UW]
		drow := f.Pix[y*FW : y*FW+FW]
		ux := ((c % UW) + UW) % UW
		inPano := y >= PY0 && y < PY0+PH
		var prow []uint32
		if inPano {
			prow = pano.Pix[(y-PY0)*PW : (y-PY0)*PW+PW]
		}
		for x := 0; x < FW; x++ {
			v := srow[ux]
			if v == keyCol && inPano {
				v = prow[(x+po)%PW]
			}
			drow[x] = v
			ux++
			if ux == UW {
				ux = 0
			}
		}
	}
	th := themes[bgLevel]
	// halos des lanternes
	for wx := (c/UW)*UW - UW; wx < c+FW+UW; wx += UW {
		for _, lx := range []int{104, 408} {
			x := float64(wx + lx - c)
			if x < -200 || x > FW+200 {
				continue
			}
			fl := .28 + .05*math.Sin(float64(tick)*.13+float64(wx)*.01) + .03*math.Sin(float64(tick)*.37)
			f.Glow(x, 424, 150, th.Glow, fl)
			f.Glow(x, 424, 50, 0xfff0c0, .25)
		}
	}
	// portes aux extrémités du niveau
	for _, ex := range []float64{0, LVLEN} {
		x := ex*S - camPx
		if x < -300 || x > FW+300 {
			continue
		}
		dx := x
		if ex != 0 {
			dx = x - 104
		}
		f.VGrad(int(dx), 420, 104, 324, 0x140a18, 0x302040)
		for i := 0; i < 6; i++ {
			hg := (i + 1) * 46
			s2 := int(dx) + i*16 + 4
			if ex == 0 {
				s2 = int(dx) + (5-i)*16 + 4
			}
			f.VGrad(s2, 744-hg, 16, hg, 0x6a5088, 0x302040)
			f.Rect(s2, 744-hg, 16, 3, 0xb8a0d8)
		}
		f.VGrad(int(dx)-14, 396, 132, 24, 0xe03a20, 0x8a1810)
		f.Rect(int(dx)-14, 396, 132, 4, 0xffd070)
		f.Rect(int(dx)-14, 420, 14, 324, 0xa81c14)
		f.Rect(int(dx)+104, 420, 14, 324, 0xa81c14)
	}
	// particules d'ambiance
	if len(amb) < 26 && tick%6 == 0 {
		l := 300 + rand.Intn(200)
		p := &Part{X: rand.Float64() * FW, H: 200 + rand.Float64()*500, Life: l, Max: l}
		switch th.Amb {
		case 0:
			p.VX, p.VH, p.C, p.Size = .15, -.05, 0xfff0d0, 1.5
		case 1:
			p.VX, p.VH, p.C, p.Size, p.Glow = (rand.Float64()-.5)*.4, (rand.Float64()-.5)*.3, 0xd0ff80, 2, true
		case 2:
			p.VX, p.VH, p.C, p.Size, p.H = .6+rand.Float64()*.5, .8, 0xffb8d0, 3, 180
		case 3:
			p.VX, p.VH, p.C, p.Size, p.Glow, p.H = (rand.Float64()-.5)*.5, -.9, 0xff8030, 1.8, true, 760
		}
		amb = append(amb, p)
	}
	for _, p := range amb {
		p.Life--
		p.X += p.VX + math.Sin(float64(p.Life)*.05)*.3
		p.H += p.VH
		a := math.Min(1, math.Min(float64(p.Life), float64(p.Max-p.Life))/40)
		if p.Glow {
			f.Glow(p.X, p.H, p.Size*6, p.C, .5*a*(.6+.4*math.Sin(float64(p.Life)*.2)))
		} else {
			f.Ellipse(p.X, p.H, p.Size, p.Size*.7, p.C, .7*a)
		}
	}
	amb = filter(amb, func(p *Part) bool { return p.Life > 0 && p.X > -20 && p.X < FW+20 && p.H > 180 && p.H < 800 })
}

/* ---------- Monde HD ---------- */

func hdShadow(f *FB, x, h, w float64) {
	k := math.Max(.25, 1-h/70)
	f.Ellipse(x, GY*S+4, w*S*k, 2.4*S*k, 0x000000, .32*k)
}

func walkPh(e *Ent) float64 { return float64(tick)*.2 + e.X*.05 }

func drawEntHD(f *FB, e *Ent, camPx float64) {
	sx := e.X*S - camPx
	sy := (GY - e.Y) * S
	if sx < -200 || sx > FW+200 {
		return
	}
	if !e.Dying {
		hdShadow(f, sx, e.Y, 7)
	}
	switch e.Type {
	case "gripper", "knifer":
		name := "walk"
		k := .3
		if e.Dying || e.St == "stagger" {
			name = "hurt"
		} else if e.Type == "gripper" && e.St == "grab" {
			name = "grab"
		} else if e.St == "throw" {
			k = .45
			if e.High {
				name = map[bool]string{true: "throwUp", false: "throwFwd"}[e.T < 14]
			} else {
				name = map[bool]string{true: "crouch", false: "lowpunch"}[e.T < 14]
			}
		}
		t, g := poseTarget(name, walkPh(e))
		cur := animStep(&e.A, t, k)
		j := drawFigure(f, sx, sy, e.Face, cur, g && !e.Dying, e.A.Rot, pals[e.Type], FigOpt{Scale: 1})
		if e.Type == "knifer" && e.St == "throw" && e.T < 14 {
			drawKnife(f, j.FHand[0], j.FHand[1]-6, e.Face)
		}
	case "tomtom":
		name := "walk"
		if e.Y > 0 || e.Dying {
			name = "jump"
		}
		t, g := poseTarget(name, float64(tick)*.35)
		cur := animStep(&e.A, t, .35)
		drawFigure(f, sx, sy, e.Face, cur, g && e.Y <= 0 && !e.Dying, e.A.Rot, pals["tomtom"], FigOpt{Scale: .58, HeadK: 1.55, Bulk: 1.2})
	case "snake":
		fc := e.Face
		for i := 9; i >= 0; i-- {
			x := sx - fc*float64(i)*7
			y := GY*S - 6 - math.Sin(float64(tick)*.3-float64(i)*.8)*4
			r := 7 - float64(i)*.35
			f.Circle(x, y, r+2, ink, 1)
			c := uint32(0x58c040)
			if i%2 == 1 {
				c = 0x2e8a30
			}
			f.Circle(x, y, r, c, 1)
			f.Circle(x-2, y-3, r*.35, 0xc8ff90, .5)
		}
		hy := GY*S - 10 - math.Sin(float64(tick)*.3)*4
		f.Circle(sx+fc*6, hy, 9, ink, 1)
		f.Circle(sx+fc*6, hy, 7.5, 0x68d048, 1)
		f.Circle(sx+fc*9, hy-3, 2, 0xffe040, 1)
		f.Circle(sx+fc*9.5, hy-3, 1, 0, 1)
		if (tick/6)%3 != 0 {
			f.Capsule(sx+fc*13, hy+2, sx+fc*20, hy+3, 1, 0xe02020, 1)
		}
	case "boss":
		drawBossHD(f, e, sx, sy)
	}
}

func drawKnife(f *FB, x, y, face float64) {
	f.Capsule(x-face*4, y, x+face*14, y, 3, ink, 1)
	f.Capsule(x+face*2, y, x+face*14, y, 2, 0xe8eef8, 1)
	f.Capsule(x+face*3, y-1, x+face*13, y-1, .8, 0xffffff, .8)
	f.Capsule(x-face*4, y, x, y, 2.2, 0x6a3418, 1)
}

func drawBossHD(f *FB, b *Ent, sx, sy float64) {
	s := bosses[b.Kind]
	name := "stand"
	k := .3
	switch b.St {
	case "approach", "back":
		name = "walk"
	case "wind":
		k = .35
		if b.Kind == "maitre" {
			name = map[bool]string{true: "kickprep", false: "crouch"}[b.High]
		} else {
			name = map[bool]string{true: "throwUp", false: "crouch"}[b.High]
		}
	case "active":
		k = .6
		if b.Kind == "maitre" {
			name = map[bool]string{true: "kick", false: "lowkick"}[b.High]
		} else {
			name = map[bool]string{true: "punch", false: "lowpunch"}[b.High]
		}
	case "hurt", "dead":
		name = "hurt"
	case "throw":
		name = map[bool]string{true: "throwUp", false: "throwFwd"}[b.T < 16]
	case "jk":
		name = "airkick"
	case "cast":
		name = map[bool]string{true: "throwUp", false: "cast"}[b.T < 22]
	}
	alpha := 1.0
	if b.St == "vanish" {
		alpha = math.Max(0, 1-float64(b.T)/24)
	} else if b.St == "appear" {
		alpha = math.Min(1, float64(b.T)/20)
	}
	if !b.Dying {
		hdShadow(f, sx, b.Y, 7*s.Scale)
	}
	if b.Kind == "sorcier" {
		f.Glow(sx, sy-60*s.Scale, 120, 0x7040c0, .25*alpha)
	}
	t, g := poseTarget(name, walkPh(b)*.8)
	cur := animStep(&b.A, t, k)
	o := FigOpt{Scale: s.Scale, Alpha: alpha, Flash: b.Flash > 0 && (b.Flash>>1)%2 == 1, Robe: b.Kind == "sorcier"}
	if b.Kind == "colosse" {
		o.Bulk = 1.25
	}
	j := drawFigure(f, sx, sy, b.Face, cur, g && !b.Dying, b.A.Rot, pals[b.Kind], o)
	fc := b.Face
	switch b.Kind {
	case "baton":
		hx, hy := j.FHand[0], j.FHand[1]
		var x1, y1, x2, y2 float64
		switch {
		case b.St == "active":
			x1, y1, x2, y2 = hx-fc*20, hy, hx+fc*118, hy
		case b.St == "wind" && b.High:
			x1, y1, x2, y2 = hx-fc*30, hy+40, hx+fc*10, hy-60
		case b.St == "wind":
			x1, y1, x2, y2 = hx-fc*80, hy+18, hx+fc*30, hy-4
		default:
			x1, y1, x2, y2 = hx+fc*4, hy+60, hx-fc*2, hy-70
		}
		f.Capsule(x1, y1, x2, y2, 4.5, ink, 1)
		f.Capsule(x1, y1, x2, y2, 3, 0xa86a34, 1)
		f.Capsule(x1, y1-1, x2, y2-1, 1, 0xf0c080, .6)
		if b.St == "active" {
			f.Capsule(x1, y1, x2, y2, 10, 0xffffff, .12)
		}
	case "boomerang":
		if !b.HasBoom && b.St != "throw" {
			drawBoomer(f, j.FHand[0], j.FHand[1]-4, 0)
		}
	case "sorcier":
		if b.St == "cast" && b.T < 22 {
			f.Glow(j.FHand[0], j.FHand[1], 40, 0xffa040, .6+.3*math.Sin(float64(tick)*.5))
		}
	case "colosse":
		if b.St == "wind" {
			f.Glow(j.FHand[0], j.FHand[1], 30, 0xff4020, .3)
		}
	}
}

func drawBoomer(f *FB, x, y float64, spin int) {
	a := float64(spin) * .45
	for _, d := range []float64{0, math.Pi / 2} {
		ex, ey := x+math.Cos(a+d)*13, y+math.Sin(a+d)*13
		f.Capsule(x, y, ex, ey, 4.5, ink, 1)
		f.Capsule(x, y, ex, ey, 3.2, 0xe8b838, 1)
		f.Capsule(x, y-1, ex, ey-1, 1, 0xfff0a0, .7)
	}
}

func drawShotHD(f *FB, s *Shot, camPx float64) {
	x := s.X*S - camPx
	y := (GY - s.H) * S
	switch s.K {
	case "bullet":
		f.Capsule(x-sgn(s.VX)*24, y-4, x, y-4, 2, 0xffd060, .35)
		f.Capsule(x-sgn(s.VX)*6, y-4, x, y-4, 2.6, 0xfff8c0, 1)
	case "knife":
		drawKnife(f, x-sgn(s.VX)*6, y-4, sgn(s.VX))
		f.Capsule(x-sgn(s.VX)*30, y-4, x-sgn(s.VX)*10, y-4, 1.5, 0xffffff, .25)
	case "fire":
		f.Glow(x, y-10, 60, 0xff6010, .7)
		fl := math.Sin(float64(tick) * .6)
		f.Circle(x, y-10, 12+fl*1.5, 0xff8020, 1)
		f.Circle(x+sgn(s.VX)*2, y-11, 8, 0xffd040, 1)
		f.Circle(x+sgn(s.VX)*3, y-12, 4, 0xfffff0, 1)
	case "vase":
		hdShadow(f, x, s.H, 4)
		f.Ellipse(x, y-14, 18, 18, ink, 1)
		f.Ellipse(x, y-14, 16, 16, 0xc87830, 1)
		f.Rect(int(x)-8, int(y)-40, 16, 12, ink)
		f.Rect(int(x)-6, int(y)-38, 12, 10, 0xc87830)
		f.Rect(int(x)-16, int(y)-18, 32, 4, 0x2a5aa0)
		f.Ellipse(x-6, y-20, 4, 7, 0xffe0b0, .5)
	case "boomer":
		drawBoomer(f, x, y-8, s.Spin)
		f.Ring(x, y-8, 14, 18, 0xfff0a0, .2)
	}
}

func renderHD(f *FB, camX float64) {
	sh := 0.0
	if shake > 0 {
		sh = (rand.Float64() - .5) * float64(shake) * 1.5
	}
	camPx := camX*S + sh
	drawBgHD(f, camPx)
	if state == "title" || state == "options" || state == "scores" {
		return
	}
	for _, e := range ents {
		if e.Type != "boss" {
			drawEntHD(f, e, camPx)
		}
	}
	if boss != nil {
		drawEntHD(f, boss, camPx)
	}
	// joueur
	sx, sy := P.X*S-camPx, (GY-P.Y)*S
	hdShadow(f, sx, P.Y, 7)
	if playerVisible() {
		name := playerPose()
		k := .32
		ph := float64(P.Walk) * .2
		if P.Atk != nil {
			k = .6
		}
		switch name {
		case "walkA", "walkB":
			name = "walk"
		case "lying":
			k = .15
		}
		t, g := poseTarget(name, ph)
		cur := animStep(&P.A, t, k)
		jx := 0.0
		if isGrabbed() && state == "play" {
			jx = float64((tick>>1)%2)*6 - 3
		}
		j := drawFigure(f, sx+jx, sy, P.Face, cur, g, 0, pals["hero"], FigOpt{Scale: 1})
		if atkActive() {
			a := P.Atk.Type
			var tx, ty float64
			if a == "punch" || a == "lowpunch" {
				tx, ty = j.FHand[0], j.FHand[1]
			} else {
				tx, ty = j.FFoot[0], j.FFoot[1]
			}
			f.Glow(tx, ty, 26, 0xffffff, .3)
		}
		if weapon != "" && (name == "punch" || name == "lowpunch") {
			hx, hy, fc := j.FHand[0], j.FHand[1], P.Face
			l := 26.0
			if weapon == "riot" {
				l = 42
			}
			f.Capsule(hx-fc*4, hy-4, hx+fc*l, hy-4, 6, ink, 1)
			f.Capsule(hx-fc*2, hy-4, hx+fc*l, hy-4, 4.2, 0x3a3a44, 1)
			f.Capsule(hx, hy-6, hx+fc*l, hy-6, 1.2, 0x9090a0, .8)
			f.Capsule(hx, hy, hx-fc*2, hy+12, 4, 0x2a2a30, 1)
			if weapon == "riot" {
				f.Capsule(hx+fc*14, hy, hx+fc*16, hy+16, 4, 0x202024, 1)
			}
			if P.ShootT > 4 {
				f.Glow(hx+fc*(l+8), hy-4, 34, 0xffd060, .9)
				f.Circle(hx+fc*(l+6), hy-4, 6, 0xfffff0, 1)
			}
		}
	}
	for _, s := range shots {
		drawShotHD(f, s, camPx)
	}
	for _, p := range parts {
		x, y := p.X*S-camPx, (GY-p.H)*S
		a := float64(p.Life) / float64(p.Max)
		if p.Size >= 5 {
			r := 30 + (1-a)*30
			f.Glow(x, y, r*1.6, 0xfff0a0, .8*a)
			for i := 0; i < 4; i++ {
				an := float64(i)*math.Pi/4 + float64(tick)*.1
				f.Capsule(x-math.Cos(an)*r, y-math.Sin(an)*r, x+math.Cos(an)*r, y+math.Sin(an)*r, 2.5*a, 0xffffff, a)
			}
			continue
		}
		if p.Glow {
			f.Glow(x, y, p.Size*10, p.C, .7*a)
			f.Circle(x, y, p.Size*1.6, lighten(p.C, .5), a)
		} else {
			f.Circle(x, y, p.Size*S*.8, p.C, a)
		}
	}
}
