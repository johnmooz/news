package main

import (
	"math"
	"math/rand"
	"strconv"
)

const (
	GY    = 190.0
	LVLEN = 1800.0
	VW    = 256.0
)

type Box struct {
	XA, XB, YA, YB float64
	Dmg            int
}

func ov(a, b Box) bool { return a.XA < b.XB && a.XB > b.XA && a.YA < b.YB && a.YB > b.YA }
func clamp(v, a, b float64) float64 {
	if v < a {
		return a
	}
	if v > b {
		return b
	}
	return v
}
func sgn(v float64) float64 {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}
func rnd() float64 { return rand.Float64() }

type Input struct{ Left, Right, Up, Down, Jump, Atk, Start, Back, Punch bool }

type Anim struct {
	Cur  Pose
	Init bool
	Rot  float64
}

type Ent struct {
	Type, Kind                     string
	X, Y, VX, VY, Face, CD, Off    float64
	St                             string
	T, HP, Max, HitID, Flash, ArmN int
	Dying, Dead, In, High, HasBoom bool
	A                              Anim
}

type Shot struct {
	K                  string
	X, H, VX, VY, Dist float64
	Ret, Dead          bool
	Owner              *Ent
	Spin               int
}

type Pop struct {
	X, Y float64
	Txt  string
	T    int
}

type Part struct {
	X, H, VX, VH, Grav, Size float64
	Life, Max                int
	C                        uint32
	Glow                     bool
}

type Attack struct {
	Type  string
	T, ID int
}

type Player struct {
	X, Y, VX, VY, Face, Energy float64
	Atk                        *Attack
	Inv, Hurt, Walk, ShakeT    int
	ShootT                     int
	Crouch                     bool
	A                          Anim
}

type AtkSpec struct {
	Dur, A, B int
	Reach     float64
	Dmg       int
}

var atkSpecs = map[string]AtkSpec{
	"kick": {18, 5, 11, 25, 2}, "punch": {12, 3, 7, 19, 3}, "lowkick": {18, 5, 11, 25, 2},
	"lowpunch": {12, 3, 7, 19, 3}, "airkick": {90, 4, 90, 23, 3},
}

var (
	state      = "title"
	stT        int
	level      = 1
	score      int
	lives      = 3
	timeLeft   = 2000
	cam        float64
	dir        = -1.0
	bossZone   bool
	boss       *Ent
	tick       int
	nameT      int
	P          Player
	ents       []*Ent
	shots      []*Shot
	pops       []*Pop
	parts      []*Part
	atkID      int
	spawnT     float64
	vaseT      float64
	shakeN     float64
	lastShake  int
	hitstop    int
	shake      int
	nextLife   = 50000
	menuSel    int
	optFrom    string
	nameBuf    = [3]int{0, 0, 0}
	namePos    int
	inp, prevI Input
	pr         Input
	titleCam   float64
	quitReq    bool
	weapon     string
	gunCD      int
	cheatBuf   string
	flash      int
)

const nameChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ.-"

func dmgK() float64   { return []float64{.6, 1, 1.4}[cfg.Difficulty] }
func spawnK() float64 { return []float64{1.3, 1, .8}[cfg.Difficulty] }
func bossK() float64  { return []float64{.75, 1, 1.25}[cfg.Difficulty] }

func addScore(n int, x, h float64, pop bool) {
	score += n
	if score >= nextLife {
		lives++
		nextLife += 50000
		sfx("life")
		pops = append(pops, &Pop{P.X, GY - 50, "1UP", 70})
	}
	if pop {
		pops = append(pops, &Pop{x, GY - h, strconv.Itoa(n), 50})
	}
}

func burst(x, h float64, n int, c uint32, sp float64, glow bool, grav float64) {
	for i := 0; i < n; i++ {
		a := rnd() * math.Pi * 2
		v := sp * (.4 + rnd()*.8)
		l := 14 + rand.Intn(14)
		parts = append(parts, &Part{X: x, H: h, VX: math.Cos(a) * v, VH: math.Sin(a)*v + sp*.3, Grav: grav, Size: .6 + rnd()*.9, Life: l, Max: l, C: c, Glow: glow})
	}
}

func dust(x float64, n int) {
	for i := 0; i < n; i++ {
		l := 16 + rand.Intn(12)
		parts = append(parts, &Part{X: x + (rnd()-.5)*8, H: 1, VX: (rnd() - .5) * 1.2, VH: rnd() * .6, Grav: 0, Size: 1 + rnd()*1.2, Life: l, Max: l, C: 0xd8b890})
	}
}

func startGame() {
	level, score, lives, nextLife = 1, 0, 3, 50000
	weapon, gunCD = "", 0
	sfx("start")
	startLevel()
}

