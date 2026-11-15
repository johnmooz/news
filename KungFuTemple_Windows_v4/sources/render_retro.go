package main

import "math"

// Rendu « rétro NES » : rectangles pixel à 256x240.

type rctx struct {
	f          *FB
	x, y, face float64
	s          float64
}

var rc rctx

func (c *rctx) r(lx, ly, w, h float64, col uint32) {
	s := c.s
	X := c.x + lx*s
	if c.face < 0 {
		X = c.x - (lx+w)*s
	}
	c.f.Rect(int(math.Round(X)), int(math.Round(c.y+ly*s)), max(1, int(math.Round(w*s))), max(1, int(math.Round(h*s))), col)
}

var retroPose = map[string]struct {
	legs   string
	dy, dx float64
	arms   string
}{
	"stand": {"stand", 0, 0, "guard"}, "walkA": {"walkA", 0, 0, "guard"}, "walkB": {"walkB", 0, 0, "guard"},
	"kickprep": {"kickprep", 0, -1, "guard"}, "kick": {"kick", 0, -2, "guard"}, "punch": {"stand", 0, 1, "punch"},
	"crouch": {"crouch", 10, 0, "guard"}, "lowkick": {"lowkick", 10, -2, "guard"}, "lowpunch": {"crouch", 10, 1, "punch"},
	"jump": {"tuck", 0, 0, "guard"}, "airkick": {"airkick", 0, -2, "guard"}, "hurt": {"stand", 0, -2, "flail"},
	"grab": {"walkA", 0, 1, "grab"}, "throwUp": {"walkA", 0, -1, "up"}, "throwFwd": {"walkA", 0, 1, "fwd"}, "cast": {"stand", 0, 0, "grab"},
}

func retroLegs(t string, p Pal) {
	r := rc.r
	P_, S := p.Pants, p.Shoe
	switch t {
	case "stand":
		r(-4, -13, 3, 12, P_)
		r(1, -13, 3, 12, P_)
		r(-5, -1, 4, 1, S)
		r(1, -1, 4, 1, S)
	case "walkA":
		r(-6, -13, 3, 12, P_)
		r(3, -13, 3, 12, P_)
		r(-4, -14, 8, 3, P_)
		r(-7, -1, 4, 1, S)
		r(3, -1, 4, 1, S)
	case "walkB":
		r(-3, -13, 3, 12, P_)
		r(0, -13, 3, 12, P_)
		r(-4, -1, 4, 1, S)
		r(0, -1, 4, 1, S)
	case "kickprep":
		r(-3, -13, 3, 12, P_)
		r(-4, -1, 4, 1, S)
		r(0, -17, 6, 3, P_)
		r(4, -16, 3, 7, P_)
		r(4, -9, 4, 1, S)
	case "kick":
		r(-3, -13, 3, 12, P_)
		r(-4, -1, 4, 1, S)
		r(0, -16, 5, 3, P_)
		r(4, -18, 5, 3, P_)
		r(8, -20, 6, 3, P_)
		r(14, -21, 3, 3, S)
	case "crouch":
		r(-6, -7, 4, 6, P_)
		r(-7, -1, 5, 1, S)
		r(-2, -8, 8, 3, P_)
		r(3, -6, 3, 5, P_)
		r(3, -1, 4, 1, S)
	case "lowkick":
		r(-6, -7, 4, 6, P_)
		r(-7, -1, 5, 1, S)
		r(-2, -6, 14, 3, P_)
		r(12, -6, 3, 3, S)
	case "tuck":
		r(-4, -15, 8, 3, P_)
		r(-5, -13, 3, 6, P_)
		r(2, -13, 3, 6, P_)
		r(-6, -8, 4, 1, S)
		r(2, -8, 4, 1, S)
	case "airkick":
		r(-4, -15, 6, 3, P_)
		r(-5, -13, 3, 6, P_)
		r(-6, -8, 4, 1, S)
		r(0, -16, 14, 3, P_)
		r(14, -17, 3, 3, S)
	}
}

