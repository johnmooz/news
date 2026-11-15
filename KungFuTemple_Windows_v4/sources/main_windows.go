//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	winmm    = syscall.NewLazyDLL("winmm.dll")

	pRegisterClassExW  = user32.NewProc("RegisterClassExW")
	pCreateWindowExW   = user32.NewProc("CreateWindowExW")
	pDefWindowProcW    = user32.NewProc("DefWindowProcW")
	pPeekMessageW      = user32.NewProc("PeekMessageW")
	pTranslateMessage  = user32.NewProc("TranslateMessage")
	pDispatchMessageW  = user32.NewProc("DispatchMessageW")
	pPostQuitMessage   = user32.NewProc("PostQuitMessage")
	pDestroyWindow     = user32.NewProc("DestroyWindow")
	pGetDC             = user32.NewProc("GetDC")
	pGetClientRect     = user32.NewProc("GetClientRect")
	pGetWindowRect     = user32.NewProc("GetWindowRect")
	pAdjustWindowRect  = user32.NewProc("AdjustWindowRect")
	pLoadCursorW       = user32.NewProc("LoadCursorW")
	pSetWindowLongPtrW = user32.NewProc("SetWindowLongPtrW")
	pSetWindowPos      = user32.NewProc("SetWindowPos")
	pMonitorFromWindow = user32.NewProc("MonitorFromWindow")
	pGetMonitorInfoW   = user32.NewProc("GetMonitorInfoW")
	pSetProcessDPIAw   = user32.NewProc("SetProcessDPIAware")
	pMessageBoxW       = user32.NewProc("MessageBoxW")
	pGetSystemMetrics  = user32.NewProc("GetSystemMetrics")
	pValidateRect      = user32.NewProc("ValidateRect")
	pCreateIcon        = user32.NewProc("CreateIcon")
	pSendMessageW      = user32.NewProc("SendMessageW")
	pShowCursor        = user32.NewProc("ShowCursor")

	pStretchDIBits     = gdi32.NewProc("StretchDIBits")
	pPatBlt            = gdi32.NewProc("PatBlt")
	pSetStretchBltMode = gdi32.NewProc("SetStretchBltMode")
	pSetBrushOrgEx     = gdi32.NewProc("SetBrushOrgEx")
	pCreateCompatDC    = gdi32.NewProc("CreateCompatibleDC")
	pCreateDIBSection  = gdi32.NewProc("CreateDIBSection")
	pSelectObject      = gdi32.NewProc("SelectObject")
	pStretchBlt        = gdi32.NewProc("StretchBlt")
	pGdiFlush          = gdi32.NewProc("GdiFlush")
	pBeginPaint        = user32.NewProc("BeginPaint")
	pInvalidateRect    = user32.NewProc("InvalidateRect")
	pUpdateWindow      = user32.NewProc("UpdateWindow")
	pSetWindowTextW    = user32.NewProc("SetWindowTextW")
	pSetDIBitsToDevice = gdi32.NewProc("SetDIBitsToDevice")
	pEndPaint          = user32.NewProc("EndPaint")

	pGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	pVirtualAlloc     = kernel32.NewProc("VirtualAlloc")

	pTimeBeginPeriod      = winmm.NewProc("timeBeginPeriod")
	pWaveOutOpen          = winmm.NewProc("waveOutOpen")
	pWaveOutPrepareHeader = winmm.NewProc("waveOutPrepareHeader")
	pWaveOutWrite         = winmm.NewProc("waveOutWrite")
	pXInputGetState       *syscall.LazyProc
)

