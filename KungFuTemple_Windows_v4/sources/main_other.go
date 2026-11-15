//go:build !windows

package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"os"
)

func platFullscreen(bool) {}

func savePNG(fb *FB, name string) {
	img := image.NewRGBA(image.Rect(0, 0, fb.W, fb.H))
	for i, c := range fb.Pix {
		img.Pix[i*4], img.Pix[i*4+1], img.Pix[i*4+2], img.Pix[i*4+3] = uint8(c>>16), uint8(c>>8), uint8(c), 255
	}
	f, _ := os.Create(name)
	png.Encode(f, img)
	f.Close()
}

func writeWAV(name string, s []int16) {
	f, _ := os.Create(name)
	defer f.Close()
	n := len(s) * 2
	w := func(v any) { binary.Write(f, binary.LittleEndian, v) }
	f.WriteString("RIFF")
	w(uint32(36 + n))
	f.WriteString("WAVEfmt ")
	w(uint32(16))
	w(uint16(1))
	w(uint16(1))
	w(uint32(SR))
	w(uint32(SR * 2))
	w(uint16(2))
	w(uint16(16))
	f.WriteString("data")
	w(uint32(n))
	w(s)
}

// Bot de test : avance, frappe, se libère.
func bot(i int) Input {
	in := Input{}
	if state != "play" {
		in.Atk = i%20 < 2
		return in
	}
	fwdLeft := dir < 0
	if isGrabbed() {
		if i%4 < 2 {
			in.Left = true
		} else {
			in.Right = true
		}
		return in
	}
	var near *Ent
	nd := 1e9
	for _, e := range ents {
		if e.Dying {
			continue
		}
		d := e.X - P.X
		if d < 0 {
			d = -d
		}
		if d < nd {
			nd, near = d, e
		}
	}
	for _, s := range shots {
		d := s.X - P.X
		if d < 0 {
			d = -d
		}
		if s.K == "knife" && d < 30 {
			if s.H > 15 {
				in.Down = true
			} else {
				in.Jump = i%3 == 0
			}
		}
	}
	if near != nil && nd < 27 {
		if near.X < P.X {
			in.Left = i%8 == 0
		} else {
			in.Right = i%8 == 0
		}
		if near.Type == "tomtom" && near.Y <= 0 || near.Type == "snake" || (near.Type == "boss" && near.St == "wind" && near.High) {
			in.Down = true
		}
		in.Atk = i%6 < 2
	} else if fwdLeft {
		in.Left = true
	} else {
		in.Right = true
	}
	return in
}

func main() {
	mode := "bot"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	big := NewFB(FW, FH)
	switch mode {
	case "bot":
		cfg.Difficulty = 0
		shots := map[int]string{30: "title", 700: "l1a", 1400: "l1b", 1900: "boss1"}
		seen := map[string]bool{}
		for i := 0; i < 60*60*30; i++ {
			update(bot(i))
			k := state + fmt.Sprint(level)
			if !seen[k] {
				seen[k] = true
				fmt.Println(i, k, "score", score, "lives", lives)
			}
			if n, ok := shots[i]; ok {
				cfg.HD = true
				render(big)
				savePNG(big, "shot_"+n+"_hd.png")
				cfg.HD = false
				render(big)
				savePNG(big, "shot_"+n+"_retro.png")
				cfg.HD = true
			}
			if state == "play" && bossZone && boss != nil && boss.St == "active" && i%7 == 0 && !seen["bshot"+fmt.Sprint(level)] {
				seen["bshot"+fmt.Sprint(level)] = true
				render(big)
				savePNG(big, fmt.Sprintf("shot_boss%d_active.png", level))
			}
			if state == "name" || (state == "title" && i > 100) {
				break
			}
		}
		fmt.Println("fin", state, level, score)
	case "gallery":
		state = "title"
		for i := 0; i < 60; i++ {
			update(Input{})
		}
		render(big)
		savePNG(big, "g_title.png")
		cfg.HD = false
		render(big)
		savePNG(big, "g_title_retro.png")
		cfg.HD = true
		for lv := 1; lv <= 5; lv++ {
			level = lv
			startLevel()
			state = "play"
			for i := 0; i < 500; i++ {
				update(bot(i))
			}
			render(big)
			savePNG(big, fmt.Sprintf("g_lv%d.png", lv))
		}
		cfg.CRT = true
		cfg.HD = false
		render(big)
		savePNG(big, "g_crt.png")
	case "music":
		for _, name := range []string{"f1", "f2", "f3", "f4", "f5", "boss", "title", "clear", "over", "win"} {
			t := tracks[name]
			for ci, c := range t.Chans {
				fmt.Println(name, "chan", ci, "len", len(c.Notes))
			}
			audio.trkName = ""
			audio.SetTrack(name)
			buf := make([]int16, SR*12)
			for o := 0; o < len(buf); o += 1024 {
				e := min(o+1024, len(buf))
				audio.Render(buf[o:e])
			}
			peak, sum := 0, 0.0
			for _, v := range buf {
				a := int(v)
				if a < 0 {
					a = -a
				}
				peak = max(peak, a)
				sum += float64(v) * float64(v)
			}
			fmt.Printf("%s peak %d rms %.0f\n", name, peak, (sum/float64(len(buf)))*0+0)
			writeWAV("m_"+name+".wav", buf)
		}
	}
}