func startLevel() {
	c := levels[level]
	dir = c.Dir
	P = Player{X: LVLEN - 60, Face: dir, Inv: 60, Energy: 100}
	if dir > 0 {
		P.X = 60
	}
	ents, shots, pops, parts = nil, nil, nil, nil
	timeLeft, bossZone, boss = 2000, false, nil
	spawnT, vaseT, shakeN, hitstop, shake = 110, 220, 0, 0, 0
	cam = clamp(P.X-VW/2+dir*40, 0, LVLEN-VW)
	state, stT = "intro", 0
}

/* ---------- Boîtes de collision ---------- */

func playerHB() Box {
	if P.Crouch && P.Y <= 0 {
		return Box{P.X - 5, P.X + 5, 0, 18, 0}
	}
	return Box{P.X - 5, P.X + 5, P.Y, P.Y + 30, 0}
}

func entHB(e *Ent) Box {
	switch e.Type {
	case "tomtom":
		return Box{e.X - 5, e.X + 5, e.Y, e.Y + 16, 0}
	case "snake":
		return Box{e.X - 7, e.X + 7, 0, 5, 0}
	case "boss":
		s := bosses[e.Kind].Scale
		return Box{e.X - 6*s, e.X + 6*s, e.Y, e.Y + 31*s, 0}
	}
	return Box{e.X - 5, e.X + 5, e.Y, e.Y + 30, 0}
}

func atkActive() bool {
	if P.Atk == nil {
		return false
	}
	d := atkSpecs[P.Atk.Type]
	return P.Atk.T >= d.A && P.Atk.T <= d.B
}

func atkBox() Box {
	a := P.Atk
	d := atkSpecs[a.Type]
	var ya, yb float64
	switch a.Type {
	case "kick", "punch":
		ya, yb = 16, 28
	case "airkick":
		ya, yb = P.Y+8, P.Y+22
	default:
		ya, yb = 1, 12
	}
	x1, x2 := P.X+P.Face*3, P.X+P.Face*d.Reach
	return Box{math.Min(x1, x2), math.Max(x1, x2), ya, yb, d.Dmg}
}

/* ---------- Entités ---------- */

func makeEnemy(t string, x float64) *Ent {
	hp := 1
	if t == "knifer" {
		hp = 2
	}
	return &Ent{Type: t, X: x, Face: 1, St: "walk", HP: hp, CD: 40 + rnd()*40, HitID: -1}
}

func aliveCount() int {
	n := 0
	for _, e := range ents {
		if !e.Dying {
			n++
		}
	}
	return n
}

func spawnEnemy(c LevelCfg) {
	tot := 0
	for _, w := range c.W {
		tot += w.W
	}
	r := rnd() * float64(tot)
	t := "gripper"
	for _, w := range c.W {
		r -= float64(w.W)
		if r < 0 {
			t = w.Name
			break
		}
	}
	side := -dir
	if rnd() < .68 {
		side = dir
	}
	x := cam - 12
	if side > 0 {
		x = cam + VW + 12
	}
	if x < 4 || x > LVLEN-4 {
		return
	}
	ents = append(ents, makeEnemy(t, x))
}

func spawnBoss() {
	k := levels[level].Boss
	s := bosses[k]
	hp := int(math.Round(float64(s.HP) * bossK()))
	x := LVLEN - 26
	if dir < 0 {
		x = 26
	}
	boss = &Ent{Type: "boss", Kind: k, X: x, Face: -dir, St: "approach", HP: hp, Max: hp, HitID: -1, CD: 60}
	ents = append(ents, boss)
	nameT = 160
	sfx("boss")
}

func kill(e *Ent, pts int) {
	e.Dying, e.St = true, "dead"
	f := sgn(e.X - P.X)
	if f == 0 {
		f = 1
	}
	e.VX, e.VY, e.Y = f*1.6, 2.6, math.Max(e.Y, .01)
	addScore(pts, e.X, e.Y+34, true)
}

func isGrabbed() bool {
	for _, e := range ents {
		if e.Type == "gripper" && e.St == "grab" && !e.Dying {
			return true
		}
	}
	return false
}

func hitPlayer(d float64) {
	if P.Inv > 0 || state != "play" {
		return
	}
	P.Energy -= d * dmgK()
	P.Inv, P.Hurt = 45, 14
	shake = 8
	sfx("hurt")
	burst(P.X, P.Y+20, 8, 0xffffff, 1.6, true, .05)
	if P.Energy <= 0 {
		killPlayer()
	}
}

func killPlayer() {
	P.Energy = 0
	state, stT = "dying", 0
	P.Atk, P.Crouch = nil, false
	sfx("die")
	for _, e := range ents {
		if e.Type == "gripper" && e.St == "grab" {
			e.St = "walk"
			e.X += e.Off * 2
		}
	}
}