const (
	wsOverlappedWindow = 0x00CF0000
	wsVisible          = 0x10000000
	wsPopup            = 0x80000000
	csOwnDC            = 0x20
	csHRedraw          = 0x2
	csVRedraw          = 0x1
	wmDestroy          = 0x0002
	wmPaint            = 0x000F
	wmClose            = 0x0010
	wmEraseBkgnd       = 0x0014
	wmKillFocus        = 0x0008
	wmSetIcon          = 0x0080
	wmKeyDown          = 0x0100
	wmKeyUp            = 0x0101
	wmSysKeyDown       = 0x0104
	wmSysKeyUp         = 0x0105
	wmSysChar          = 0x0106
	wmQuit             = 0x0012
	pmRemove           = 1
	swpNoZOrder        = 0x0004
	swpFrameChanged    = 0x0020
	swpNoOwnerZOrder   = 0x0200
	srcCopy            = 0x00CC0020
	blackness          = 0x00000042
	vkReturn           = 0x0D
	vkEscape           = 0x1B
	vkSpace            = 0x20
	vkLeft             = 0x25
	vkUp               = 0x26
	vkRight            = 0x27
	vkDown             = 0x28
	vkF11              = 0x7A
)

type wndClassExW struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type msgT struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	ptX     int32
	ptY     int32
	priv    uint32
}

type rectT struct{ L, T, R, B int32 }

type monitorInfo struct {
	cbSize    uint32
	rcMonitor rectT
	rcWork    rectT
	dwFlags   uint32
}

type bitmapInfoHeader struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

type bitmapInfo struct {
	h      bitmapInfoHeader
	colors [4]uint32
}

type waveFormatEx struct {
	wFormatTag      uint16
	nChannels       uint16
	nSamplesPerSec  uint32
	nAvgBytesPerSec uint32
	nBlockAlign     uint16
	wBitsPerSample  uint16
	cbSize          uint16
}

type waveHdr struct {
	lpData          uintptr
	dwBufferLength  uint32
	dwBytesRecorded uint32
	dwUser          uintptr
	dwFlags         uint32
	dwLoops         uint32
	lpNext          uintptr
	reserved        uintptr
}

type xinputState struct {
	packet  uint32
	buttons uint16
	lt, rt  uint8
	lx, ly  int16
	rx, ry  int16
}

var (
	hwnd     uintptr
	hdc      uintptr
	keys     [256]bool
	isFS     bool
	prevRect rectT
	padOK    = true
	padWait  int
)

func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

func msgBox(text, title string) {
	pMessageBoxW.Call(0, uintptr(unsafe.Pointer(u16(text))), uintptr(unsafe.Pointer(u16(title))), 0x10)
}

func wndProc(h, msg, wp, lp uintptr) uintptr {
	switch msg {
	case wmClose:
		pDestroyWindow.Call(h)
		return 0
	case wmDestroy:
		pPostQuitMessage.Call(0)
		return 0
	case wmEraseBkgnd:
		return 1
	case wmPaint:
		var ps [80]byte
		dc, _, _ := pBeginPaint.Call(h, uintptr(unsafe.Pointer(&ps[0])))
		if lastFB != nil && dc != 0 {
			present(dc, lastFB)
		}
		pEndPaint.Call(h, uintptr(unsafe.Pointer(&ps[0])))
		return 0
	case wmKillFocus:
		for i := range keys {
			keys[i] = false
		}
		return 0
	case wmKeyDown, wmSysKeyDown:
		vk := wp & 0xFF
		repeat := lp&(1<<30) != 0
		if !repeat && (vk == vkF11 || (msg == wmSysKeyDown && vk == vkReturn)) {
			cfg.Fullscreen = !isFS
			platFullscreen(cfg.Fullscreen)
			writeSave()
			return 0
		}
		if !repeat && vk >= 'A' && vk <= 'Z' {
			cheatKey(rune(vk + 32))
		}
		keys[vk] = true
		if msg == wmSysKeyDown && vk != 0x73 { // laisse Alt+F4
			return 0
		}
	case wmKeyUp, wmSysKeyUp:
		keys[wp&0xFF] = false
		if msg == wmSysKeyUp {
			return 0
		}
	case wmSysChar:
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(h, msg, wp, lp)
	return r
}

func platFullscreen(on bool) {
	if hwnd == 0 || on == isFS {
		return
	}
	gwlStyle := ^uintptr(15) // -16
	if on {
		pGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&prevRect)))
		pSetWindowLongPtrW.Call(hwnd, gwlStyle, wsPopup|wsVisible)
		mon, _, _ := pMonitorFromWindow.Call(hwnd, 2)
		mi := monitorInfo{cbSize: uint32(unsafe.Sizeof(monitorInfo{}))}
		pGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
		r := mi.rcMonitor
		pSetWindowPos.Call(hwnd, 0, uintptr(r.L), uintptr(r.T), uintptr(r.R-r.L), uintptr(r.B-r.T), swpFrameChanged|swpNoOwnerZOrder)
		pShowCursor.Call(0)
	} else {
		pSetWindowLongPtrW.Call(hwnd, gwlStyle, wsOverlappedWindow|wsVisible)
		r := prevRect
		pSetWindowPos.Call(hwnd, 0, uintptr(r.L), uintptr(r.T), uintptr(r.R-r.L), uintptr(r.B-r.T), swpFrameChanged|swpNoZOrder|swpNoOwnerZOrder)
		pShowCursor.Call(1)
	}
	isFS = on
}

