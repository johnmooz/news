package main

import (
	"math"
	"strings"
	"sync"
)

const SR = 44100

const (
	iKoto = iota
	iShami
	iBass
	iFlute
	iTaiko
	iTaikoS
	iWood
	iGong
	iCym
	iSq
	iSaw
	iTri
	iSine
	iNoise
)

type voice struct {
	inst               int
	f, f2, t, dur, vol float64
	music, done        bool
	ks                 []float64
	ki                 int
	ph, lp, lp2        float64
}

type nt struct{ m, dur int }

type Chan struct {
	Inst  int
	Vol   float64
	Notes []nt
}

type Track struct {
	BPM     float64
	SPB     int
	Loop    bool
	Chans   []Chan
	Perc    string
	PercVol float64
}

func (t *Track) length() int { return len(t.Chans[0].Notes) }
func (t *Track) stepSec() float64 {
	return 60 / t.BPM / float64(t.SPB)
}

type Synth struct {
	mu             sync.Mutex
	vs             []*voice
	trk            *Track
	trkName        string
	step           int
	acc            float64
	MusVol, SfxVol float64
	seed           uint32
}

var audio = &Synth{seed: 2463534242, MusVol: 7, SfxVol: 8}

func mf(m float64) float64 { return 440 * math.Pow(2, (m-69)/12) }

func (s *Synth) rnd() float64 {
	s.seed ^= s.seed << 13
	s.seed ^= s.seed >> 17
	s.seed ^= s.seed << 5
	return float64(s.seed)/4294967296.0*2 - 1
}

// add ajoute une voix (doit être appelé sous verrou).
func (s *Synth) add(inst int, f, dur, vol float64, music bool, delay, f2 float64) {
	if len(s.vs) > 64 {
		return
	}
	s.vs = append(s.vs, &voice{inst: inst, f: f, f2: f2, t: -delay, dur: dur, vol: vol, music: music})
}

func (s *Synth) sample(v *voice) float64 {
	const dt = 1.0 / SR
	if v.t < 0 {
		v.t += dt
		return 0
	}
	t := v.t
	var o float64
	switch v.inst {
	case iKoto, iShami, iBass:
		if v.ks == nil {
			n := int(SR / v.f)
			if n < 2 {
				n = 2
			}
			v.ks = make([]float64, n)
			for i := range v.ks {
				v.ks[i] = s.rnd()
			}
			passes := map[int]int{iKoto: 2, iShami: 0, iBass: 5}[v.inst]
			for p := 0; p < passes; p++ {
				for i := range v.ks {
					v.ks[i] = (v.ks[i] + v.ks[(i+1)%n]) * .5
				}
			}
		}
		n := len(v.ks)
		j := (v.ki + 1) % n
		o = v.ks[v.ki]
		damp := .996
		if v.inst == iShami {
			damp = .991
		} else if v.inst == iBass {
			damp = .997
		}
		v.ks[v.ki] = (v.ks[v.ki] + v.ks[j]) * .5 * damp
		v.ki = j
		if v.inst == iShami {
			o *= 1.3
		}
		if r := v.dur - t; r < .06 {
			o *= r / .06
		}
	case iFlute:
		vib := 0.0
		if t > .15 {
			vib = .007 * math.Sin(2*math.Pi*5.3*t) * math.Min(1, (t-.15)*4)
		}
		v.ph += 2 * math.Pi * v.f * (1 + vib) * dt
		env := math.Min(1, t/.06)
		if r := v.dur - t; r < .12 {
			env *= math.Max(0, r/.12)
		}
		n := s.rnd()
		v.lp += (n - v.lp) * .12
		o = (math.Sin(v.ph)*.8+.22*math.Sin(2*v.ph)+.06*math.Sin(3*v.ph))*env + v.lp*(.5*math.Exp(-t*10)+.05*env)
	case iTaiko:
		f := 52 + 115*math.Exp(-t*14)
		v.ph += 2 * math.Pi * f * dt
		n := s.rnd()
		v.lp += (n - v.lp) * .08
		o = math.Sin(v.ph)*math.Exp(-t*6)*1.3 + v.lp*math.Exp(-t*30)*1.2
	case iTaikoS:
		f := 150 + 230*math.Exp(-t*20)
		v.ph += 2 * math.Pi * f * dt
		n := s.rnd()
		v.lp += (n - v.lp) * .25
		o = math.Sin(v.ph)*math.Exp(-t*15) + v.lp*math.Exp(-t*45)*.8
	case iWood:
		f := v.f * (1 + .25*math.Exp(-t*80))
		v.ph += 2 * math.Pi * f * dt
		o = math.Sin(v.ph) * math.Exp(-t*55)
	case iGong:
		ks := [5]float64{1, 1.48, 1.99, 2.63, 3.31}
		as := [5]float64{1, .6, .45, .3, .2}
		for i, k := range ks {
			o += math.Sin(2*math.Pi*v.f*k*t+float64(i)) * as[i] * math.Exp(-t*(.6+float64(i)*.55))
		}
		o *= math.Min(1, t/.008) * .55
		n := s.rnd()
		v.lp += (n - v.lp) * .2
		o += v.lp * math.Exp(-t*12) * .4
	case iCym:
		n := s.rnd()
		v.lp += (n - v.lp) * .35
		o = (n - v.lp) * math.Exp(-t*16) * .7
	case iSq, iSaw, iTri, iSine:
		f := v.f
		if v.f2 > 0 {
			f = v.f * math.Pow(v.f2/v.f, math.Min(1, t/v.dur))
		}
		v.ph += f * dt
		p := v.ph - math.Floor(v.ph)
		switch v.inst {
		case iSq:
			if p < .5 {
				o = .6
			} else {
				o = -.6
			}
		case iSaw:
			o = (p*2 - 1) * .6
		case iTri:
			o = 1 - 4*math.Abs(p-.5)
		default:
			o = math.Sin(2 * math.Pi * p)
		}
		o *= 1 - t/v.dur
	case iNoise:
		n := s.rnd()
		v.lp += (n - v.lp) * v.f
		o = v.lp * (1 - t/v.dur) * 1.5
	}
	v.t += dt
	if v.t >= v.dur {
		v.done = true
	}
	return o * v.vol
}