func onHit(e *Ent, b Box) {
	hb := entHB(e)
	h := (math.Max(b.YA, hb.YA) + math.Min(b.YB, hb.YB)) / 2
	burst(e.X-P.Face*3, h, 10, 0xfff0a0, 2.2, true, .08)
	parts = append(parts, &Part{X: e.X - P.Face*3, H: h, Life: 8, Max: 8, Size: 6, C: 0xffffff, Glow: true})
	if e.Type == "boss" {
		bossHit(e, b.Dmg)
		return
	}
	sfx("hit")
	hitstop = 3
	switch e.Type {
	case "gripper":
		kill(e, 100)
	case "tomtom":
		if e.Y > 0 {
			kill(e, 500)
		} else {
			kill(e, 300)
		}
	case "snake":
		kill(e, 200)
	case "knifer":
		e.HP--
		if e.HP <= 0 {
			kill(e, 500)
		} else {
			e.St, e.T = "stagger", 0
			e.X += sgn(e.X-P.X) * 6
		}
	}
}

func bossHit(b *Ent, dmg int) {
	s := bosses[b.Kind]
	if b.St == "vanish" || b.St == "appear" {
		return
	}
	if s.Block > 0 && (b.St == "approach" || b.St == "back") && rnd() < s.Block {
		sfx("block")
		pops = append(pops, &Pop{b.X, GY - 40, "PARADE", 30})
		return
	}
	sfx("hit")
	hitstop, shake = 5, 6
	b.HP -= dmg
	b.Flash = 10
	addScore(100, 0, 0, false)
	if b.HP <= 0 {
		b.HP = 0
		b.Dying, b.St = true, "dead"
		f := sgn(b.X - P.X)
		if f == 0 {
			f = 1
		}
		b.VX, b.VY, b.Y = f*1.4, 3, .01
		addScore(s.Pts, b.X, 48, true)
		state, stT = "clear", 0
		shots = nil
		shake = 20
		for _, e := range ents {
			if e != b && !e.Dying {
				e.Dying = true
				e.VX = sgn(e.X - P.X)
				if e.VX == 0 {
					e.VX = 1
				}
				e.VY, e.Y = 2, math.Max(.01, e.Y)
			}
		}
		sfx("bossdown")
		return
	}
	stag := true
	if s.Armor > 0 {
		b.ArmN++
		stag = b.ArmN >= s.Armor
		if stag {
			b.ArmN = 0
		}
	}
	if stag && b.St != "jk" {
		b.St, b.T = "hurt", 0
		if f := sgn(P.X - b.X); f != 0 {
			b.Face = f
		}
	}
}

/* ---------- Boucle principale ---------- */

func edge(a, b bool) bool { return a && !b }

func menuNav(n int) {
	if pr.Up {
		menuSel = (menuSel + n - 1) % n
		sfx("menu")
	}
	if pr.Down {
		menuSel = (menuSel + 1) % n
		sfx("menu")
	}
}

func update(in Input) {
	inp = in
	pr = Input{edge(in.Left, prevI.Left), edge(in.Right, prevI.Right), edge(in.Up, prevI.Up), edge(in.Down, prevI.Down),
		edge(in.Jump, prevI.Jump), edge(in.Atk, prevI.Atk), edge(in.Start, prevI.Start), edge(in.Back, prevI.Back), edge(in.Punch, prevI.Punch)}
	prevI = in
	tick++
	ok := pr.Atk || pr.Start || pr.Punch
	if flash > 0 {
		flash--
	}
	switch state {
	case "title":
		titleCam = math.Mod(titleCam+.3, LVLEN-VW)
		menuNav(4)
		if ok {
			sfx("select")
			switch menuSel {
			case 0:
				startGame()
			case 1:
				optFrom, state, menuSel = "title", "options", 0
			case 2:
				state = "scores"
			case 3:
				quitReq = true
			}
		}
	case "options":
		updateOptions()
	case "scores":
		titleCam = math.Mod(titleCam+.3, LVLEN-VW)
		if ok || pr.Back {
			state, menuSel = "title", 2
			sfx("select")
		}
	case "intro":
		stT++
		if stT > 170 || (stT > 30 && ok) {
			state = "play"
		}
	case "play":
		if pr.Back || pr.Start {
			state, menuSel = "pause", 0
			sfx("select")
			break
		}
		updatePlay()
	case "pause":
		menuNav(3)
		if pr.Back {
			state = "play"
		} else if ok {
			sfx("select")
			switch menuSel {
			case 0:
				state = "play"
			case 1:
				optFrom, state, menuSel = "pause", "options", 0
			case 2:
				endGame()
			}
		}
	case "dying":
		stT++
		updateDying()
		if stT > 140 {
			lives--
			if lives < 0 {
				state, stT = "gameover", 0
			} else {
				startLevel()
			}
		}
	case "clear":
		stT++
		updateDying()
		if stT > 100 && timeLeft > 0 {
			d := min(timeLeft, 12)
			timeLeft -= d
			addScore(d*5, 0, 0, false)
			if tick%3 == 0 {
				sfx("tick")
			}
		}
		if stT > 100 && timeLeft <= 0 && stT > 300 {
			level++
			if level > 5 {
				state, stT = "win", 0
			} else {
				startLevel()
			}
		}
	case "gameover", "win":
		stT++
		if stT > 90 && ok {
			endGame()
		}
	case "name":
		updateName()
	}
	if state != "pause" && state != "options" {
		for _, p := range pops {
			p.T--
			p.Y -= .4
		}
		pops = filter(pops, func(p *Pop) bool { return p.T > 0 })
		for _, p := range parts {
			p.Life--
			p.X += p.VX
			p.H += p.VH
			p.VH -= p.Grav
			p.VX *= .96
			if p.H < 0 {
				p.H, p.VH = 0, -p.VH*.3
			}
		}
		parts = filter(parts, func(p *Part) bool { return p.Life > 0 })
		if shake > 0 {
			shake--
		}
	}
}