func retroUpper(dy, dx float64, arms string, p Pal) {
	o := func(lx, ly, w, h float64, c uint32) { rc.r(lx+dx, ly+dy, w, h, c) }
	if arms == "flail" {
		o(-7, -29, 2, 7, p.Top)
		o(-7, -31, 2, 2, p.Skin)
	} else {
		o(-6, -23, 2, 7, p.Top)
		o(-6, -16, 2, 2, p.Skin)
	}
	o(-4, -15, 8, 3, p.Pants)
	o(-4, -24, 8, 9, p.Top)
	o(-4, -16, 8, 2, p.Belt)
	lap := p.Belt
	if lap == 0x181818 {
		lap = 0xd8d8d8
	}
	o(-1, -24, 1, 5, lap)
	switch arms {
	case "guard":
		o(2, -23, 3, 4, p.Top)
		o(4, -26, 2, 4, p.Skin)
	case "punch":
		o(2, -23, 5, 3, p.Top)
		o(7, -23, 6, 2, p.Skin)
		o(13, -24, 2, 3, p.Skin)
	case "grab":
		o(2, -24, 8, 2, p.Top)
		o(10, -25, 2, 3, p.Skin)
		o(2, -20, 7, 2, p.Top)
		o(9, -21, 2, 3, p.Skin)
	case "up":
		o(1, -29, 2, 6, p.Top)
		o(1, -32, 2, 3, p.Skin)
	case "fwd":
		o(2, -24, 9, 2, p.Top)
		o(11, -25, 2, 3, p.Skin)
	case "flail":
		o(4, -29, 2, 7, p.Top)
		o(4, -31, 2, 2, p.Skin)
	}
	o(-1, -25, 2, 1, p.Skin)
	o(-3, -31, 6, 6, p.Skin)
	if p.Bald {
		o(-2, -32, 4, 1, p.Skin)
	} else {
		o(-3, -32, 6, 2, p.Hair)
		o(-3, -31, 2, 4, p.Hair)
	}
	o(1, -29, 1, 1, 0x101010)
	if p.HasBeard {
		o(0, -27, 3, 3, p.Beard)
	}
	if p.HasBand {
		o(-3, -30, 6, 1, p.Band)
		o(-6, -30, 3, 1, p.Band)
		o(-7, -29, 2, 1, p.Band)
	}
}

func retroHuman(x, fy, face float64, pose string, p Pal, s float64) {
	rc.x, rc.y, rc.face, rc.s = math.Round(x), math.Round(fy), face, s
	r := rc.r
	if pose == "lying" {
		r(-15, -4, 11, 4, p.Pants)
		r(-17, -4, 2, 4, p.Shoe)
		r(-4, -5, 9, 5, p.Top)
		r(-1, -5, 2, 5, p.Belt)
		r(5, -6, 6, 6, p.Skin)
		r(9, -6, 2, 6, p.Hair)
		return
	}
	d, ok := retroPose[pose]
	if !ok {
		d = retroPose["stand"]
	}
	retroLegs(d.legs, p)
	retroUpper(d.dy, d.dx, d.arms, p)
}

func retroTiles(f *FB, y, h int, c int) {
	f.Rect(0, y, 256, h, 0x1c6020)
	off := ((c % 8) + 8) % 8
	for row := 0; row*5 < h; row++ {
		yy := y + row*5
		o := (row % 2) * 4
		for x := -off - 8 + o; x < 264; x += 8 {
			f.Rect(x+1, yy+1, 6, 1, 0x88e070)
			f.Rect(x, yy+2, 1, 2, 0x88e070)
			f.Rect(x+7, yy+2, 1, 2, 0x88e070)
			f.Rect(x+2, yy+2, 4, 2, 0x3c9838)
		}
	}
}

