package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

type Pal struct {
	Skin, Hair, Top, Pants, Belt, Shoe, Band, Beard uint32
	Bald, HasBand, HasBeard                         bool
}

var pals = map[string]Pal{
	"hero":      {Skin: 0xf8b878, Hair: 0x281810, Top: 0xf4f4f0, Pants: 0xf4f4f0, Belt: 0x181818, Shoe: 0x181818, Band: 0xe02828, HasBand: true},
	"gripper":   {Skin: 0xe8a060, Hair: 0xe8a060, Bald: true, Top: 0xb048d0, Pants: 0xf0e8d0, Belt: 0x602878, Shoe: 0x603010},
	"knifer":    {Skin: 0xf0b070, Hair: 0x181818, Top: 0x2858b8, Pants: 0x2858b8, Belt: 0xf0d030, Shoe: 0x181818},
	"tomtom":    {Skin: 0xf0b070, Hair: 0x181818, Top: 0xd83030, Pants: 0xf0e8d0, Belt: 0x181818, Shoe: 0x181818},
	"baton":     {Skin: 0xf0b070, Hair: 0x181818, Top: 0x309838, Pants: 0xf0e8d0, Belt: 0x185020, Shoe: 0x402010, Band: 0xf0d030, HasBand: true},
	"boomerang": {Skin: 0xe8a060, Hair: 0xc05818, Top: 0xe8e8e8, Pants: 0x3848c0, Belt: 0xc02020, Shoe: 0x181818},
	"colosse":   {Skin: 0xd88850, Hair: 0xd88850, Bald: true, Top: 0x782888, Pants: 0x282828, Belt: 0xf0b030, Shoe: 0x282828, Beard: 0x382010, HasBeard: true},
	"sorcier":   {Skin: 0xe8b890, Hair: 0xf0f0f0, Top: 0x382858, Pants: 0x382858, Belt: 0xc03030, Shoe: 0x101010, Beard: 0xf0f0f0, HasBeard: true},
	"maitre":    {Skin: 0xf0b070, Hair: 0x181818, Top: 0xc82020, Pants: 0x181818, Belt: 0xf0d030, Shoe: 0x181818, Band: 0xf0d030, HasBand: true},
}

type BossSpec struct {
	Name                    string
	HP                      int
	Speed, Reach            float64
	Wind, Active, Rec, Back int
	CD, Dmg                 float64
	Pts                     int
	Ranged, Magic           bool
	Armor                   int
	Scale, Block            float64
}

var bosses = map[string]BossSpec{
	"baton":     {Name: "LE MAÎTRE DU BÂTON", HP: 20, Speed: .9, Reach: 36, Wind: 18, Active: 8, Rec: 16, Back: 24, CD: 40, Dmg: 10, Pts: 2000, Scale: 1},
	"boomerang": {Name: "LE LANCEUR DE BOOMERANG", HP: 22, Speed: .8, Reach: 22, Wind: 14, Active: 7, Rec: 14, Back: 30, CD: 50, Dmg: 9, Pts: 3000, Ranged: true, Scale: 1},
	"colosse":   {Name: "LE COLOSSE", HP: 28, Speed: .55, Reach: 28, Wind: 24, Active: 8, Rec: 22, Back: 16, CD: 45, Dmg: 14, Pts: 4000, Armor: 3, Scale: 1.35},
	"sorcier":   {Name: "LE SORCIER", HP: 22, Speed: .5, Reach: 22, Wind: 14, Active: 6, Rec: 14, Back: 20, CD: 60, Dmg: 10, Pts: 5000, Magic: true, Scale: 1},
	"maitre":    {Name: "LE GRAND MAÎTRE", HP: 32, Speed: 1.5, Reach: 24, Wind: 10, Active: 6, Rec: 10, Back: 14, CD: 25, Dmg: 9, Pts: 8000, Block: .3, Scale: 1},
}

type WT struct {
	Name string
	W    int
}
type LevelCfg struct {
	Dir, Rate float64
	Max       int
	W         []WT
	Vases     bool
	Boss      string
}

var levels = []LevelCfg{{},
	{Dir: -1, Rate: 70, Max: 5, W: []WT{{"gripper", 7}, {"knifer", 1}}, Boss: "baton"},
	{Dir: 1, Rate: 64, Max: 5, W: []WT{{"gripper", 5}, {"knifer", 1}, {"tomtom", 2}}, Vases: true, Boss: "boomerang"},
	{Dir: -1, Rate: 58, Max: 6, W: []WT{{"gripper", 5}, {"knifer", 2}, {"tomtom", 2}}, Boss: "colosse"},
	{Dir: 1, Rate: 52, Max: 6, W: []WT{{"gripper", 4}, {"knifer", 2}, {"tomtom", 3}}, Vases: true, Boss: "sorcier"},
	{Dir: -1, Rate: 46, Max: 7, W: []WT{{"gripper", 5}, {"knifer", 3}, {"tomtom", 3}}, Vases: true, Boss: "maitre"},
}

/* ---------- Réglages et meilleurs scores ---------- */

type Settings struct {
	Difficulty int
	HD, CRT    bool
	Music, Sfx int
	Fullscreen bool
}
type HS struct {
	Name         string
	Score, Level int
}

var cfg = Settings{Difficulty: 1, HD: true, Music: 7, Sfx: 8}
var hiscores = []HS{{"LEE", 60000, 5}, {"KIM", 40000, 4}, {"WU", 25000, 3}, {"TAO", 12000, 2}, {"MEI", 5000, 1}}

func savePath() string {
	d, err := os.UserConfigDir()
	if err != nil {
		d = "."
	}
	return filepath.Join(d, "KungFuTemple", "save.json")
}

func loadSave() {
	b, err := os.ReadFile(savePath())
	if err != nil {
		return
	}
	var s struct {
		Settings Settings
		Scores   []HS
	}
	if json.Unmarshal(b, &s) == nil {
		if s.Settings.Music > 0 || s.Settings.Sfx > 0 || s.Settings.HD || s.Settings.Difficulty > 0 {
			cfg = s.Settings
		}
		if cfg.Difficulty < 0 || cfg.Difficulty > 2 {
			cfg.Difficulty = 1
		}
		if len(s.Scores) > 0 {
			hiscores = s.Scores
		}
	}
}

func writeSave() {
	p := savePath()
	os.MkdirAll(filepath.Dir(p), 0o755)
	b, _ := json.MarshalIndent(struct {
		Settings Settings
		Scores   []HS
	}{cfg, hiscores}, "", "  ")
	os.WriteFile(p, b, 0o644)
}

func qualifies(sc int) bool {
	return sc > 0 && (len(hiscores) < 5 || sc > hiscores[len(hiscores)-1].Score)
}

func insertScore(name string, sc, lv int) {
	hiscores = append(hiscores, HS{name, sc, lv})
	sort.SliceStable(hiscores, func(i, j int) bool { return hiscores[i].Score > hiscores[j].Score })
	if len(hiscores) > 5 {
		hiscores = hiscores[:5]
	}
	writeSave()
}

func topScore() int {
	if len(hiscores) == 0 {
		return 0
	}
	return hiscores[0].Score
}