func filter[T any](s []T, f func(T) bool) []T {
	out := s[:0]
	for _, v := range s {
		if f(v) {
			out = append(out, v)
		}
	}
	return out
}

func endGame() {
	if qualifies(score) {
		state, namePos, nameBuf = "name", 0, [3]int{0, 0, 0}
	} else {
		state, menuSel = "title", 0
	}
}

func updateName() {
	if pr.Up {
		nameBuf[namePos] = (nameBuf[namePos] + 1) % len(nameChars)
		sfx("menu")
	}
	if pr.Down {
		nameBuf[namePos] = (nameBuf[namePos] + len(nameChars) - 1) % len(nameChars)
		sfx("menu")
	}
	if pr.Left && namePos > 0 {
		namePos--
		sfx("menu")
	}
	if pr.Right && namePos < 2 {
		namePos++
		sfx("menu")
	}
	if pr.Atk || pr.Start {
		sfx("select")
		if namePos < 2 && !pr.Start {
			namePos++
		} else {
			n := string([]byte{nameChars[nameBuf[0]], nameChars[nameBuf[1]], nameChars[nameBuf[2]]})
			insertScore(n, score, level)
			state, menuSel = "scores", 0
		}
	}
}

func updateOptions() {
	menuNav(7)
	ch := 0
	if pr.Left {
		ch = -1
	}
	if pr.Right || pr.Atk || pr.Start {
		ch = 1
	}
	if pr.Back || (menuSel == 6 && (pr.Atk || pr.Start)) {
		writeSave()
		state = optFrom
		menuSel = 1
		sfx("select")
		return
	}
	if ch == 0 {
		return
	}
	sfx("menu")
	switch menuSel {
	case 0:
		cfg.Difficulty = (cfg.Difficulty + 3 + ch) % 3
	case 1:
		cfg.HD = !cfg.HD
	case 2:
		cfg.CRT = !cfg.CRT
	case 3:
		cfg.Music = int(clamp(float64(cfg.Music+ch), 0, 10))
	case 4:
		cfg.Sfx = int(clamp(float64(cfg.Sfx+ch), 0, 10))
	case 5:
		cfg.Fullscreen = !cfg.Fullscreen
		platFullscreen(cfg.Fullscreen)
	}
}

func updateDying() {
	for _, e := range ents {
		if e.Dying {
			e.X += e.VX
			e.Y += e.VY
			e.VY -= .2
			e.A.Rot += .12 * sgn(e.VX)
			if e.Y < -90 {
				e.Dead = true
			}
		}
	}
	ents = filter(ents, func(e *Ent) bool { return !e.Dead })
}