func retroBG(f *FB, camX float64) {
	c := int(math.Round(camX))
	g := int(GY)
	retroTiles(f, 48, 13, c)
	f.Rect(0, 60, 256, 3, 0xc83018)
	f.Rect(0, 63, 256, 1, 0xf8b048)
	f.Rect(0, 64, 256, g-68, 0x8890f0)
	f.Rect(0, 80, 256, 1, 0xf8b048)
	f.Rect(0, 81, 256, 3, 0xc83018)
	for wx := (c/128)*128 - 128; wx < c+256+128; wx += 128 {
		sx := wx - c + 64
		f.Rect(sx-3, 64, 6, 16, 0xc83018)
		f.Rect(sx-12, 66, 9, 9, 0xc83018)
		f.Rect(sx+3, 66, 9, 9, 0xc83018)
		f.Rect(sx-10, 68, 5, 1, 0xf8b048)
		f.Rect(sx-10, 68, 1, 5, 0xf8b048)
		f.Rect(sx-10, 72, 3, 1, 0xf8b048)
		f.Rect(sx+5, 68, 5, 1, 0xf8b048)
		f.Rect(sx+9, 68, 1, 5, 0xf8b048)
		f.Rect(sx+7, 72, 3, 1, 0xf8b048)
		lx := sx + 64
		f.Rect(lx, 84, 1, 10, 0x402020)
		f.Rect(lx-3, 94, 7, 10, 0xd83020)
		f.Rect(lx-3, 96, 7, 1, 0xf8d040)
		f.Rect(lx-3, 101, 7, 1, 0xf8d040)
		f.Rect(lx-1, 104, 3, 3, 0xf8d040)
	}
	for _, ex := range []int{0, int(LVLEN)} {
		sx := ex - c
		if sx < -40 || sx > 296 {
			continue
		}
		dx := sx
		if ex != 0 {
			dx = sx - 22
		}
		f.Rect(dx, 108, 22, g-108, 0x302048)
		f.Rect(dx-2, 104, 26, 4, 0xc83018)
		for i := 0; i < 5; i++ {
			hg := (i + 1) * 14
			s2 := dx + i*4 + 1
			if ex == 0 {
				s2 = dx + (4-i)*4 + 1
			}
			f.Rect(s2, g-hg, 4, hg, 0x604878)
			f.Rect(s2, g-hg, 4, 1, 0xa088c0)
		}
	}
	f.Rect(0, g-4, 256, 18, 0xf0a040)
	for _, y := range []int{g - 1, g + 3, g + 7, g + 11} {
		f.Rect(0, y, 256, 1, 0xb05818)
	}
	f.Rect(0, g+14, 256, 12, 0xc83018)
	ko := ((c % 12) + 12) % 12
	ky := g + 16
	K := uint32(0xf8c870)
	for x := -ko - 12; x < 268; x += 12 {
		f.Rect(x, ky, 10, 1, K)
		f.Rect(x+9, ky, 1, 6, K)
		f.Rect(x+4, ky+5, 6, 1, K)
		f.Rect(x+4, ky+2, 1, 4, K)
		f.Rect(x+4, ky+2, 3, 1, K)
		f.Rect(x, ky, 1, 8, K)
		f.Rect(x, ky+7, 12, 1, K)
	}
	retroTiles(f, g+26, 240-(g+26), c)
}

func walkPose() string {
	if (tick>>3)%2 == 1 {
		return "walkA"
	}
	return "walkB"
}

func retroTom(f *FB, e *Ent, camX float64) {
	rc.f, rc.x, rc.y, rc.face, rc.s = f, math.Round(e.X-camX), math.Round(GY-e.Y), e.Face, 1
	r := rc.r
	red, sk, wh := uint32(0xd83030), uint32(0xf0b070), uint32(0xf0e8d0)
	if e.Y > 0 || e.Dying {
		r(-5, -12, 10, 9, red)
		r(-3, -15, 6, 5, sk)
		r(-3, -16, 6, 2, 0x181818)
		r(-5, -4, 4, 3, wh)
		r(1, -4, 4, 3, wh)
		r(4, -11, 3, 2, sk)
		return
	}
	w := float64((tick >> 3) % 2)
	r(-4+w, -5, 3, 5, wh)
	r(1-w, -5, 3, 5, wh)
	r(-5, -1, 4, 1, 0x181818)
	r(1, -1, 4, 1, 0x181818)
	r(-5, -11, 10, 7, red)
	r(-5, -6, 10, 1, 0x181818)
	r(-3, -16, 6, 5, sk)
	r(-3, -17, 6, 2, 0x181818)
	r(1, -14, 1, 1, 0)
	r(4, -10, 3, 2, sk)
}

