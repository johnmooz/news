package main

import (
	"fmt"
	"math"
)

const (
	FW = 1024
	FH = 960
)

// UI dessinée en coordonnées logiques (256x240), en rétro (k=1) ou HD (k=4).
type UI struct {
	f  *FB
	k  float64
	hd bool
}

func (u UI) fs(sz int) int {
	if u.hd {
		return sz * 3
	}
	return sz
}

func (u UI) T(s string, x, y float64, sz int, c uint32, al int) {
	u.f.Text(s, int(x*u.k), int(y*u.k), u.fs(sz), c, al, u.hd)
}

func (u UI) R(x, y, w, h float64, c uint32) {
	u.f.Rect(int(x*u.k), int(y*u.k), int(math.Ceil(w*u.k)), int(math.Ceil(h*u.k)), c)
}

func (u UI) Shade(a float64) { u.f.RectA(0, int(48*u.k), u.f.W, u.f.H, 0, a) }

func (u UI) Bar(x, y, w, h, frac float64, c, bg uint32) {
	frac = clamp(frac, 0, 1)
	if !u.hd {
		u.R(x, y, w, h, bg)
		u.R(x+1, y+1, (w-2)*frac, h-2, c)
		return
	}
	k := u.k
	r := h * k / 2
	x1, x2, yc := x*k+r, (x+w)*k-r, (y+h/2)*k
	u.f.Capsule(x1, yc, x2, yc, r+3, 0x000000, 1)
	u.f.Capsule(x1, yc, x2, yc, r, bg, 1)
	if frac > 0 {
		xe := x1 + (x2-x1)*frac
		u.f.Capsule(x1, yc, xe, yc, r-2, c, 1)
		u.f.Capsule(x1, yc-r*.35, xe, yc-r*.35, r*.3, lighten(c, .5), .7)
	}
}

func pad(n, l int) string { return fmt.Sprintf("%0*d", l, max(n, 0)) }

func drawHUD(u UI) {
	if u.hd {
		u.f.VGrad(0, 0, FW, 192, 0x0c0a12, 0x1e1418)
		u.f.Rect(0, 186, FW, 3, 0xe8b040)
		u.f.Rect(0, 189, FW, 3, 0x5a3a10)
	} else {
		u.R(0, 0, 256, 48, 0)
	}
	u.T("1-"+pad(score, 6), 8, 5, 1, 0xffffff, 0)
	u.T("TOP-"+pad(max(topScore(), score), 6), 92, 5, 1, 0xf06060, 0)
	u.T("NIV-"+fmt.Sprint(level), 200, 5, 1, 0xffffff, 0)
	if state == "title" || state == "options" || state == "scores" {
		u.T("KUNG-FU TEMPLE · ÉDITION WINDOWS", 128, 26, 1, 0x9890a0, 1)
		return
	}
	u.T("PLAYER", 8, 18, 1, 0xf05830, 0)
	u.Bar(50, 18, 66, 7, P.Energy/100, 0xf08030, 0x402010)
	u.T("ENEMY", 8, 30, 1, 0x5080f0, 0)
	fr := 1.0
	if boss != nil {
		fr = float64(boss.HP) / float64(boss.Max)
	}
	u.Bar(50, 30, 66, 7, fr, 0x5890f8, 0x102040)
	for i := 0; i < 5; i++ {
		c := uint32(0x5080f0)
		if i+1 == level {
			c = 0xf08030
		}
		if u.hd {
			cx, cy := (127+float64(i)*12)*u.k, 21.5*u.k
			if i+1 == level {
				u.f.Glow(cx, cy, 44, 0xf08030, .55)
			}
			if i < 4 {
				u.f.Rect(int(cx+12), int(cy-1), 24, 3, 0x605868)
			}
			u.f.Circle(cx, cy, 15, 0, 1)
			u.f.Circle(cx, cy, 13, c, 1)
			u.f.Circle(cx-4, cy-4, 4, lighten(c, .5), .6)
			u.f.Text(fmt.Sprint(i+1), int(cx)+1, int(cy)-6, 2, 0x000000, 1, false)
		} else {
			u.R(124+float64(i)*12, 19, 7, 6, c)
			if i < 4 {
				u.R(132+float64(i)*12, 21, 3, 1, 0x888888)
			}
		}
	}
	// vies
	if u.hd {
		cx, cy := 130*u.k, 35*u.k
		u.f.Circle(cx, cy, 14, ink, 1)
		u.f.Circle(cx, cy, 12, 0x281810, 1)
		u.f.Circle(cx+3, cy+2, 10, pals["hero"].Skin, 1)
		u.f.Capsule(cx-12, cy-3, cx+12, cy-2, 2.5, 0xe02828, 1)
	} else {
		u.R(125, 32, 6, 6, pals["hero"].Skin)
		u.R(125, 31, 6, 2, pals["hero"].Hair)
		u.R(125, 33, 6, 1, 0xe02828)
	}
	u.T("×"+fmt.Sprint(max(lives, 0)), 136, 31, 1, 0xffffff, 0)
	if weapon != "" {
		w := "PISTOLET"
		if weapon == "riot" {
			w = "MITRAILLETTE"
		}
		u.T(w, 170, 31, 1, 0xf8d040, 1)
	}
	u.T("TIME", 206, 18, 1, 0xf05830, 0)
	tc := uint32(0xffffff)
	if timeLeft < 300 && (tick/15)%2 == 0 {
		tc = 0xff4040
	}
	u.T(pad(timeLeft, 4), 206, 30, 1, tc, 0)
	if cfg.Music == 0 {
		u.T("♪ OFF", 250, 40, 1, 0x888888, 2)
	}
}