func makeIcon(hinst uintptr) uintptr {
	const n = 32
	f := NewFB(n, n)
	f.Circle(16, 16, 15.5, 0xe8b040, 1)
	f.Circle(16, 16, 13.5, 0xb81c18, 1)
	f.Circle(16, 17, 7.5, 0x281810, 1)
	f.Circle(17.5, 18, 6.2, 0xf8b878, 1)
	f.Capsule(8, 15, 24, 15.5, 1.6, 0xffffff, 1)
	f.Capsule(8, 15, 4, 12, 1.2, 0xffffff, 1)
	f.Circle(20, 18, 1, 0x101010, 1)
	xor := make([]byte, n*n*4)
	and := make([]byte, n*n/8)
	for i, c := range f.Pix {
		xor[i*4], xor[i*4+1], xor[i*4+2] = byte(c), byte(c>>8), byte(c>>16)
		if c != 0 {
			xor[i*4+3] = 255
		}
	}
	h, _, _ := pCreateIcon.Call(hinst, n, n, 1, 32, uintptr(unsafe.Pointer(&and[0])), uintptr(unsafe.Pointer(&xor[0])))
	return h
}

func initXInput() {
	for _, n := range []string{"xinput1_4.dll", "xinput1_3.dll", "xinput9_1_0.dll"} {
		d := syscall.NewLazyDLL(n)
		if d.Load() != nil {
			continue
		}
		p := d.NewProc("XInputGetState")
		if p.Find() == nil {
			pXInputGetState = p
			return
		}
	}
}

func readPad() (xinputState, bool) {
	var st xinputState
	if pXInputGetState == nil {
		return st, false
	}
	if !padOK {
		padWait--
		if padWait > 0 {
			return st, false
		}
	}
	r, _, _ := pXInputGetState.Call(0, uintptr(unsafe.Pointer(&st)))
	padOK = r == 0
	if !padOK {
		padWait = 120
	}
	return st, padOK
}

func readInput() Input {
	k := func(v ...int) bool {
		for _, c := range v {
			if keys[c] {
				return true
			}
		}
		return false
	}
	in := Input{
		Left:  k(vkLeft, 'Q', 'A'),
		Right: k(vkRight, 'D'),
		Up:    k(vkUp, 'Z', 'W'),
		Down:  k(vkDown, 'S'),
		Jump:  k(vkUp, 'Z', 'W', 'L', vkSpace),
		Atk:   k('M'),
		Punch: k('K'),
		Start: k(vkReturn),
		Back:  k(vkEscape),
	}
	if st, ok := readPad(); ok {
		b := st.buttons
		in.Up = in.Up || b&0x1 != 0 || st.ly > 16000
		in.Down = in.Down || b&0x2 != 0 || st.ly < -16000
		in.Left = in.Left || b&0x4 != 0 || st.lx < -16000
		in.Right = in.Right || b&0x8 != 0 || st.lx > 16000
		in.Start = in.Start || b&0x10 != 0
		in.Back = in.Back || b&0x20 != 0
		in.Jump = in.Jump || b&0x1000 != 0 || b&0x1 != 0
		in.Atk = in.Atk || b&0x4000 != 0
		in.Punch = in.Punch || b&0x2000 != 0
		if state != "play" && b&0x1000 != 0 {
			in.Atk = true
		}
		if state != "play" && b&0x2000 != 0 {
			in.Punch, in.Back = false, true
		}
	}
	return in
}