func retroBoss(f *FB, b *Ent, camX float64) {
	s := bosses[b.Kind]
	pose := "stand"
	switch b.St {
	case "approach", "back":
		pose = walkPose()
	case "wind":
		if b.Kind == "maitre" {
			pose = map[bool]string{true: "kickprep", false: "crouch"}[b.High]
		} else {
			pose = map[bool]string{true: "throwUp", false: "crouch"}[b.High]
		}
	case "active":
		if b.Kind == "maitre" {
			pose = map[bool]string{true: "kick", false: "lowkick"}[b.High]
		} else {
			pose = map[bool]string{true: "punch", false: "lowpunch"}[b.High]
		}
	case "hurt", "dead":
		pose = "hurt"
	case "throw":
		pose = map[bool]string{true: "throwUp", false: "throwFwd"}[b.T < 16]
	case "jk":
		pose = "airkick"
	case "cast":
		pose = map[bool]string{true: "throwUp", false: "cast"}[b.T < 22]
	case "vanish", "appear":
		if (b.T>>1)%2 == 1 {
			return
		}
	}
	if b.Flash > 0 && (b.Flash>>1)%2 == 1 {
		return
	}
	retroHuman(b.X-camX, GY-b.Y, b.Face, pose, pals[b.Kind], s.Scale)
	r := rc.r
	st := uint32(0xa06030)
	switch b.Kind {
	case "baton":
		switch {
		case b.St == "active":
			yy := -10.0
			if b.High {
				yy = -23
			}
			r(4, yy, 34, 2, st)
		case b.St == "wind" && b.High:
			r(0, -46, 2, 22, st)
		case b.St == "wind":
			r(-16, -6, 18, 2, st)
		default:
			r(5, -34, 2, 32, st)
		}
	case "boomerang":
		if !b.HasBoom && b.St != "throw" {
			r(5, -28, 2, 4, 0xe0c040)
			r(5, -28, 4, 2, 0xe0c040)
		}
	case "sorcier":
		if pose != "crouch" && pose != "lowpunch" && pose != "hurt" {
			r(-5, -15, 10, 13, pals["sorcier"].Top)
			r(-5, -6, 10, 1, 0xc03030)
		}
		if b.St == "cast" && b.T < 22 && (tick>>1)%2 == 1 {
			r(0, -35, 3, 3, 0xf8d040)
		}
	}
}

func retroEnt(f *FB, e *Ent, camX float64) {
	sx, fy := e.X-camX, GY-e.Y
	switch e.Type {
	case "gripper":
		pose := walkPose()
		if e.Dying {
			pose = "hurt"
		} else if e.St == "grab" {
			pose = "grab"
		}
		retroHuman(sx, fy, e.Face, pose, pals["gripper"], 1)
	case "knifer":
		pose := walkPose()
		if e.Dying || e.St == "stagger" {
			pose = "hurt"
		} else if e.St == "throw" {
			if e.High {
				pose = map[bool]string{true: "throwUp", false: "throwFwd"}[e.T < 14]
			} else {
				pose = map[bool]string{true: "crouch", false: "lowpunch"}[e.T < 14]
			}
		}
		retroHuman(sx, fy, e.Face, pose, pals["knifer"], 1)
		if e.St == "throw" && e.T < 14 {
			if e.High {
				rc.r(0, -36, 2, 4, 0xe0e0e0)
			} else {
				rc.r(6, -14, 4, 1, 0xe0e0e0)
			}
		}
	case "tomtom":
		retroTom(f, e, camX)
	case "snake":
		X := int(math.Round(sx))
		w := (tick >> 2) % 2
		fc := int(e.Face)
		for i := 0; i < 5; i++ {
			c := uint32(0x70d050)
			if i%2 == 1 {
				c = 0x38a038
			}
			o := 0
			if fc > 0 {
				o = 2
			}
			f.Rect(X-fc*i*3-o, int(GY)-2-((i+w)%2), 3, 2, c)
		}
		if fc > 0 {
			f.Rect(X+1, int(GY)-4, 3, 3, 0x70d050)
			f.Rect(X+4, int(GY)-3, 2, 1, 0xe02020)
		} else {
			f.Rect(X-4, int(GY)-4, 3, 3, 0x70d050)
			f.Rect(X-6, int(GY)-3, 2, 1, 0xe02020)
		}
	case "boss":
		retroBoss(f, e, camX)
	}
}