func menuItem(u UI, txt string, y float64, sel bool) {
	c := uint32(0xc8c0b0)
	if sel {
		c = 0xf8d040
		blink := (tick/8)%2 == 0
		w := float64(textWidth(txt, u.fs(1))) / u.k
		if blink {
			u.T("►", 128-w/2-10, y, 1, 0xf05830, 0)
		}
		if u.hd {
			u.f.Glow(128*u.k, (y+3)*u.k, 120, 0xf8a040, .12)
		}
	}
	u.T(txt, 128, y, 1, c, 1)
}

func diffName() string { return []string{"FACILE", "NORMAL", "DIFFICILE"}[cfg.Difficulty] }

func gauge(v int) string {
	s := ""
	for i := 0; i < 10; i++ {
		if i < v {
			s += "■"
		} else {
			s += "□"
		}
	}
	return s
}

func onOff(b bool) string {
	if b {
		return "OUI"
	}
	return "NON"
}

func drawTitle(u UI) {
	u.Shade(.55)
	if u.hd {
		u.f.RectA(int(56*u.k), int(126*u.k), int(144*u.k), int(64*u.k), 0, .5)
		u.f.Rect(int(56*u.k), int(126*u.k), int(144*u.k), 2, 0xe8b040)
		u.f.Rect(int(56*u.k), int(190*u.k)-2, int(144*u.k), 2, 0xe8b040)
	}
	if u.hd {
		u.f.Glow(128*u.k, 78*u.k, 360, 0xff6020, .25)
		u.f.Text("KUNG-FU", int(128*u.k)+6, int(58*u.k)+6, 14, 0x300808, 1, false)
		u.f.Text("KUNG-FU", int(128*u.k), int(58*u.k), 14, 0xf05830, 1, false)
		u.f.RectA(0, int(58*u.k)+50, FW, 6, 0xffd060, .25)
		u.f.Text("TEMPLE", int(128*u.k), int(88*u.k), 8, 0xf8d040, 1, true)
	} else {
		u.T("KUNG-FU", 128, 58, 3, 0xf05830, 1)
		u.T("TEMPLE", 128, 84, 2, 0xf8d040, 1)
	}
	u.T("5 ÉTAGES · 5 MAÎTRES", 128, 108, 1, 0xffffff, 1)
}