func (s *Synth) doStep() {
	t := s.trk
	ss := t.stepSec()
	for _, c := range t.Chans {
		n := c.Notes[s.step%len(c.Notes)]
		if n.m <= 0 {
			continue
		}
		d := float64(n.dur) * ss * .95
		switch c.Inst {
		case iKoto, iShami:
			d = math.Max(d, 1.2)
		case iBass:
			d = math.Max(d, .8)
		}
		s.add(c.Inst, mf(float64(n.m)), d, c.Vol, true, 0, 0)
	}
	if t.Perc != "" {
		pv := t.PercVol
		switch t.Perc[s.step%len(t.Perc)] {
		case 'T':
			s.add(iTaiko, 0, .7, .9*pv, true, 0, 0)
		case 't':
			s.add(iTaikoS, 0, .35, .55*pv, true, 0, 0)
		case 'w':
			s.add(iWood, 1150, .12, .35*pv, true, 0, 0)
		case 'W':
			s.add(iWood, 820, .12, .35*pv, true, 0, 0)
		case 'c':
			s.add(iCym, 0, .3, .25*pv, true, 0, 0)
		case 'g':
			s.add(iGong, 98, 4, .55*pv, true, 0, 0)
		case 'X':
			s.add(iTaiko, 0, .7, .9*pv, true, 0, 0)
			s.add(iCym, 0, .3, .25*pv, true, 0, 0)
		}
	}
	s.step++
	if s.step >= t.length() {
		if t.Loop {
			s.step = 0
		} else {
			s.trk = nil
		}
	}
}

// Render remplit un tampon mono 16 bits.
func (s *Synth) Render(out []int16) {
	s.mu.Lock()
	defer s.mu.Unlock()
	mv := math.Pow(s.MusVol/10, 1.6) * .8
	sv := math.Pow(s.SfxVol/10, 1.6)
	for i := range out {
		if s.trk != nil {
			s.acc--
			if s.acc <= 0 {
				s.doStep()
				if s.trk != nil {
					s.acc += s.trk.stepSec() * SR
				}
			}
		}
		mix := 0.0
		for _, v := range s.vs {
			if v.done {
				continue
			}
			x := s.sample(v)
			if v.music {
				mix += x * mv
			} else {
				mix += x * sv
			}
		}
		mix = math.Tanh(mix * .8)
		out[i] = int16(mix * 30000)
	}
	k := s.vs[:0]
	for _, v := range s.vs {
		if !v.done {
			k = append(k, v)
		}
	}
	for i := len(k); i < len(s.vs); i++ {
		s.vs[i] = nil
	}
	s.vs = k
}

func (s *Synth) SetTrack(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if name == s.trkName {
		return
	}
	s.trkName = name
	s.trk = tracks[name]
	s.step, s.acc = 0, SR*.05
}

func (s *Synth) SetVolumes(m, x int) {
	s.mu.Lock()
	s.MusVol, s.SfxVol = float64(m), float64(x)
	s.mu.Unlock()
}