func updatePlay() {
	if hitstop > 0 {
		hitstop--
		return
	}
	c := levels[level]
	if tick%10 == 0 {
		timeLeft--
		if timeLeft <= 0 {
			timeLeft = 0
			killPlayer()
			return
		}
	}
	updatePlayer()
	if state != "play" {
		return
	}
	if !bossZone {
		tg := clamp(P.X-VW/2+dir*40, 0, LVLEN-VW)
		cam += (tg - cam) * .12
		d := LVLEN - P.X
		if dir < 0 {
			d = P.X
		}
		if d < 230 {
			bossZone = true
			spawnBoss()
		}
	} else {
		tg := 0.0
		if dir > 0 {
			tg = LVLEN - VW
		}
		cam += (tg - cam) * .08
	}
	P.X = clamp(P.X, math.Max(8, cam+8), math.Min(LVLEN-8, cam+VW-8))
	if !bossZone {
		spawnT--
		if spawnT <= 0 {
			r := c.Rate * spawnK()
			spawnT = r + rnd()*r*.6
			if aliveCount() < c.Max {
				spawnEnemy(c)
			}
		}
		if c.Vases {
			vaseT--
			if vaseT <= 0 {
				vaseT = 150 + rnd()*130
				shots = append(shots, &Shot{K: "vase", X: clamp(P.X+dir*(30+rnd()*70), cam+10, cam+VW-10), H: 112})
			}
		}
	}
	if atkActive() {
		box := atkBox()
		for _, e := range ents {
			if e.Dying || e.HitID == P.Atk.ID {
				continue
			}
			if ov(box, entHB(e)) {
				e.HitID = P.Atk.ID
				onHit(e, box)
				if state != "play" {
					return
				}
			}
		}
		for _, s := range shots {
			if s.K == "vase" && !s.Dead && ov(box, Box{s.X - 4, s.X + 4, s.H, s.H + 9, 0}) {
				s.Dead = true
				sfx("vase")
				addScore(200, s.X, s.H+12, true)
				burst(s.X, s.H+4, 12, 0xc87830, 1.8, false, .15)
			}
		}
	}
	for _, e := range ents {
		if e.Dying {
			e.X += e.VX
			e.Y += e.VY
			e.VY -= .2
			e.A.Rot += .12 * sgn(e.VX)
			if e.Y < -90 {
				e.Dead = true
			}
			continue
		}
		dx := P.X - e.X
		ad := math.Abs(dx)
		e.T++
		if e.Type != "boss" && e.X > cam+10 && e.X < cam+VW-10 {
			e.In = true
		}
		switch e.Type {
		case "gripper":
			if e.St == "walk" {
				e.Face = sgn(dx)
				if e.Face == 0 {
					e.Face = 1
				}
				e.X += e.Face * .85
				if ad < 9 && P.Y < 4 && P.Inv == 0 {
					e.St = "grab"
					s := 1.0
					if e.X < P.X {
						s = -1
					}
					e.Off = s * (6 + rnd()*3)
					sfx("grab")
				}
			} else {
				e.X = P.X + e.Off
				e.Face = -sgn(e.Off)
				P.Energy -= .11 * dmgK()
				if P.Energy <= 0 {
					killPlayer()
					return
				}
			}
		case "knifer":
			if f := sgn(dx); f != 0 {
				e.Face = f
			}
			switch e.St {
			case "stagger":
				if e.T > 20 {
					e.St, e.T = "walk", 0
				}
			case "throw":
				if e.T == 14 {
					h := 5.0
					if e.High {
						h = 24
					}
					shots = append(shots, &Shot{K: "knife", X: e.X + e.Face*8, H: h, VX: e.Face * 2.4})
					sfx("knife")
				}
				if e.T >= 28 {
					e.St, e.T, e.CD = "walk", 0, 60+rnd()*60
				}
			default:
				if ad > 85 {
					e.X += e.Face * .9
				} else if ad < 55 {
					e.X -= e.Face * .7
				}
				e.CD--
				if e.CD <= 0 && ad < 140 && e.In {
					e.St, e.T, e.High = "throw", 0, rnd() < .5
				}
			}
			if e.In {
				e.X = clamp(e.X, cam+10, cam+VW-10)
			}
		case "tomtom":
			if e.Y > 0 || e.VY > 0 {
				e.X += e.VX
				e.Y += e.VY
				e.VY -= .2
				e.A.Rot += .3 * e.Face
				if e.Y <= 0 {
					e.Y, e.VY, e.VX, e.A.Rot = 0, 0, 0, 0
					dust(e.X, 3)
				}
			} else {
				if f := sgn(dx); f != 0 {
					e.Face = f
				}
				e.CD--
				if ad < 42 && e.CD <= 0 {
					e.VY, e.VX, e.Y, e.CD = 3.6, e.Face*1.7, .01, 80
				} else if ad > 6 {
					e.X += e.Face * 1.1
				}
			}
			if P.Inv == 0 && ov(entHB(e), playerHB()) {
				hitPlayer(6)
				e.VX, e.VY, e.Y = -e.Face*1.5, 2.4, math.Max(e.Y, .01)
			}
		case "snake":
			if f := sgn(dx); f != 0 {
				e.Face = f
			}
			e.X += e.Face * 1.2
			if ov(entHB(e), playerHB()) {
				hitPlayer(5)
			}
		case "boss":
			updateBoss(e)
		}
		if state != "play" {
			return
		}
	}
	ents = filter(ents, func(e *Ent) bool {
		return !e.Dead && (e.Type == "boss" || e.Dying || (e.X > cam-60 && e.X < cam+VW+60))
	})
	for _, s := range shots {
		if s.Dead {
			continue
		}
		switch s.K {
		case "bullet":
			s.X += s.VX
			dmg := 2
			if weapon == "riot" {
				dmg = 1
			}
			for _, e := range ents {
				if e.Dying {
					continue
				}
				hb := entHB(e)
				if s.X > hb.XA && s.X < hb.XB && s.H >= hb.YA && s.H <= hb.YB {
					s.Dead = true
					onHit(e, Box{0, 0, s.H - 1, s.H + 1, dmg})
					break
				}
			}
			for _, v := range shots {
				if v.K == "vase" && !v.Dead && math.Abs(v.X-s.X) < 5 && s.H >= v.H && s.H <= v.H+9 {
					v.Dead, s.Dead = true, true
					addScore(200, v.X, v.H+12, true)
				}
			}
		case "knife":
			s.X += s.VX
			if ov(Box{s.X - 4, s.X + 4, s.H, s.H + 2, 0}, playerHB()) {
				hitPlayer(8)
				s.Dead = true
			}
		case "fire":
			s.X += s.VX
			if tick%2 == 0 {
				parts = append(parts, &Part{X: s.X - s.VX*2, H: s.H + 2 + (rnd()-.5)*3, VX: -s.VX * .2, VH: .2, Size: 1.2 + rnd(), Life: 14, Max: 14, C: 0xff8020, Glow: true})
			}
			if ov(Box{s.X - 4, s.X + 4, s.H, s.H + 5, 0}, playerHB()) {
				hitPlayer(10)
				s.Dead = true
			}
		case "vase":
			s.VY -= .1
			s.H += s.VY
			if ov(Box{s.X - 4, s.X + 4, s.H, s.H + 9, 0}, playerHB()) {
				hitPlayer(6)
				s.Dead = true
				burst(s.X, s.H+4, 12, 0xc87830, 1.8, false, .15)
				break
			}
			if s.H <= 0 {
				s.Dead = true
				sfx("vase")
				burst(s.X, 3, 12, 0xc87830, 1.8, false, .15)
				if rnd() < .65 {
					ents = append(ents, makeEnemy("snake", s.X))
				}
			}
		case "boomer":
			o := s.Owner
			s.Spin++
			if !s.Ret {
				s.X += s.VX
				s.Dist += math.Abs(s.VX)
				if s.Dist > 115 {
					s.Ret = true
				}
			} else {
				s.X += sgn(o.X-s.X) * 2.6
				if math.Abs(o.X-s.X) < 5 || o.Dying {
					s.Dead = true
					o.HasBoom = false
				}
			}
			if ov(Box{s.X - 4, s.X + 4, s.H, s.H + 4, 0}, playerHB()) {
				hitPlayer(9)
			}
		}
		if s.K != "boomer" && s.K != "vase" && (s.X < cam-20 || s.X > cam+VW+20) {
			s.Dead = true
		}
	}
	shots = filter(shots, func(s *Shot) bool { return !s.Dead })
}