func startAudio() {
	defer func() { recover() }()
	if pWaveOutOpen.Find() != nil {
		return
	}
	wfx := waveFormatEx{1, 1, SR, SR * 2, 2, 16, 0}
	var hwo uintptr
	r, _, _ := pWaveOutOpen.Call(uintptr(unsafe.Pointer(&hwo)), 0xFFFFFFFF, uintptr(unsafe.Pointer(&wfx)), 0, 0, 0)
	if r != 0 {
		return
	}
	const nb, ns = 4, 768
	hs := unsafe.Sizeof(waveHdr{})
	mem, _, _ := pVirtualAlloc.Call(0, uintptr(nb)*(hs+ns*2), 0x3000, 0x04)
	if mem == 0 {
		return
	}
	hdrs := make([]*waveHdr, nb)
	bufs := make([][]int16, nb)
	for i := 0; i < nb; i++ {
		h := (*waveHdr)(unsafe.Pointer(mem + uintptr(i)*hs))
		data := mem + uintptr(nb)*hs + uintptr(i)*ns*2
		h.lpData, h.dwBufferLength = data, ns*2
		hdrs[i] = h
		bufs[i] = unsafe.Slice((*int16)(unsafe.Pointer(data)), ns)
		pWaveOutPrepareHeader.Call(hwo, uintptr(unsafe.Pointer(h)), hs)
		audio.Render(bufs[i])
		pWaveOutWrite.Call(hwo, uintptr(unsafe.Pointer(h)), hs)
	}
	go func() {
		defer func() { recover() }()
		for {
			for i, h := range hdrs {
				fl := atomic.LoadUint32(&h.dwFlags)
				if fl&1 != 0 {
					audio.Render(bufs[i])
					atomic.StoreUint32(&h.dwFlags, fl&^1)
					pWaveOutWrite.Call(hwo, uintptr(unsafe.Pointer(h)), hs)
				}
			}
			time.Sleep(2 * time.Millisecond)
		}
	}()
}

var bmi = bitmapInfo{h: bitmapInfoHeader{biSize: uint32(unsafe.Sizeof(bitmapInfoHeader{})), biWidth: FW, biHeight: -FH, biPlanes: 1, biBitCount: 32}}

var (
	memDC   uintptr
	dibPix  []uint32
	lastFB  *FB
	logN    int
	useDIB  = true
	failCnt int
)

// dlog écrit les premiers événements d'affichage dans un journal de diagnostic.
func dlog(format string, a ...any) {
	if logN > 40 {
		return
	}
	logN++
	p := filepath.Join(filepath.Dir(savePath()), "affichage_log.txt")
	os.MkdirAll(filepath.Dir(p), 0o755)
	fl := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if logN == 1 {
		fl = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}
	if f, err := os.OpenFile(p, fl, 0o644); err == nil {
		fmt.Fprintf(f, format+"\r\n", a...)
		f.Close()
	}
}

func initDIB() {
	dc, _, e1 := pCreateCompatDC.Call(hdc)
	var bits uintptr
	dib, _, e2 := pCreateDIBSection.Call(hdc, uintptr(unsafe.Pointer(&bmi)), 0, uintptr(unsafe.Pointer(&bits)), 0, 0)
	dlog("hdc=%x memDC=%x (%v) dib=%x bits=%x (%v)", hdc, dc, e1, dib, bits, e2)
	if dc == 0 || dib == 0 || bits == 0 {
		useDIB = false
		return
	}
	pSelectObject.Call(dc, dib)
	memDC = dc
	dibPix = unsafe.Slice((*uint32)(unsafe.Pointer(bits)), FW*FH)
}

var (
	method   int // 0 StretchDIBits, 1 SetDIBitsToDevice, 2 section DIB
	lastR    uintptr
	frames   int
	captured bool
)