func sfx(n string) {
	s := audio
	s.mu.Lock()
	defer s.mu.Unlock()
	a := s.add
	switch n {
	case "swing":
		a(iNoise, .22, .08, .22, false, 0, 0)
		a(iSine, 520, .07, .12, false, 0, 180)
	case "hit":
		a(iNoise, .5, .1, .45, false, 0, 0)
		a(iSq, 170, .1, .18, false, 0, 60)
		a(iTaikoS, 0, .3, .5, false, 0, 0)
	case "jump":
		a(iSq, 260, .16, .09, false, 0, 620)
	case "land":
		a(iTaikoS, 0, .2, .18, false, 0, 0)
	case "hurt":
		a(iSaw, 230, .28, .25, false, 0, 70)
		a(iNoise, .3, .12, .25, false, 0, 0)
	case "grab":
		a(iSq, 140, .12, .14, false, 0, 110)
	case "free":
		a(iSq, 400, .16, .16, false, 0, 950)
	case "knife":
		a(iTri, 1500, .08, .16, false, 0, 900)
		a(iNoise, .6, .05, .1, false, 0, 0)
	case "fire":
		a(iNoise, .12, .3, .35, false, 0, 0)
		a(iSaw, 300, .22, .12, false, 0, 110)
	case "block":
		a(iWood, 1700, .1, .4, false, 0, 0)
		a(iSq, 1000, .05, .1, false, 0, 1300)
	case "die":
		for i, m := range []float64{76, 74, 72, 69, 64, 57} {
			a(iKoto, mf(m), 1.2, .45, false, float64(i)*.2, 0)
		}
	case "bossdown":
		a(iGong, 87, 4, .7, false, 0, 0)
		a(iTaiko, 0, .8, .9, false, 0, 0)
		a(iTaiko, 0, .8, .7, false, .25, 0)
	case "start":
		for i, m := range []float64{69, 72, 76, 81} {
			a(iKoto, mf(m), 1, .4, false, float64(i)*.09, 0)
		}
	case "boss":
		a(iTaiko, 0, .8, .9, false, 0, 0)
		a(iTaiko, 0, .8, .9, false, .22, 0)
		a(iTaiko, 0, .8, .9, false, .44, 0)
		a(iGong, 73, 4, .6, false, .66, 0)
	case "life":
		for i, m := range []float64{84, 88, 91, 96} {
			a(iSq, mf(m), .1, .1, false, float64(i)*.07, 0)
		}
	case "menu":
		a(iWood, 1600, .08, .3, false, 0, 0)
	case "select":
		a(iKoto, mf(81), .6, .4, false, 0, 0)
		a(iWood, 1100, .08, .25, false, 0, 0)
	case "tick":
		a(iSq, 880, .03, .05, false, 0, 0)
	case "vase":
		a(iNoise, .4, .2, .35, false, 0, 0)
		a(iWood, 700, .1, .3, false, 0, 0)
		a(iWood, 950, .1, .25, false, .04, 0)
	case "cheat":
		for i, m := range []float64{72, 76, 79, 84, 88} {
			a(iSq, mf(m), .12, .1, false, float64(i)*.07, 0)
		}
		a(iTri, mf(96), .45, .15, false, .38, 0)
	case "gun":
		a(iNoise, .55, .12, .5, false, 0, 0)
		a(iSq, 900, .08, .12, false, 0, 150)
	case "riot":
		a(iNoise, .6, .05, .35, false, 0, 0)
		a(iSq, 160, .05, .1, false, 0, 80)
	case "boom":
		a(iNoise, .08, .7, .7, false, 0, 0)
		a(iTaiko, 0, .9, 1, false, 0, 0)
		a(iSaw, 90, .6, .2, false, 0, 30)
	case "magic":
		a(iSine, 400, .35, .14, false, 0, 1300)
		a(iSine, 600, .35, .1, false, .05, 1900)
	}
}

/* ---------- Partitions ---------- */

var noteIdx = map[string]int{"C": 0, "D": 2, "E": 4, "F": 5, "G": 7, "A": 9, "B": 11}

// parse transforme "E5 . G5 - | ..." en pas : "." silence, "-" tenue.
func parse(src string) []nt {
	var out []nt
	for _, tok := range strings.Fields(src) {
		switch tok {
		case "|":
			continue
		case ".":
			out = append(out, nt{0, 1})
		case "-":
			out = append(out, nt{-1, 0})
			for i := len(out) - 2; i >= 0; i-- {
				if out[i].m != -1 {
					out[i].dur++
					break
				}
			}
		default:
			sem := noteIdx[tok[:1]]
			rest := tok[1:]
			if strings.HasPrefix(rest, "#") {
				sem++
				rest = rest[1:]
			} else if strings.HasPrefix(rest, "b") {
				sem--
				rest = rest[1:]
			}
			oct := int(rest[0] - '0')
			out = append(out, nt{(oct+1)*12 + sem, 1})
		}
	}
	return out
}

func rep(s string, n int) string { return strings.TrimSpace(strings.Repeat(s+" ", n)) }

// perc construit un motif de percussion avec gong en tête de boucle.
func percLoop(pat string, total int, gongAt ...int) string {
	b := []byte(strings.Repeat(pat, total/len(pat)+1)[:total])
	for _, g := range gongAt {
		b[g] = 'g'
	}
	return string(b)
}