func updatePlayer() {
	if P.Inv > 0 {
		P.Inv--
	}
	if P.Hurt > 0 {
		P.Hurt--
	}
	if P.ShakeT > 0 {
		P.ShakeT--
	}
	if P.ShootT > 0 {
		P.ShootT--
	}
	onGround := P.Y <= 0 && P.VY <= 0
	if isGrabbed() {
		P.Atk, P.Crouch = nil, false
		if pr.Left || pr.Right {
			d := 1
			if pr.Left {
				d = -1
			}
			if d != lastShake {
				shakeN++
			} else {
				shakeN += .4
			}
			lastShake = d
			P.Face = float64(d)
			P.ShakeT = 6
		}
		if shakeN >= 8 {
			shakeN = 0
			sfx("free")
			for _, e := range ents {
				if e.Type == "gripper" && e.St == "grab" && !e.Dying {
					kill(e, 100)
				}
			}
			P.Inv = 20
		}
	} else {
		shakeN = 0
		if onGround {
			if P.Atk != nil && (P.Atk.Type == "lowkick" || P.Atk.Type == "lowpunch") {
				P.Crouch = true
			} else if P.Atk != nil {
				P.Crouch = false
			} else {
				P.Crouch = inp.Down
			}
			if P.Atk == nil {
				if pr.Jump && !P.Crouch {
					P.VY = 4.0
					P.VX = 0
					if inp.Left {
						P.VX = -1.3
					} else if inp.Right {
						P.VX = 1.3
					}
					if P.VX != 0 {
						P.Face = sgn(P.VX)
					}
					P.Y = .01
					sfx("jump")
					dust(P.X, 4)
				} else if !P.Crouch {
					mv := 0.0
					if inp.Left {
						mv = -1
					}
					if inp.Right {
						mv = 1
					}
					if mv != 0 {
						P.Face = mv
						P.X += mv * 1.25
						P.Walk++
						if P.Walk%16 == 0 {
							dust(P.X, 1)
						}
					} else {
						P.Walk = 0
					}
				} else {
					if inp.Left {
						P.Face = -1
					}
					if inp.Right {
						P.Face = 1
					}
				}
			}
			if pr.Atk && P.Atk == nil && P.Y <= 0 {
				if P.Crouch {
					startAttack("lowkick")
				} else {
					startAttack("kick")
				}
			} else if pr.Punch && P.Atk == nil && P.Y <= 0 && weapon == "" {
				if P.Crouch {
					startAttack("lowpunch")
				} else {
					startAttack("punch")
				}
			}
		} else if (pr.Atk || (pr.Punch && weapon == "")) && P.Atk == nil {
			startAttack("airkick")
		}
		if weapon != "" {
			if gunCD > 0 {
				gunCD--
			}
			want := pr.Punch
			if weapon == "riot" {
				want = inp.Punch
			}
			if want && gunCD <= 0 {
				shoot()
			}
		}
	}
	if P.Y > 0 || P.VY > 0 {
		P.Y += P.VY
		P.VY -= .22
		P.X += P.VX
		if P.Y <= 0 {
			P.Y, P.VY, P.VX = 0, 0, 0
			if P.Atk != nil && P.Atk.Type == "airkick" {
				P.Atk = nil
			}
			dust(P.X, 5)
			sfx("land")
		}
	}
	if P.Atk != nil {
		P.Atk.T++
		if P.Atk.T >= atkSpecs[P.Atk.Type].Dur {
			P.Atk = nil
		}
	}
}