// saveBMP enregistre l'image calculée par le jeu (diagnostic).
func saveBMP(fb *FB, p string) {
	w, h := fb.W, fb.H
	row := w * 3
	data := make([]byte, 54+row*h)
	le := func(o int, v uint32) {
		data[o], data[o+1], data[o+2], data[o+3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	}
	data[0], data[1] = 'B', 'M'
	le(2, uint32(len(data)))
	le(10, 54)
	le(14, 40)
	le(18, uint32(w))
	le(22, uint32(h))
	data[26], data[28] = 1, 24
	le(34, uint32(row*h))
	for y := 0; y < h; y++ {
		o := 54 + (h-1-y)*row
		for x := 0; x < w; x++ {
			c := fb.Pix[y*w+x]
			data[o+x*3], data[o+x*3+1], data[o+x*3+2] = byte(c), byte(c>>8), byte(c>>16)
		}
	}
	os.WriteFile(p, data, 0o644)
}

func present(dc uintptr, fb *FB) {
	var rc rectT
	pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	cw, chh := int(rc.R), int(rc.B)
	if cw <= 0 || chh <= 0 {
		return
	}
	dw, dh := cw, cw*FH/FW
	if dh > chh {
		dh, dw = chh, chh*FW/FH
	}
	ox, oy := (cw-dw)/2, (chh-dh)/2
	if ox > 0 {
		pPatBlt.Call(dc, 0, 0, uintptr(ox), uintptr(chh), blackness)
		pPatBlt.Call(dc, uintptr(ox+dw), 0, uintptr(cw-ox-dw), uintptr(chh), blackness)
	}
	if oy > 0 {
		pPatBlt.Call(dc, 0, 0, uintptr(cw), uintptr(oy), blackness)
		pPatBlt.Call(dc, 0, uintptr(oy+dh), uintptr(cw), uintptr(chh-oy-dh), blackness)
	}
	pSetStretchBltMode.Call(dc, 3)
	var r uintptr
	var err error
	switch method {
	case 0:
		r, _, err = pStretchDIBits.Call(dc, uintptr(ox), uintptr(oy), uintptr(dw), uintptr(dh), 0, 0, FW, FH,
			uintptr(unsafe.Pointer(&fb.Pix[0])), uintptr(unsafe.Pointer(&bmi)), 0, srcCopy)
	case 1:
		r, _, err = pSetDIBitsToDevice.Call(dc, uintptr(ox), uintptr(oy), FW, FH, 0, 0, 0, FH,
			uintptr(unsafe.Pointer(&fb.Pix[0])), uintptr(unsafe.Pointer(&bmi)), 0)
	case 2:
		if memDC == 0 {
			r = 0
			break
		}
		for i, c := range fb.Pix {
			dibPix[i] = c | 0xFF000000
		}
		pGdiFlush.Call()
		r, _, err = pStretchBlt.Call(dc, uintptr(ox), uintptr(oy), uintptr(dw), uintptr(dh), memDC, 0, 0, FW, FH, srcCopy)
	}
	lastR = r
	if logN < 12 {
		nz := 0
		for i := 0; i < len(fb.Pix); i += 97 {
			if fb.Pix[i]&0xFFFFFF != 0 {
				nz++
			}
		}
		dlog("present methode=%d dc=%x client=%dx%d dest=%d,%d %dx%d r=%d err=%v pixels_non_noirs=%d/%d", method, dc, cw, chh, ox, oy, dw, dh, r, err, nz, len(fb.Pix)/97)
	}
	if r == 0 {
		failCnt++
		if failCnt >= 3 {
			failCnt = 0
			method = (method + 1) % 3
			dlog("échec répété, passage à la méthode %d", method)
		}
	} else {
		failCnt = 0
	}
}

func main() {
	runtime.LockOSThread()
	defer func() {
		if r := recover(); r != nil {
			txt := fmt.Sprintf("%v\n\n%s", r, debug.Stack())
			p := filepath.Join(filepath.Dir(savePath()), "crash.txt")
			os.MkdirAll(filepath.Dir(p), 0o755)
			os.WriteFile(p, []byte(txt), 0o644)
			msgBox("Le jeu a rencontré une erreur.\nDétails enregistrés dans :\n"+p, "Kung-Fu Temple")
		}
	}()
	loadSave()
	if pSetProcessDPIAw.Find() == nil {
		pSetProcessDPIAw.Call()
	}
	if pTimeBeginPeriod.Find() == nil {
		pTimeBeginPeriod.Call(1)
	}
	hinst, _, _ := pGetModuleHandleW.Call(0)
	cursor, _, _ := pLoadCursorW.Call(0, 32512)
	icon := makeIcon(hinst)
	cls := u16("KungFuTempleWnd")
	wc := wndClassExW{
		style: csHRedraw | csVRedraw, lpfnWndProc: syscall.NewCallback(wndProc),
		hInstance: hinst, hIcon: icon, hCursor: cursor, lpszClassName: cls, hIconSm: icon,
	}
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	if a, _, _ := pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); a == 0 {
		msgBox("Impossible de créer la fenêtre.", "Kung-Fu Temple")
		return
	}
	sw, _, _ := pGetSystemMetrics.Call(0)
	sh, _, _ := pGetSystemMetrics.Call(1)
	ch := min(FH, int(sh)*82/100)
	cw := ch * FW / FH
	rc := rectT{0, 0, int32(cw), int32(ch)}
	pAdjustWindowRect.Call(uintptr(unsafe.Pointer(&rc)), wsOverlappedWindow, 0)
	w, h := int(rc.R-rc.L), int(rc.B-rc.T)
	x, y := (int(sw)-w)/2, (int(sh)-h)/2
	hwnd, _, _ = pCreateWindowExW.Call(0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(u16("Kung-Fu Temple"))),
		wsOverlappedWindow|wsVisible, uintptr(max(x, 0)), uintptr(max(y, 0)), uintptr(w), uintptr(h), 0, 0, hinst, 0)
	if hwnd == 0 {
		msgBox("Impossible de créer la fenêtre.", "Kung-Fu Temple")
		return
	}
	pSendMessageW.Call(hwnd, wmSetIcon, 1, icon)
	pSendMessageW.Call(hwnd, wmSetIcon, 0, icon)
	hdc, _, _ = pGetDC.Call(hwnd)
	dlog("démarrage hwnd=%x hdc=%x", hwnd, hdc)
	initDIB()
	if cfg.Fullscreen {
		platFullscreen(true)
	}
	initXInput()
	audio.SetVolumes(cfg.Music, cfg.Sfx)
	startAudio()

	fb := NewFB(FW, FH)
	step := time.Second / 60
	next := time.Now()
	var m msgT
	for {
		for {
			r, _, _ := pPeekMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0, pmRemove)
			if r == 0 {
				break
			}
			if m.message == wmQuit {
				writeSave()
				return
			}
			pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
			pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
		}
		now := time.Now()
		n := 0
		for !now.Before(next) && n < 5 {
			update(readInput())
			audio.SetVolumes(cfg.Music, cfg.Sfx)
			audio.SetTrack(musicFor())
			next = next.Add(step)
			n++
		}
		if n == 5 {
			next = now.Add(step)
		}
		if quitReq {
			writeSave()
			pDestroyWindow.Call(hwnd)
			quitReq = false
		}
		if n > 0 {
			render(fb)
			lastFB = fb
			pInvalidateRect.Call(hwnd, 0, 0)
			pUpdateWindow.Call(hwnd)
			frames++
			if !captured && frames == 90 {
				captured = true
				saveBMP(fb, filepath.Join(filepath.Dir(savePath()), "capture.bmp"))
				dlog("capture.bmp enregistrée")
			}
			if frames%60 == 0 {
				t := fmt.Sprintf("Kung-Fu Temple   [affichage : méthode %d, résultat %d]", method, lastR)
				pSetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(u16(t))))
			}
		}
		if d := time.Until(next); d > 0 {
			time.Sleep(d)
		}
	}
}
