package main

// Musiques originales, gammes pentatoniques et japonaises.
var tracks = map[string]*Track{}

func init() {
	arp := func(chords map[string]string, seq ...string) string {
		s := ""
		for _, c := range seq {
			s += chords[c] + " | "
		}
		return s
	}

	// Étage 1 : la mineur pentatonique, koto.
	tracks["f1"] = &Track{BPM: 108, SPB: 2, Loop: true, PercVol: .8,
		Chans: []Chan{
			{iKoto, .42, parse("E5 . G5 E5 D5 . C5 D5 | E5 - - - A4 . C5 . | D5 . E5 D5 C5 . A4 G4 | A4 - - - E4 . G4 . | A5 . G5 E5 G5 . E5 D5 | E5 . D5 C5 A4 . C5 . | D5 E5 D5 C5 A4 . G4 E4 | A4 - - - - - - -")},
			{iBass, .5, parse("A2 - - - E3 - - - | A2 - - - E3 - - - | D3 - - - A2 - - - | E3 - - - E2 - - - | A2 - - - E3 - - - | C3 - - - G2 - - - | D3 - - - E3 - - - | A2 - - - A2 - E3 -")},
		},
		Perc: percLoop("t...w.w.t...w..w", 64, 0)}

	// Étage 2 : gamme yo de ré, flûte + arpèges de koto.
	ch2 := map[string]string{
		"D": "D4 A4 D5 A4 E5 A4 D5 A4", "G": "G3 D4 G4 D4 A4 D4 G4 D4",
		"A": "A3 E4 A4 E4 B4 E4 A4 E4", "E": "E4 B4 E5 B4 D5 B4 E5 B4",
	}
	tracks["f2"] = &Track{BPM: 100, SPB: 2, Loop: true, PercVol: .6,
		Chans: []Chan{
			{iFlute, .32, parse("A5 - B5 A5 G5 - E5 G5 | A5 - - - D6 - B5 - | A5 G5 E5 D5 E5 - G5 - | E5 - - - - - - - | D5 - E5 G5 A5 - B5 A5 | G5 - E5 - D5 - B4 - | D5 E5 G5 A5 G5 E5 D5 B4 | D5 - - - - - - -")},
			{iKoto, .22, parse(arp(ch2, "D", "D", "G", "A", "D", "G", "E", "D"))},
			{iBass, .5, parse("D3 - - - - - - - | D3 - - - A2 - - - | G2 - - - - - - - | A2 - - - - - - - | D3 - - - - - - - | G2 - - - - - - - | E2 - - - - - - - | D3 - - - A2 - - -")},
		},
		Perc: percLoop("w.......w...w...", 64, 0)}

	// Étage 3 : gamme miyako-bushi de mi, shamisen.
	tracks["f3"] = &Track{BPM: 116, SPB: 2, Loop: true, PercVol: .8,
		Chans: []Chan{
			{iShami, .4, parse("E5 . F5 . A5 B5 A5 F5 | E5 - - - . . B4 C5 | E5 F5 E5 C5 B4 - A4 - | B4 - - - . . . . | A5 - B5 - C6 B5 A5 F5 | E5 - F5 E5 C5 - B4 - | C5 B4 A4 B4 C5 E5 F5 E5 | E5 - - - . . . .")},
			{iBass, .55, parse("E2 - - - B2 - - - | E2 - - - B2 - - - | A2 - - - E2 - - - | B2 - - - B2 - - - | F2 - - - C3 - - - | E2 - - - B2 - - - | A2 - - - F2 - - - | E2 - - - B2 - E2 -")},
			{iKoto, .16, parse(rep("E4 B4 E5 B4", 16))},
		},
		Perc: percLoop("T...t.w.T.t.w..w", 64, 0)}

	// Étage 4 : hirajoshi de do, flûte inquiétante.
	tracks["f4"] = &Track{BPM: 92, SPB: 2, Loop: true, PercVol: .7,
		Chans: []Chan{
			{iFlute, .34, parse("G5 - - Ab5 G5 - Eb5 - | D5 - - - C5 - - - | Eb5 - D5 C5 G4 - Ab4 - | G4 - - - - - - - | C5 - D5 Eb5 G5 - Ab5 G5 | Eb5 - D5 - C5 - - - | D5 Eb5 D5 C5 Ab4 - G4 - | C5 - - - - - - -")},
			{iKoto, .2, parse(rep("C4 G4 C5 G4 Eb5 G4 C5 G4", 3) + " " + "G3 D4 G4 D4 Ab4 D4 G4 D4 " + rep("C4 G4 C5 G4 Eb5 G4 C5 G4", 3) + " Ab3 Eb4 Ab4 Eb4 C5 Eb4 Ab4 Eb4")},
			{iBass, .55, parse("C2 - - - - - - - | Ab2 - - - - - - - | Eb2 - - - - - - - | G2 - - - - - - - | C2 - - - - - - - | Ab2 - - - - - - - | D2 - - - G2 - - - | C2 - - - - - - -")},
		},
		Perc: percLoop("T.......w.......", 64, 0, 32)}

	// Étage 5 : rapide et intense.
	tracks["f5"] = &Track{BPM: 128, SPB: 2, Loop: true, PercVol: .9,
		Chans: []Chan{
			{iShami, .38, parse("A5 G5 E5 G5 A5 . C6 A5 | G5 E5 D5 E5 G5 - - - | A5 G5 E5 D5 C5 D5 E5 G5 | A5 - - - A4 . A4 . | C6 A5 G5 A5 C6 - D6 C6 | A5 - G5 - E5 - D5 - | E5 G5 A5 C6 A5 G5 E5 D5 | E5 - - - A4 . A4 .")},
			{iBass, .55, parse("A2 . A2 . E3 . A2 . | G2 . G2 . D3 . G2 . | F2 . F2 . C3 . F2 . | E2 . E2 . E3 . E2 . | A2 . A2 . E3 . A2 . | C3 . C3 . G2 . C3 . | D3 . D3 . A2 . D3 . | E3 . E3 . E2 . E2 .")},
			{iFlute, .14, parse("A4 - - - - - - - | G4 - - - - - - - | F4 - - - - - - - | E4 - - - - - - - | A4 - - - - - - - | C5 - - - - - - - | D5 - - - - - - - | E5 - - - - - - -")},
		},
		Perc: percLoop("T.c.Tt.cT.c.TTtc", 64, 0)}

	// Boss : hirajoshi de la, tambours taiko.
	tracks["boss"] = &Track{BPM: 138, SPB: 2, Loop: true, PercVol: 1,
		Chans: []Chan{
			{iShami, .36, parse("A5 - - F5 E5 - C5 B4 | A4 - - - . . E5 F5 | A5 - B5 - C6 B5 A5 F5 | E5 - - - . . . . | F5 - E5 - C5 - B4 C5 | A4 - - - E5 - F5 - | A5 B5 C6 B5 A5 F5 E5 C5 | A4 - - - . . . .")},
			{iKoto, .22, parse(rep("A3 C4 B3 A3 E3 F3 E3 A3", 4) + " " + rep("F3 A3 C4 A3 E4 C4 B3 A3", 2) + " " + rep("A3 C4 B3 A3 E3 F3 E3 A3", 2))},
			{iBass, .6, parse("A2 . . A2 . . A2 . | F2 . . F2 . . E2 . | A2 . . A2 . . A2 . | E2 . . E2 . . E2 . | F2 . . F2 . . F2 . | A2 . . A2 . . A2 . | F2 . . F2 . . E2 . | A2 . . A2 . . E2 .")},
		},
		Perc: percLoop("T..T..t.T..T.tTt", 64, 0)}

	// Titre : flûte lente et arpèges.
	cht := map[string]string{
		"A": "A3 E4 A4 C5 E5 C5 A4 E4", "C": "C4 G4 C5 E5 G5 E5 C5 G4",
		"G": "G3 D4 G4 A4 D5 A4 G4 D4", "D": "D4 A4 D5 E5 A5 E5 D5 A4",
	}
	tracks["title"] = &Track{BPM: 76, SPB: 2, Loop: true, PercVol: .6,
		Chans: []Chan{
			{iFlute, .32, parse("E5 - - - G5 - A5 - | G5 - E5 - D5 - - - | C5 - D5 - E5 - G5 E5 | D5 - - - - - - - | E5 - - - A5 - C6 - | A5 - G5 - E5 - - - | D5 - E5 - C5 - A4 - | A4 - - - - - - -")},
			{iKoto, .2, parse(arp(cht, "A", "G", "C", "G", "A", "C", "D", "A"))},
			{iBass, .4, parse("A2 - - - - - - - | G2 - - - - - - - | C3 - - - - - - - | G2 - - - - - - - | A2 - - - - - - - | C3 - - - - - - - | D3 - - - - - - - | A2 - - - - - - -")},
		},
		Perc: percLoop("................", 64, 0, 32)}

	tracks["clear"] = &Track{BPM: 150, SPB: 2, PercVol: .9,
		Chans: []Chan{
			{iKoto, .45, parse("A5 C6 D6 E6 - G6 - - A6 - - - - - - - . . . .")},
			{iBass, .5, parse("A3 - - - - - - - E3 - - - A2 - - - . . . .")},
		},
		Perc: "T.......T...X......."}

	tracks["over"] = &Track{BPM: 80, SPB: 2, PercVol: .7,
		Chans: []Chan{
			{iFlute, .34, parse("E5 - D5 - C5 - A4 - - - E4 - - - - - . . . .")},
			{iBass, .5, parse("A2 - - - - - - - E2 - - - A2 - - - . . . .")},
		},
		Perc: "g..................."}

	chw := map[string]string{
		"D": "D4 A4 D5 F#5 A5 F#5 D5 A4", "G": "G3 D4 G4 B4 D5 B4 G4 D4",
		"E": "E4 B4 E5 G5 B5 G5 E5 B4", "A": "A3 E4 A4 C#5 E5 C#5 A4 E4",
	}
	tracks["win"] = &Track{BPM: 112, SPB: 2, Loop: true, PercVol: .8,
		Chans: []Chan{
			{iFlute, .32, parse("D5 E5 G5 A5 B5 - A5 B5 | D6 - - - B5 - A5 - | G5 A5 B5 D6 E6 - D6 B5 | D6 - - - - - - - | B5 - A5 - G5 - E5 - | G5 - A5 - B5 - - - | A5 B5 A5 G5 E5 - D5 - | D5 - - - - - - -")},
			{iKoto, .2, parse(arp(chw, "D", "G", "E", "D", "G", "E", "A", "D"))},
			{iBass, .5, parse("D3 - - - - - - - | G2 - - - - - - - | E2 - - - - - - - | D3 - - - - - - - | G2 - - - - - - - | E2 - - - - - - - | A2 - - - - - - - | D3 - - - A2 - - -")},
		},
		Perc: percLoop("T...t.T.T...t.w.", 64, 0)}
}