func shoot() {
	h := P.Y + 21
	if P.Crouch && P.Y <= 0 {
		h = 4
	}
	shots = append(shots, &Shot{K: "bullet", X: P.X + P.Face*14, H: h, VX: P.Face * 5.5})
	parts = append(parts, &Part{X: P.X + P.Face*17, H: h, Life: 4, Max: 4, Size: 2, C: 0xfff080, Glow: true})
	P.ShootT = 7
	if weapon == "riot" {
		gunCD = 5
		sfx("riot")
	} else {
		gunCD = 16
		sfx("gun")
	}
}

// cheatKey reçoit chaque lettre tapée pendant la partie.
func cheatKey(r rune) {
	if state != "play" || r < 'a' || r > 'z' {
		return
	}
	cheatBuf += string(r)
	if len(cheatBuf) > 12 {
		cheatBuf = cheatBuf[len(cheatBuf)-12:]
	}
	msg := func(t string) { pops = append(pops, &Pop{P.X, GY - 62, t, 100}) }
	for _, c := range []string{"gungun", "riotgun", "life", "boom"} {
		if len(cheatBuf) < len(c) || cheatBuf[len(cheatBuf)-len(c):] != c {
			continue
		}
		cheatBuf = ""
		sfx("cheat")
		switch c {
		case "gungun":
			weapon = "gun"
			msg("PISTOLET !")
		case "riotgun":
			weapon = "riot"
			msg("MITRAILLETTE !")
		case "life":
			P.Energy = 100
			msg("VIE 100% !")
		case "boom":
			flash, shake = 24, 20
			sfx("boom")
			shots = filter(shots, func(s *Shot) bool { return s.K == "bullet" })
			for _, e := range ents {
				if e.Dying || e.X < cam-10 || e.X > cam+VW+10 {
					continue
				}
				burst(e.X, 16, 16, 0xffa040, 2.5, true, .05)
				if e.Type == "boss" {
					if state == "play" {
						e.St = "recover"
						bossHit(e, 9999)
					}
				} else {
					kill(e, 100)
				}
			}
			msg("BOUM !")
		}
		return
	}
}

func closeFoe() bool {
	for _, e := range ents {
		if !e.Dying && math.Abs(e.X-P.X) < 15 && sgn(e.X-P.X) == P.Face {
			return true
		}
	}
	return false
}

func startAttack(t string) {
	atkID++
	P.Atk = &Attack{t, 0, atkID}
	sfx("swing")
}