func retroShot(f *FB, s *Shot, camX float64) {
	X, Y := int(math.Round(s.X-camX)), int(math.Round(GY-s.H))
	switch s.K {
	case "bullet":
		f.Rect(X-3, Y-1, 5, 2, 0xfff4a0)
	case "knife":
		f.Rect(X-4, Y-1, 6, 1, 0xe8e8e8)
		if s.VX > 0 {
			f.Rect(X-6, Y-2, 2, 3, 0x804020)
		} else {
			f.Rect(X+2, Y-2, 2, 3, 0x804020)
		}
	case "fire":
		fl := (tick >> 1) % 2
		c := uint32(0xf8d040)
		if fl == 1 {
			c = 0xf8a020
		}
		f.Rect(X-3, Y-5, 6, 5, c)
		f.Rect(X-2, Y-4, 3, 3, 0xfff8c0)
	case "vase":
		f.Rect(X-4, Y-8, 8, 7, 0xc87830)
		f.Rect(X-2, Y-10, 4, 2, 0xc87830)
		f.Rect(X-4, Y-5, 8, 1, 0xf0d070)
		f.Rect(X-3, Y-1, 6, 1, 0x804818)
	case "boomer":
		if (s.Spin>>2)%2 == 1 {
			f.Rect(X-4, Y-4, 8, 2, 0xe0c040)
			f.Rect(X-4, Y-4, 2, 6, 0xe0c040)
		} else {
			f.Rect(X-4, Y-2, 8, 2, 0xe0c040)
			f.Rect(X+2, Y-6, 2, 6, 0xe0c040)
		}
	}
}

func playerPose() string {
	switch {
	case state == "dying":
		if stT < 20 {
			return "hurt"
		}
		return "lying"
	case isGrabbed():
		if P.ShakeT > 0 {
			return "hurt"
		}
		return "stand"
	case P.Atk != nil:
		a := P.Atk
		switch a.Type {
		case "kick":
			if a.T < 5 {
				return "kickprep"
			}
			return "kick"
		case "punch":
			if a.T < 3 {
				return "stand"
			}
			return "punch"
		case "airkick":
			if a.T < 4 {
				return "jump"
			}
			return "airkick"
		}
		return a.Type
	case P.ShootT > 0 && weapon != "":
		if P.Crouch && P.Y <= 0 {
			return "lowpunch"
		}
		return "punch"
	case P.Hurt > 0:
		return "hurt"
	case P.Y > 0:
		return "jump"
	case P.Crouch:
		return "crouch"
	case P.Walk > 0:
		if (P.Walk>>3)%2 == 1 {
			return "walkA"
		}
		return "walkB"
	}
	return "stand"
}

func playerVisible() bool { return !(state == "play" && P.Inv > 0 && (P.Inv>>2)%2 == 1) }

func renderRetro(f *FB, camX float64) {
	rc.f = f
	f.Clear(0)
	retroBG(f, camX)
	if state == "title" || state == "options" || state == "scores" {
		return
	}
	for _, e := range ents {
		if e.Type != "boss" {
			retroEnt(f, e, camX)
		}
	}
	if boss != nil {
		retroEnt(f, boss, camX)
	}
	if playerVisible() {
		jx := 0.0
		if isGrabbed() && state == "play" {
			jx = float64((tick>>1)%2)*2 - 1
		}
		pose := playerPose()
		retroHuman(P.X-camX+jx, GY-P.Y, P.Face, pose, pals["hero"], 1)
		if weapon != "" && (pose == "punch" || pose == "lowpunch") {
			dy := 0.0
			if pose == "lowpunch" {
				dy = 10
			}
			g := func(a, b, c, d float64, col uint32) { rc.r(a+1, b+dy, c, d, col) }
			if weapon == "riot" {
				g(11, -26, 11, 3, 0x404048)
				g(14, -23, 2, 4, 0x202020)
			} else {
				g(12, -26, 6, 2, 0x303038)
				g(13, -24, 2, 3, 0x303038)
			}
			if P.ShootT > 4 {
				x := 18.0
				if weapon == "riot" {
					x = 22
				}
				g(x, -27, 3, 3, 0xfff080)
			}
		}
	}
	for _, s := range shots {
		retroShot(f, s, camX)
	}
	for _, p := range parts {
		X, Y := int(math.Round(p.X-camX)), int(math.Round(GY-p.H))
		if p.Size >= 5 {
			f.Rect(X-4, Y, 9, 1, 0xffffff)
			f.Rect(X, Y-4, 1, 9, 0xffffff)
			f.Rect(X-2, Y-2, 5, 5, 0xf8f070)
			continue
		}
		sz := 1
		if p.Size > 1.2 {
			sz = 2
		}
		f.Rect(X, Y, sz, sz, p.C)
	}
}