func drawUI(u UI) {
	camX := cam
	inMenu := state == "title" || state == "options" || state == "scores"
	if !inMenu {
		for _, p := range pops {
			c := uint32(0xffffff)
			if p.Txt == "1UP" {
				c = 0x80ff80
			} else if p.Txt == "PARADE" {
				c = 0x80c0ff
			}
			u.T(p.Txt, p.X-camX, p.Y, 1, c, 1)
		}
	}
	drawHUD(u)
	if nameT > 0 && boss != nil && state == "play" {
		nameT--
		c := uint32(0xffffff)
		if (nameT>>3)%2 == 1 {
			c = 0xf8d040
		}
		u.T(bosses[boss.Kind].Name, 128, 96, 1, c, 1)
	}
	switch state {
	case "title":
		drawTitle(u)
		items := []string{"JOUER", "OPTIONS", "MEILLEURS SCORES", "QUITTER"}
		for i, it := range items {
			menuItem(u, it, 134+float64(i)*14, i == menuSel)
		}
		u.T("Z/↑ S/↓ : CHOISIR   M : VALIDER", 128, 200, 1, 0x9890a0, 1)
		u.T("F11 : PLEIN ÉCRAN", 128, 222, 1, 0x605868, 1)
	case "options":
		u.Shade(.7)
		u.T("OPTIONS", 128, 60, 2, 0xf8d040, 1)
		items := []string{
			"DIFFICULTÉ : " + diffName(),
			"GRAPHISMES : " + map[bool]string{true: "HD", false: "RÉTRO NES"}[cfg.HD],
			"FILTRE CRT : " + onOff(cfg.CRT),
			"MUSIQUE " + gauge(cfg.Music),
			"EFFETS  " + gauge(cfg.Sfx),
			"PLEIN ÉCRAN : " + onOff(cfg.Fullscreen),
			"RETOUR",
		}
		for i, it := range items {
			menuItem(u, it, 88+float64(i)*15, i == menuSel)
		}
		u.T("Q/D : MODIFIER   ÉCHAP : RETOUR", 128, 204, 1, 0x9890a0, 1)
	case "scores":
		u.Shade(.7)
		u.T("MEILLEURS SCORES", 128, 64, 2, 0xf8d040, 1)
		for i, h := range hiscores {
			y := 96 + float64(i)*16
			c := uint32(0xffffff)
			if i == 0 {
				c = 0xf8d040
			}
			u.T(fmt.Sprintf("%d.", i+1), 44, y, 1, c, 0)
			u.T(h.Name, 70, y, 1, c, 0)
			u.T(pad(h.Score, 6), 150, y, 1, c, 2)
			u.T("NIV "+fmt.Sprint(h.Level), 214, y, 1, 0x9890a0, 2)
		}
		u.T("M : RETOUR", 128, 200, 1, 0x9890a0, 1)
	case "intro":
		u.Shade(.5)
		u.T("ÉTAGE "+fmt.Sprint(level), 128, 84, 2, 0xf8d040, 1)
		ar := "► ► ►"
		txt := "AVANCEZ VERS LA DROITE"
		if dir < 0 {
			ar, txt = "◄ ◄ ◄", "AVANCEZ VERS LA GAUCHE"
		}
		u.T(ar, 128, 106, 2, 0xffffff, 1)
		u.T(txt, 128, 126, 1, 0xd8d0c8, 1)
		u.T("BOSS : "+bosses[levels[level].Boss].Name, 128, 142, 1, 0xf06060, 1)
		if level == 1 {
			u.T("Q/D MARCHER · S ACCROUPI · Z/↑/L SAUTER", 128, 166, 1, 0x9890a0, 1)
			u.T("M PIED · K POING · Q/D VITE = SE LIBÉRER", 128, 178, 1, 0x9890a0, 1)
		}
	case "pause":
		u.Shade(.6)
		u.T("PAUSE", 128, 84, 2, 0xffffff, 1)
		for i, it := range []string{"REPRENDRE", "OPTIONS", "ABANDONNER"} {
			menuItem(u, it, 116+float64(i)*15, i == menuSel)
		}
	case "clear":
		if stT > 60 {
			u.Shade(.4)
			u.T("ÉTAGE "+fmt.Sprint(level)+" TERMINÉ !", 128, 96, 2, 0xf8d040, 1)
			u.T("BONUS TEMPS", 128, 122, 1, 0xffffff, 1)
		}
	case "gameover":
		u.Shade(.65)
		u.T("GAME OVER", 128, 96, 3, 0xf05830, 1)
		u.T("SCORE "+pad(score, 6), 128, 132, 1, 0xffffff, 1)
		if stT > 90 && (tick>>5)%2 == 0 {
			u.T("APPUYEZ SUR M", 128, 162, 1, 0xd8d0c8, 1)
		}
	case "win":
		u.Shade(.65)
		u.T("VICTOIRE !", 128, 80, 3, 0xf8d040, 1)
		u.T("LE TEMPLE EST LIBÉRÉ.", 128, 116, 1, 0xffffff, 1)
		u.T("SCORE "+pad(score, 6), 128, 134, 1, 0xffffff, 1)
		if stT > 90 && (tick>>5)%2 == 0 {
			u.T("APPUYEZ SUR M", 128, 166, 1, 0xd8d0c8, 1)
		}
	case "name":
		u.Shade(.7)
		u.T("NOUVEAU RECORD !", 128, 70, 2, 0xf8d040, 1)
		u.T("SCORE "+pad(score, 6), 128, 96, 1, 0xffffff, 1)
		for i := 0; i < 3; i++ {
			x := 104 + float64(i)*24
			c := uint32(0xffffff)
			if i == namePos {
				c = 0xf8d040
				if (tick/8)%2 == 0 {
					u.R(x-6, 142, 12, 2, 0xf8d040)
				}
			}
			u.T(string(nameChars[nameBuf[i]]), x, 122, 2, c, 1)
		}
		u.T("Z/S LETTRE · Q/D POSITION · M OK", 128, 170, 1, 0x9890a0, 1)
	}
}

// render produit l'image finale 1024x960.
var small = NewFB(256, 240)

func flashFX(f *FB) {
	if flash > 0 {
		f.RectA(0, f.H/5, f.W, f.H, 0xffffff, float64(flash)/30)
	}
}

func render(big *FB) {
	camX := cam
	if state == "title" || state == "options" || state == "scores" {
		camX = titleCam
	}
	if cfg.HD {
		renderHD(big, camX)
		flashFX(big)
		drawUI(UI{big, 4, true})
		if cfg.CRT {
			big.CRT(false)
		}
		return
	}
	renderRetro(small, camX)
	flashFX(small)
	drawUI(UI{small, 1, false})
	big.UpscaleFrom(small, 4)
	if cfg.CRT {
		big.CRT(true)
	}
}