func updateBoss(b *Ent) {
	s := bosses[b.Kind]
	dx := P.X - b.X
	ad := math.Abs(dx)
	if b.Flash > 0 {
		b.Flash--
	}
	if b.Y > 0 || b.VY > 0 {
		b.X += b.VX
		b.Y += b.VY
		b.VY -= .2
		if b.Y <= 0 {
			b.Y, b.VY, b.VX = 0, 0, 0
			dust(b.X, 6)
		}
	}
	faceP := func() {
		if f := sgn(dx); f != 0 {
			b.Face = f
		}
	}
	switch b.St {
	case "approach":
		faceP()
		b.CD--
		if s.Magic {
			if ad < s.Reach-4 && b.CD <= 0 {
				b.St, b.T, b.High = "wind", 0, rnd() < .5
				break
			}
			if ad < 40 {
				b.X -= b.Face * s.Speed
			}
			if b.CD <= 0 {
				if rnd() < .45 {
					b.St, b.T = "vanish", 0
					sfx("magic")
				} else {
					b.St, b.T, b.High = "cast", 0, rnd() < .5
				}
			}
			break
		}
		if s.Ranged && ad > 60 && b.CD <= 0 && !b.HasBoom {
			b.St, b.T, b.High = "throw", 0, rnd() < .5
			break
		}
		if b.Kind == "maitre" && ad > 45 && ad < 85 && b.CD <= 0 && rnd() < .03 {
			b.St, b.T, b.VY, b.VX, b.Y = "jk", 0, 3.4, b.Face*2.2, .01
			sfx("jump")
			break
		}
		if ad > s.Reach-6 {
			b.X += b.Face * s.Speed
		} else if b.CD <= 0 {
			b.St, b.T, b.High = "wind", 0, rnd() < .5
		}
	case "wind":
		if b.T >= s.Wind {
			b.St, b.T = "active", 0
			sfx("swing")
		}
	case "active":
		ya, yb := 0.0, 12.0
		if b.High {
			ya, yb = 20, 30
		}
		box := Box{math.Min(b.X, b.X+b.Face*s.Reach), math.Max(b.X, b.X+b.Face*s.Reach), ya, yb, 0}
		if ov(box, playerHB()) {
			hitPlayer(s.Dmg)
		}
		if b.T >= s.Active {
			b.St, b.T = "recover", 0
		}
	case "recover":
		if b.T >= s.Rec {
			b.St, b.T = "back", 0
		}
	case "back":
		b.X -= b.Face * s.Speed * .8
		if b.T >= s.Back {
			b.St, b.T, b.CD = "approach", 0, s.CD*(.6+rnd()*.8)
		}
	case "hurt":
		if b.T < 8 {
			b.X -= b.Face * 1.2
		}
		if b.T >= 14 {
			b.St, b.T = "back", 0
		}
	case "throw":
		if b.T == 16 {
			h := 6.0
			if b.High {
				h = 23
			}
			shots = append(shots, &Shot{K: "boomer", X: b.X + b.Face*8, H: h, VX: b.Face * 2.4, Owner: b})
			b.HasBoom = true
			sfx("knife")
		}
		if b.T >= 26 {
			b.St, b.T, b.CD = "approach", 0, 80
		}
	case "jk":
		if b.Y > 0 {
			box := Box{math.Min(b.X, b.X+b.Face*20), math.Max(b.X, b.X+b.Face*20), b.Y + 8, b.Y + 22, 0}
			if ov(box, playerHB()) {
				hitPlayer(9)
			}
		} else if b.T > 4 {
			b.St, b.T = "recover", 0
		}
	case "vanish":
		if b.T >= 24 {
			side := 1.0
			if rnd() < .5 {
				side = -1
			}
			b.X = clamp(P.X+side*(50+rnd()*30), cam+14, cam+VW-14)
			b.Face = sgn(P.X - b.X)
			if b.Face == 0 {
				b.Face = 1
			}
			b.St, b.T = "appear", 0
			burst(b.X, 16, 14, 0xb070ff, 1.5, true, 0)
		}
	case "appear":
		if b.T >= 20 {
			b.St, b.T, b.High = "cast", 0, rnd() < .5
		}
	case "cast":
		faceP()
		if b.T == 22 {
			h := 4.0
			if b.High {
				h = 22
			}
			shots = append(shots, &Shot{K: "fire", X: b.X + b.Face*8, H: h, VX: b.Face * 2})
			sfx("fire")
		}
		if b.T >= 34 {
			b.St, b.T, b.CD = "approach", 0, 50+rnd()*50
		}
	}
	b.T++
	b.X = clamp(b.X, cam+8, cam+VW-8)
}

// Morceau de musique voulu selon l'état du jeu.
func musicFor() string {
	st := state
	if st == "options" && optFrom == "pause" {
		st = "pause"
	}
	switch st {
	case "title", "options", "scores":
		return "title"
	case "intro", "play", "pause":
		if bossZone {
			return "boss"
		}
		return "f" + strconv.Itoa(level)
	case "clear":
		return "clear"
	case "dying":
		return ""
	case "gameover":
		return "over"
	case "win", "name":
		return "win"
	}
	return ""
}
