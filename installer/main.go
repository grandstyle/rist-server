package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

//go:embed RIST-Server.exe
var embeddedServerExe []byte

//go:embed app.ico
var embeddedAppIcon []byte

var (
	user32  = syscall.NewLazyDLL("user32.dll")
	gdi32   = syscall.NewLazyDLL("gdi32.dll")
	shell32 = syscall.NewLazyDLL("shell32.dll")
	dwmapi  = syscall.NewLazyDLL("dwmapi.dll")
	ole32   = syscall.NewLazyDLL("ole32.dll")

	procRegisterClassExW       = user32.NewProc("RegisterClassExW")
	procCreateWindowExW        = user32.NewProc("CreateWindowExW")
	procDefWindowProcW         = user32.NewProc("DefWindowProcW")
	procGetMessageW            = user32.NewProc("GetMessageW")
	procTranslateMessage       = user32.NewProc("TranslateMessage")
	procDispatchMessageW       = user32.NewProc("DispatchMessageW")
	procPostQuitMessage        = user32.NewProc("PostQuitMessage")
	procSendMessageW           = user32.NewProc("SendMessageW")
	procBeginPaint             = user32.NewProc("BeginPaint")
	procEndPaint               = user32.NewProc("EndPaint")
	procInvalidateRect         = user32.NewProc("InvalidateRect")
	procGetClientRect          = user32.NewProc("GetClientRect")
	procLoadCursorW            = user32.NewProc("LoadCursorW")
	procCreateIconFromResource = user32.NewProc("CreateIconFromResourceEx")
	procDrawTextW              = user32.NewProc("DrawTextW")
	procFillRect               = user32.NewProc("FillRect")
	procEnableWindow           = user32.NewProc("EnableWindow")
	procMessageBoxW            = user32.NewProc("MessageBoxW")
	procShellExecuteW          = shell32.NewProc("ShellExecuteW")
	procDwmSetWindowAttribute  = dwmapi.NewProc("DwmSetWindowAttribute")
	procGetWindowTextW         = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW   = user32.NewProc("GetWindowTextLengthW")
	procSetWindowTextW         = user32.NewProc("SetWindowTextW")
	procSetCursor              = user32.NewProc("SetCursor")

	procCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	procCreatePen              = gdi32.NewProc("CreatePen")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procSetBkColor             = gdi32.NewProc("SetBkColor")
	procSetTextColor           = gdi32.NewProc("SetTextColor")
	procSetBkMode              = gdi32.NewProc("SetBkMode")
	procCreateFontW            = gdi32.NewProc("CreateFontW")
	procRoundRect              = gdi32.NewProc("RoundRect")
	procRectangle              = gdi32.NewProc("Rectangle")
	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procDeleteDC               = gdi32.NewProc("DeleteDC")

	procSHBrowseForFolderW   = shell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")
	procCoTaskMemFree        = ole32.NewProc("CoTaskMemFree")
)

const (
	WS_OVERLAPPED   = 0x00000000
	WS_CHILD        = 0x40000000
	WS_VISIBLE      = 0x10000000
	WS_CAPTION      = 0x00C00000
	WS_SYSMENU      = 0x00080000
	WS_MINIMIZEBOX  = 0x00020000
	WS_TABSTOP      = 0x00010000
	WS_EX_APPWINDOW = 0x00040000
	WS_EX_CLIENTEDGE = 0x00000200

	ES_LEFT        = 0x0000
	ES_CENTER      = 0x0001
	ES_MULTILINE   = 0x0004
	ES_AUTOHSCROLL = 0x0080

	EM_SETRECT = 0x00B3

	BS_OWNERDRAW    = 0x0000000B
	BS_AUTOCHECKBOX = 0x00000003
	BM_GETCHECK     = 0x00F0
	BM_SETCHECK     = 0x00F1
	BST_UNCHECKED   = 0x0000
	BST_CHECKED     = 0x0001

	WM_CREATE         = 0x0001
	WM_DESTROY        = 0x0002
	WM_PAINT          = 0x000F
	WM_ERASEBKGND     = 0x0014
	WM_COMMAND        = 0x0111
	WM_CTLCOLOREDIT   = 0x0133
	WM_CTLCOLORSTATIC = 0x0138
	WM_DRAWITEM       = 0x002B
	WM_LBUTTONDOWN    = 0x0201
	WM_SETCURSOR      = 0x0020
	WM_SETFONT        = 0x0030
	WM_CLOSE          = 0x0010

	TRANSPARENT = 1
	SRCCOPY     = 0x00CC0020

	DT_LEFT       = 0x00000000
	DT_CENTER     = 0x00000001
	DT_VCENTER    = 0x00000004
	DT_SINGLELINE = 0x00000020
	DT_NOPREFIX   = 0x00000800

	IDC_ARROW = 32512

	ID_BTN_INSTALL   = 201
	ID_CHK_DESKTOP   = 202
	ID_CHK_STARTMENU = 203
	ID_CHK_LAUNCH    = 204
	ID_EDIT_TARGET   = 205
	ID_BTN_BROWSE    = 206
	ID_BTN_UNINSTALL = 207
)

type RECT struct {
	Left, Top, Right, Bottom int32
}

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type DRAWITEMSTRUCT struct {
	CtlType    uint32
	CtlID      uint32
	ItemID     uint32
	ItemAction uint32
	ItemState  uint32
	HwndItem   uintptr
	HDC        uintptr
	RcItem     RECT
	ItemData   uintptr
}

func rgb(r, g, b byte) uint32 {
	return uint32(r) | (uint32(g) << 8) | (uint32(g) << 16)
}

var (
	colBg           = rgb(8, 10, 15)     // #080A0F
	colCardBg       = rgb(14, 18, 28)    // #0E121C
	colCardBorder   = rgb(28, 38, 58)    // #1C263A
	colInputBg      = rgb(20, 26, 40)    // #141A28
	colCyanAccent   = rgb(0, 240, 255)   // #00F0FF
	colPinkAccent   = rgb(255, 0, 85)    // #FF0055
	colGreenAccent  = rgb(0, 255, 128)   // #00FF80
	colYellowAccent = rgb(255, 230, 0)   // #FFE600
	colTextWhite    = rgb(255, 255, 255) // #FFFFFF
	colTextMuted    = rgb(120, 140, 175) // #788CAF
)

type SetupUI struct {
	hwnd             uintptr
	hwndEditPath     uintptr
	hwndBtnBrws      uintptr
	hwndBtn          uintptr
	hwndBtnUninstall uintptr
	hwndChkDesk      uintptr
	hwndChkStart     uintptr
	hwndChkRun       uintptr

	hCursorArrow uintptr
	hCursorHand  uintptr

	brushBg       uintptr
	brushCard     uintptr
	brushInput    uintptr
	penCardBorder uintptr
	penCyanAccent uintptr

	fontBrand   uintptr
	fontSection uintptr
	fontRegular uintptr
	fontMedium  uintptr
	fontInput   uintptr
	fontSmall   uintptr

	installing       bool
	done             bool
	alreadyInstalled bool   // true wenn App bereits im Zielordner vorhanden ist
	status           string
	targetDir        string // Zielordner für neue Installation
	installedDir     string // tatsächlich installierter Pfad (für Deinstallation)
}

var gSetup *SetupUI

func main() {
	// Prüfe auf Deinstallations-Modus
	if len(os.Args) > 1 && (os.Args[1] == "--uninstall" || os.Args[1] == "/uninstall") {
		runUninstall()
		return
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}
	targetDir := filepath.Join(localAppData, "Programs", "RIST Server")

	// Prüfen ob RIST Server bereits installiert ist (Standardpfad + Registry-Fallback)
	alreadyInstalled := false
	installedDir := "" // tatsächlich installierter Pfad (kann vom Standardpfad abweichen)

	// 1. Standardpfad prüfen
	if _, err := os.Stat(filepath.Join(targetDir, "RIST-Server.exe")); err == nil {
		alreadyInstalled = true
		installedDir = targetDir
	}

	// 2. Registry-Fallback: InstallLocation aus Uninstall-Eintrag lesen
	if !alreadyInstalled {
		regCmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
			`(Get-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\RISTServer' -ErrorAction SilentlyContinue).InstallLocation`)
		if out, err := regCmd.Output(); err == nil {
			loc := strings.TrimSpace(string(out))
			if loc != "" {
				if _, err2 := os.Stat(filepath.Join(loc, "RIST-Server.exe")); err2 == nil {
					alreadyInstalled = true
					installedDir = loc
					targetDir = loc // Zielordner auf installierten Pfad setzen
				} else {
					// Registry-Eintrag vorhanden aber Dateien fehlen → trotzdem anbieten zu bereinigen
					alreadyInstalled = true
					installedDir = loc
				}
			}
		}
	}

	statusText := "Bereit zur Installation."
	if alreadyInstalled {
		statusText = "RIST Server ist bereits installiert. Aktualisieren oder deinstallieren?"
	}

	ui := &SetupUI{
		targetDir:        targetDir,
		installedDir:     installedDir,
		status:           statusText,
		alreadyInstalled: alreadyInstalled,
	}
	gSetup = ui

	ui.brushBg = createSolidBrush(colBg)
	ui.brushCard = createSolidBrush(colCardBg)
	ui.brushInput = createSolidBrush(colInputBg)
	ui.penCardBorder = createPen(0, 1, colCardBorder)
	ui.penCyanAccent = createPen(0, 2, colCyanAccent)

	ui.fontBrand = createFont("Segoe UI", 18, 800)
	ui.fontSection = createFont("Segoe UI", 11, 700)
	ui.fontMedium = createFont("Segoe UI", 12, 600)
	ui.fontInput = createFont("Segoe UI", 11, 600)
	ui.fontRegular = createFont("Segoe UI", 10, 400)
	ui.fontSmall = createFont("Segoe UI", 9, 400)

	className, _ := syscall.UTF16PtrFromString("RISTSetupClass")
	windowTitle, _ := syscall.UTF16PtrFromString("RIST RECEIVER // SETUP")

	hCursorArrow, _, _ := procLoadCursorW.Call(0, uintptr(IDC_ARROW))
	hCursorHand, _, _ := procLoadCursorW.Call(0, uintptr(32649)) // IDC_HAND
	ui.hCursorArrow = hCursorArrow
	ui.hCursorHand = hCursorHand

	var hIcon uintptr
	if len(embeddedAppIcon) > 0 {
		hIcon, _, _ = procCreateIconFromResource.Call(
			uintptr(unsafe.Pointer(&embeddedAppIcon[0])),
			uintptr(len(embeddedAppIcon)),
			1,
			0x00030000,
			256, 256,
			0,
		)
	}

	wndClass := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		Style:         0x0003,
		LpfnWndProc:   syscall.NewCallback(setupWndProc),
		HInstance:     0,
		HIcon:         hIcon,
		HCursor:       hCursorArrow,
		HbrBackground: ui.brushBg,
		LpszClassName: className,
		HIconSm:       hIcon,
	}

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wndClass)))

	windowWidth := int32(470)
	windowHeight := int32(430) // Standardgröße ohne Deinstallations-Button
	if alreadyInstalled {
		windowHeight = 480 // Ausreichend Platz für Deinstallations-Button, Statuszeile, Ko-fi und unteren Abstand
	}

	hwnd, _, _ := procCreateWindowExW.Call(
		WS_EX_APPWINDOW,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowTitle)),
		WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU|WS_MINIMIZEBOX|WS_VISIBLE,
		0x80000000, 0x80000000,
		uintptr(windowWidth), uintptr(windowHeight),
		0, 0, 0, 0,
	)

	if hwnd == 0 {
		return
	}
	ui.hwnd = hwnd

	darkMode := int32(1)
	procDwmSetWindowAttribute.Call(
		hwnd,
		20, // DWMWA_USE_IMMERSIVE_DARK_MODE
		uintptr(unsafe.Pointer(&darkMode)),
		uintptr(unsafe.Sizeof(darkMode)),
	)

	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func setupWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	ui := gSetup
	if ui == nil {
		r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return r
	}

	switch msg {
	case WM_CREATE:
		ui.onCreate(hwnd)
		return 0

	case WM_COMMAND:
		ctrlID := int(wParam & 0xFFFF)
		if ctrlID == ID_BTN_INSTALL {
			if ui.done {
				procPostQuitMessage.Call(0)
			} else {
				ui.onInstall()
			}
		} else if ctrlID == ID_BTN_BROWSE {
			ui.onBrowseFolder()
		} else if ctrlID == ID_BTN_UNINSTALL {
			ui.onUninstallClick()
		}
		return 0

	case WM_CTLCOLOREDIT:
		hdc := wParam
		procSetBkColor.Call(hdc, uintptr(colInputBg))
		procSetTextColor.Call(hdc, uintptr(colCyanAccent))
		return ui.brushInput

	case WM_CTLCOLORSTATIC:
		hdc := wParam
		procSetBkMode.Call(hdc, TRANSPARENT)
		procSetTextColor.Call(hdc, uintptr(colTextMuted))
		return ui.brushCard

	case WM_LBUTTONDOWN:
		x := int32(lParam & 0xFFFF)
		y := int32((lParam >> 16) & 0xFFFF)
		// Ko-fi Support Link Klickbereich (dynamisch je nach Button-Anzahl)
		kofiTop := int32(366)
		if ui.alreadyInstalled {
			kofiTop = 412
		}
		if x >= 80 && x <= 390 && y >= kofiTop && y <= kofiTop+26 {
			ui.openKofiLink()
			return 0
		}

	case WM_SETCURSOR:
		var pt struct{ X, Y int32 }
		user32.NewProc("GetCursorPos").Call(uintptr(unsafe.Pointer(&pt)))
		user32.NewProc("ScreenToClient").Call(hwnd, uintptr(unsafe.Pointer(&pt)))
		kofiTopCur := int32(366)
		if ui.alreadyInstalled {
			kofiTopCur = 412
		}
		if pt.X >= 80 && pt.X <= 390 && pt.Y >= kofiTopCur && pt.Y <= kofiTopCur+26 {
			procSetCursor.Call(ui.hCursorHand)
			return 1
		}

	case WM_DRAWITEM:
		dis := (*DRAWITEMSTRUCT)(unsafe.Pointer(lParam))
		if dis.CtlID == ID_BTN_INSTALL {
			ui.drawButton(dis)
			return 1
		} else if dis.CtlID == ID_BTN_BROWSE {
			ui.drawBrowseButton(dis)
			return 1
		} else if dis.CtlID == ID_BTN_UNINSTALL {
			ui.drawUninstallButton(dis)
			return 1
		}

	case WM_ERASEBKGND:
		return 1

	case WM_PAINT:
		var ps struct {
			Hdc         uintptr
			FErase      int32
			RcPaint     RECT
			FRestore    int32
			FIncUpdate  int32
			RgbReserved [32]byte
		}
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		ui.onPaint(hdc)
		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_CLOSE, WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func (ui *SetupUI) onCreate(hwnd uintptr) {
	editClass, _ := syscall.UTF16PtrFromString("EDIT")
	btnClass, _ := syscall.UTF16PtrFromString("BUTTON")

	// Zielordner Edit Field (Exakt horizontal & vertikal zentrierter Text)
	txtDir, _ := syscall.UTF16PtrFromString(ui.targetDir)
	ui.hwndEditPath, _, _ = procCreateWindowExW.Call(
		WS_EX_CLIENTEDGE,
		uintptr(unsafe.Pointer(editClass)),
		uintptr(unsafe.Pointer(txtDir)),
		WS_CHILD|WS_VISIBLE|ES_MULTILINE|ES_CENTER|ES_AUTOHSCROLL|WS_TABSTOP,
		40, 116, 276, 26,
		hwnd,
		uintptr(ID_EDIT_TARGET),
		0, 0,
	)

	// Schriftart und vertikale Margins für exakte Mittigkeit im 26px Feld
	procSendMessageW.Call(ui.hwndEditPath, WM_SETFONT, ui.fontInput, 1)
	var editMargin = RECT{Left: 4, Top: 3, Right: 270, Bottom: 23}
	procSendMessageW.Call(ui.hwndEditPath, EM_SETRECT, 0, uintptr(unsafe.Pointer(&editMargin)))

	// Durchsuchen Button
	txtBrws, _ := syscall.UTF16PtrFromString("DURCHSUCHEN...")
	ui.hwndBtnBrws, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(txtBrws)),
		WS_CHILD|WS_VISIBLE|BS_OWNERDRAW|WS_TABSTOP,
		324, 116, 106, 26,
		hwnd,
		uintptr(ID_BTN_BROWSE),
		0, 0,
	)

	// Checkboxen
	txtDesk, _ := syscall.UTF16PtrFromString("Desktop-Verknüpfung erstellen")
	ui.hwndChkDesk, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(txtDesk)),
		WS_CHILD|WS_VISIBLE|BS_AUTOCHECKBOX|WS_TABSTOP,
		40, 156, 390, 22,
		hwnd,
		uintptr(ID_CHK_DESKTOP),
		0, 0,
	)

	txtStart, _ := syscall.UTF16PtrFromString("Im Windows-Startmenü eintragen")
	ui.hwndChkStart, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(txtStart)),
		WS_CHILD|WS_VISIBLE|BS_AUTOCHECKBOX|WS_TABSTOP,
		40, 184, 390, 22,
		hwnd,
		uintptr(ID_CHK_STARTMENU),
		0, 0,
	)

	txtRun, _ := syscall.UTF16PtrFromString("RIST Server nach Installation sofort starten")
	ui.hwndChkRun, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(txtRun)),
		WS_CHILD|WS_VISIBLE|BS_AUTOCHECKBOX|WS_TABSTOP,
		40, 212, 390, 22,
		hwnd,
		uintptr(ID_CHK_LAUNCH),
		0, 0,
	)

	procSendMessageW.Call(ui.hwndChkDesk, BM_SETCHECK, uintptr(BST_CHECKED), 0)
	procSendMessageW.Call(ui.hwndChkStart, BM_SETCHECK, uintptr(BST_CHECKED), 0)
	procSendMessageW.Call(ui.hwndChkRun, BM_SETCHECK, uintptr(BST_CHECKED), 0)

	procSendMessageW.Call(ui.hwndChkDesk, WM_SETFONT, ui.fontRegular, 1)
	procSendMessageW.Call(ui.hwndChkStart, WM_SETFONT, ui.fontRegular, 1)
	procSendMessageW.Call(ui.hwndChkRun, WM_SETFONT, ui.fontRegular, 1)

	// Install Button (Symmetrisch mit Card: w=422)
	btnText, _ := syscall.UTF16PtrFromString("JETZT INSTALLIEREN")
	ui.hwndBtn, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(btnText)),
		WS_CHILD|WS_VISIBLE|BS_OWNERDRAW|WS_TABSTOP,
		24, 290, 422, 46,
		hwnd,
		uintptr(ID_BTN_INSTALL),
		0, 0,
	)

	// Deinstallations-Button – nur sichtbar wenn App bereits installiert ist
	uninstBtnText, _ := syscall.UTF16PtrFromString("DEINSTALLIEREN")
	visibility := uint32(WS_CHILD | WS_VISIBLE | BS_OWNERDRAW | WS_TABSTOP)
	if !ui.alreadyInstalled {
		visibility = WS_CHILD | BS_OWNERDRAW // ohne WS_VISIBLE = versteckt
	}
	ui.hwndBtnUninstall, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(uninstBtnText)),
		uintptr(visibility),
		24, 344, 422, 34,
		hwnd,
		uintptr(ID_BTN_UNINSTALL),
		0, 0,
	)
}

func (ui *SetupUI) onBrowseFolder() {
	title, _ := syscall.UTF16PtrFromString("Wähle den Zielordner für RIST Server:")
	var bi struct {
		HwndOwner      uintptr
		PidlRoot       uintptr
		PszDisplayName *uint16
		LpszTitle      *uint16
		UlFlags        uint32
		Lpfn           uintptr
		LParam         uintptr
		IImage         int32
	}
	bi.HwndOwner = ui.hwnd
	bi.LpszTitle = title
	bi.UlFlags = 0x00000040 | 0x00000010 // BIF_NEWDIALOGSTYLE | BIF_EDITBOX

	pidl, _, _ := procSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return
	}
	defer procCoTaskMemFree.Call(pidl)

	buf := make([]uint16, 260)
	procSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&buf[0])))
	path := syscall.UTF16ToString(buf)
	if path != "" {
		newTarget := filepath.Join(path, "RIST Server")
		txt, _ := syscall.UTF16PtrFromString(newTarget)
		procSetWindowTextW.Call(ui.hwndEditPath, uintptr(unsafe.Pointer(txt)))
		var editMargin = RECT{Left: 4, Top: 3, Right: 270, Bottom: 23}
		procSendMessageW.Call(ui.hwndEditPath, EM_SETRECT, 0, uintptr(unsafe.Pointer(&editMargin)))
	}
}

func (ui *SetupUI) onPaint(hdc uintptr) {
	var rc RECT
	procGetClientRect.Call(ui.hwnd, uintptr(unsafe.Pointer(&rc)))
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top

	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	memBmp, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
	oldBmp, _, _ := procSelectObject.Call(memDC, memBmp)

	procFillRect.Call(memDC, uintptr(unsafe.Pointer(&rc)), ui.brushBg)

	// Header
	procSetBkMode.Call(memDC, TRANSPARENT)
	procSetTextColor.Call(memDC, uintptr(colCyanAccent))
	oldFont, _, _ := procSelectObject.Call(memDC, ui.fontBrand)
	drawText(memDC, "// RIST RECEIVER SETUP", 24, 16, 440, 42, DT_LEFT|DT_SINGLELINE)

	procSetTextColor.Call(memDC, uintptr(colTextMuted))
	procSelectObject.Call(memDC, ui.fontRegular)
	drawText(memDC, "Single-File RIST Relay Server für Windows", 24, 44, 440, 62, DT_LEFT|DT_SINGLELINE)

	// Card: Zielverzeichnis & Optionen
	ui.drawCard(memDC, 24, 74, 422, 196, "INSTALLATIONS-OPTIONEN")

	procSetTextColor.Call(memDC, uintptr(colTextWhite))
	procSelectObject.Call(memDC, ui.fontRegular)
	drawText(memDC, "Zielordner für die Installation:", 40, 94, 430, 112, DT_LEFT|DT_SINGLELINE)

	// Statuszeile & Ko-fi: Position hängt davon ab ob Deinstallations-Button sichtbar ist
	// Install-Button endet bei y=290+46=336
	// Deinstallations-Button bei y=344+34=378 (nur wenn alreadyInstalled)
	var statusY, kofiY int32
	if ui.alreadyInstalled {
		statusY = 388 // unter dem Deinstallations-Button
		kofiY = 412   // schön mit Abstand vor dem unteren Rand (Fensterhöhe 480)
	} else {
		statusY = 346 // direkt unter dem Install-Button
		kofiY = 370
	}

	procSetTextColor.Call(memDC, uintptr(colYellowAccent))
	if ui.done {
		procSetTextColor.Call(memDC, uintptr(colGreenAccent))
	} else if ui.alreadyInstalled && !ui.installing {
		procSetTextColor.Call(memDC, uintptr(colCyanAccent))
	}
	procSelectObject.Call(memDC, ui.fontRegular)
	drawText(memDC, ui.status, 24, statusY, 446, statusY+18, DT_CENTER|DT_SINGLELINE)

	// Ko-fi Support Link im Footer
	ui.drawKofiFooterLink(memDC, 24, kofiY, 422, 22)

	procBitBlt.Call(hdc, 0, 0, uintptr(w), uintptr(h), memDC, 0, 0, SRCCOPY)

	procSelectObject.Call(memDC, oldFont)
	procSelectObject.Call(memDC, oldBmp)
	procDeleteObject.Call(memBmp)
	procDeleteDC.Call(memDC)
}

func (ui *SetupUI) openKofiLink() {
	kofiURL, _ := syscall.UTF16PtrFromString("https://ko-fi.com/jumpgeil")
	op, _ := syscall.UTF16PtrFromString("open")
	procShellExecuteW.Call(0, uintptr(unsafe.Pointer(op)), uintptr(unsafe.Pointer(kofiURL)), 0, 0, 1)
}

func (ui *SetupUI) drawKofiFooterLink(hdc uintptr, x, y, w, h int32) {
	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(colPinkAccent))
	oldFont, _, _ := procSelectObject.Call(hdc, ui.fontSection)
	drawText(hdc, "♥ SUPPORT:  KO-FI.COM/JUMPGEIL", x, y, x+w, y+h, DT_CENTER|DT_VCENTER|DT_SINGLELINE)
	procSelectObject.Call(hdc, oldFont)
}

func (ui *SetupUI) drawCard(hdc uintptr, x, y, w, h int32, title string) {
	oldBrush, _, _ := procSelectObject.Call(hdc, ui.brushCard)
	oldPen, _, _ := procSelectObject.Call(hdc, ui.penCardBorder)

	procRoundRect.Call(hdc, uintptr(x), uintptr(y), uintptr(x+w), uintptr(y+h), 8, 8)

	cyanPen := createPen(0, 2, colCyanAccent)
	procSelectObject.Call(hdc, cyanPen)
	procRectangle.Call(hdc, uintptr(x+14), uintptr(y), uintptr(x+60), uintptr(y+2))

	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(colCyanAccent))
	procSelectObject.Call(hdc, ui.fontSection)
	drawText(hdc, title, x+16, y+8, x+w-16, y+24, DT_LEFT|DT_SINGLELINE)

	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(cyanPen)
}

func (ui *SetupUI) drawBrowseButton(dis *DRAWITEMSTRUCT) {
	hdc := dis.HDC
	rc := dis.RcItem

	btnBrush := createSolidBrush(rgb(20, 28, 44))
	btnPen := createPen(0, 1, colCyanAccent)

	oldBrush, _, _ := procSelectObject.Call(hdc, btnBrush)
	oldPen, _, _ := procSelectObject.Call(hdc, btnPen)

	procRoundRect.Call(hdc, uintptr(rc.Left), uintptr(rc.Top), uintptr(rc.Right), uintptr(rc.Bottom), 5, 5)

	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(colCyanAccent))
	oldFont, _, _ := procSelectObject.Call(hdc, ui.fontSmall)

	drawText(hdc, "DURCHSUCHEN...", rc.Left, rc.Top, rc.Right, rc.Bottom, DT_CENTER|DT_VCENTER|DT_SINGLELINE)

	procSelectObject.Call(hdc, oldFont)
	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(btnBrush)
	procDeleteObject.Call(btnPen)
}

func (ui *SetupUI) drawButton(dis *DRAWITEMSTRUCT) {
	hdc := dis.HDC
	rc := dis.RcItem

	var bgCol, borderCol uint32
	var label string

	if ui.done {
		bgCol = rgb(0, 100, 50)
		borderCol = colGreenAccent
		label = "✔   FERTIGSTELLEN & SCHLIESSEN"
	} else if ui.installing {
		bgCol = rgb(40, 34, 0)
		borderCol = colYellowAccent
		label = "⏳   INSTALLIERE..."
	} else {
		bgCol = rgb(0, 140, 75)
		borderCol = colGreenAccent
		label = "▶   JETZT INSTALLIEREN"
	}

	btnBrush := createSolidBrush(bgCol)
	btnPen := createPen(0, 2, borderCol)

	oldBrush, _, _ := procSelectObject.Call(hdc, btnBrush)
	oldPen, _, _ := procSelectObject.Call(hdc, btnPen)

	procRoundRect.Call(hdc, uintptr(rc.Left), uintptr(rc.Top), uintptr(rc.Right), uintptr(rc.Bottom), 8, 8)

	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(colTextWhite))
	oldFont, _, _ := procSelectObject.Call(hdc, ui.fontMedium)

	drawText(hdc, label, rc.Left, rc.Top, rc.Right, rc.Bottom, DT_CENTER|DT_VCENTER|DT_SINGLELINE)

	procSelectObject.Call(hdc, oldFont)
	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(btnBrush)
	procDeleteObject.Call(btnPen)
}

func (ui *SetupUI) drawUninstallButton(dis *DRAWITEMSTRUCT) {
	hdc := dis.HDC
	rc := dis.RcItem

	bgCol := rgb(80, 10, 20)     // dunkles Rot
	borderCol := colPinkAccent   // Pink/Rot-Akzent

	btnBrush := createSolidBrush(bgCol)
	btnPen := createPen(0, 2, borderCol)

	// Weiße Ecken verhindern
	bgBrush := createSolidBrush(colBg)
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), bgBrush)
	procDeleteObject.Call(bgBrush)

	oldBrush, _, _ := procSelectObject.Call(hdc, btnBrush)
	oldPen, _, _ := procSelectObject.Call(hdc, btnPen)

	procRoundRect.Call(hdc, uintptr(rc.Left), uintptr(rc.Top), uintptr(rc.Right), uintptr(rc.Bottom), 8, 8)

	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(colTextWhite))
	oldFont, _, _ := procSelectObject.Call(hdc, ui.fontMedium)

	drawText(hdc, "✕   DEINSTALLIEREN", rc.Left, rc.Top, rc.Right, rc.Bottom, DT_CENTER|DT_VCENTER|DT_SINGLELINE)

	procSelectObject.Call(hdc, oldFont)
	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(btnBrush)
	procDeleteObject.Call(btnPen)
}

// onUninstallClick führt die Deinstallation durch (wie --uninstall Modus, aber im GUI)
func (ui *SetupUI) onUninstallClick() {
	if ui.installing {
		return
	}

	titlePtr, _ := syscall.UTF16PtrFromString("RIST Server Deinstallation")
	msgPtr, _ := syscall.UTF16PtrFromString("Möchtest du RIST Server und alle Verknüpfungen wirklich vom System entfernen?")
	ret, _, _ := procMessageBoxW.Call(uintptr(ui.hwnd), uintptr(unsafe.Pointer(msgPtr)), uintptr(unsafe.Pointer(titlePtr)), 0x00000004|0x00000020)

	if ret != 6 { // IDYES = 6
		return
	}

	// Verknüpfungen und Registry löschen
	desktopLnk := filepath.Join(os.Getenv("USERPROFILE"), "Desktop", "RIST Server.lnk")
	_ = os.Remove(desktopLnk)

	startDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "RIST Server")
	_ = os.RemoveAll(startDir)

	delRegCmd := `Remove-Item -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\RISTServer' -Recurse -ErrorAction SilentlyContinue`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", delRegCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run()

	// Ordner per Batch-Script löschen (nach Beendigung des Installers)
	dirToRemove := ui.installedDir
	if dirToRemove == "" {
		dirToRemove = ui.targetDir
	}
	batContent := fmt.Sprintf(`@echo off
ping 127.0.0.1 -n 2 > nul
rmdir /s /q "%s"
`, dirToRemove)
	tmpBat := filepath.Join(os.TempDir(), "rist_cleanup.bat")
	_ = os.WriteFile(tmpBat, []byte(batContent), 0755)

	cleanupCmd := exec.Command("cmd.exe", "/C", tmpBat)
	cleanupCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cleanupCmd.Start()

	doneTitle, _ := syscall.UTF16PtrFromString("Deinstallation abgeschlossen")
	doneMsg, _ := syscall.UTF16PtrFromString("RIST Server wurde erfolgreich vom Computer entfernt.")
	procMessageBoxW.Call(uintptr(ui.hwnd), uintptr(unsafe.Pointer(doneMsg)), uintptr(unsafe.Pointer(doneTitle)), 0x00000040)

	procPostQuitMessage.Call(0)
}

func (ui *SetupUI) onInstall() {
	if ui.installing || ui.done {
		return
	}

	customDir := getWindowText(ui.hwndEditPath)
	if strings.TrimSpace(customDir) != "" {
		ui.targetDir = strings.TrimSpace(customDir)
	}

	chkDesk, _, _ := procSendMessageW.Call(ui.hwndChkDesk, BM_GETCHECK, 0, 0)
	chkStart, _, _ := procSendMessageW.Call(ui.hwndChkStart, BM_GETCHECK, 0, 0)
	chkRun, _, _ := procSendMessageW.Call(ui.hwndChkRun, BM_GETCHECK, 0, 0)

	createDesk := (chkDesk == BST_CHECKED)
	createStart := (chkStart == BST_CHECKED)
	launchNow := (chkRun == BST_CHECKED)

	ui.installing = true
	ui.status = "Kopiere Dateien und erstelle Verknüpfungen..."
	procEnableWindow.Call(ui.hwndEditPath, 0)
	procEnableWindow.Call(ui.hwndBtnBrws, 0)
	procEnableWindow.Call(ui.hwndChkDesk, 0)
	procEnableWindow.Call(ui.hwndChkStart, 0)
	procEnableWindow.Call(ui.hwndChkRun, 0)
	procInvalidateRect.Call(ui.hwnd, 0, 0)

	go func() {
		err := performInstall(ui.targetDir, createDesk, createStart)
		if err != nil {
			ui.installing = false
			ui.status = fmt.Sprintf("Fehler: %v", err)
			procInvalidateRect.Call(ui.hwnd, 0, 0)
			return
		}

		ui.installing = false
		ui.done = true
		ui.status = "✔ Installation erfolgreich abgeschlossen!"
		procInvalidateRect.Call(ui.hwnd, 0, 0)

		if launchNow {
			exePath := filepath.Join(ui.targetDir, "RIST-Server.exe")
			op, _ := syscall.UTF16PtrFromString("open")
			pathPtr, _ := syscall.UTF16PtrFromString(exePath)
			dirPtr, _ := syscall.UTF16PtrFromString(ui.targetDir)
			procShellExecuteW.Call(0, uintptr(unsafe.Pointer(op)), uintptr(unsafe.Pointer(pathPtr)), 0, uintptr(unsafe.Pointer(dirPtr)), 1)
			time.Sleep(500 * time.Millisecond)
			procPostQuitMessage.Call(0)
		}
	}()
}

func performInstall(targetDir string, createDesktop, createStartMenu bool) error {
	// 1. Ordner anlegen
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("Konnte Zielordner nicht erstellen: %w", err)
	}

	// 2. RIST-Server.exe schreiben
	targetExe := filepath.Join(targetDir, "RIST-Server.exe")
	if len(embeddedServerExe) > 0 {
		if err := os.WriteFile(targetExe, embeddedServerExe, 0755); err != nil {
			return fmt.Errorf("Konnte RIST-Server.exe nicht schreiben: %w", err)
		}
	} else {
		// Fallback falls lokal
		curExe, _ := os.Executable()
		curDir := filepath.Dir(curExe)
		srcExe := filepath.Join(curDir, "RIST-Server.exe")
		if data, err := os.ReadFile(srcExe); err == nil {
			_ = os.WriteFile(targetExe, data, 0755)
		}
	}

	// 3. Kopiere Installer als Uninstaller in das Verzeichnis
	selfExe, _ := os.Executable()
	uninstallerPath := filepath.Join(targetDir, "uninstall.exe")
	if selfData, err := os.ReadFile(selfExe); err == nil {
		_ = os.WriteFile(uninstallerPath, selfData, 0755)
	}

	// 4. Desktop Verknüpfung
	if createDesktop {
		desktopPath := filepath.Join(os.Getenv("USERPROFILE"), "Desktop", "RIST Server.lnk")
		_ = createShortcut(desktopPath, targetExe, targetExe, "RIST Receiver & Relay Server")
	}

	// 5. Startmenü Verknüpfung
	if createStartMenu {
		startDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "RIST Server")
		_ = os.MkdirAll(startDir, 0755)
		startLnk := filepath.Join(startDir, "RIST Server.lnk")
		_ = createShortcut(startLnk, targetExe, targetExe, "RIST Receiver & Relay Server")
	}

	// 6. Windows Uninstall Registry Eintrag
	registerUninstall(targetDir, uninstallerPath, targetExe)

	return nil
}

func createShortcut(shortcutPath, targetPath, iconPath, description string) error {
	psCmd := fmt.Sprintf(
		`$ws = New-Object -ComObject WScript.Shell; $s = $ws.CreateShortcut('%s'); $s.TargetPath = '%s'; $s.WorkingDirectory = '%s'; $s.IconLocation = '%s,0'; $s.Description = '%s'; $s.Save()`,
		escape(shortcutPath), escape(targetPath), escape(filepath.Dir(targetPath)), escape(iconPath), escape(description),
	)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

func registerUninstall(targetDir, uninstallerPath, exePath string) {
	regCmd := fmt.Sprintf(
		`$k = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\RISTServer'; `+
			`New-Item -Path $k -Force | Out-Null; `+
			`Set-ItemProperty -Path $k -Name 'DisplayName' -Value 'RIST Receiver Server'; `+
			`Set-ItemProperty -Path $k -Name 'DisplayIcon' -Value '%s'; `+
			`Set-ItemProperty -Path $k -Name 'DisplayVersion' -Value '1.0.1'; `+
			`Set-ItemProperty -Path $k -Name 'Publisher' -Value 'Jumpgeil'; `+
			`Set-ItemProperty -Path $k -Name 'InstallLocation' -Value '%s'; `+
			`Set-ItemProperty -Path $k -Name 'UninstallString' -Value '\"%s\" --uninstall'`,
		escape(exePath), escape(targetDir), escape(uninstallerPath),
	)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", regCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run()
}

func runUninstall() {
	titlePtr, _ := syscall.UTF16PtrFromString("RIST Server Deinstallation")
	msgPtr, _ := syscall.UTF16PtrFromString("Möchtest du RIST Server und alle Verknüpfungen wirklich vom System entfernen?")
	ret, _, _ := procMessageBoxW.Call(0, uintptr(unsafe.Pointer(msgPtr)), uintptr(unsafe.Pointer(titlePtr)), 0x00000004|0x00000020)

	if ret != 6 { // IDYES = 6
		return
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}
	targetDir := filepath.Join(localAppData, "Programs", "RIST Server")

	// 1. Verknüpfungen löschen
	desktopLnk := filepath.Join(os.Getenv("USERPROFILE"), "Desktop", "RIST Server.lnk")
	_ = os.Remove(desktopLnk)

	startDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "RIST Server")
	_ = os.RemoveAll(startDir)

	// 2. Registry Eintrag löschen
	delRegCmd := `Remove-Item -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\RISTServer' -Recurse -ErrorAction SilentlyContinue`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", delRegCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run()

	// 3. Batch Script zur Selbstlöschung des Ordners nach Beendigung
	batContent := fmt.Sprintf(`@echo off
ping 127.0.0.1 -n 2 > nul
rmdir /s /q "%s"
`, targetDir)
	tmpBat := filepath.Join(os.TempDir(), "rist_cleanup.bat")
	_ = os.WriteFile(tmpBat, []byte(batContent), 0755)

	cleanupCmd := exec.Command("cmd.exe", "/C", tmpBat)
	cleanupCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cleanupCmd.Start()

	doneTitle, _ := syscall.UTF16PtrFromString("Deinstallation abgeschlossen")
	doneMsg, _ := syscall.UTF16PtrFromString("RIST Server wurde erfolgreich vom Computer entfernt.")
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(doneMsg)), uintptr(unsafe.Pointer(doneTitle)), 0x00000040)
}

func escape(s string) string {
	return strings.ReplaceAll(s, `'`, `''`)
}

func getWindowText(hwnd uintptr) string {
	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(length+1))
	return syscall.UTF16ToString(buf)
}

func createSolidBrush(color uint32) uintptr {
	h, _, _ := procCreateSolidBrush.Call(uintptr(color))
	return h
}

func createPen(style int32, width int32, color uint32) uintptr {
	h, _, _ := procCreatePen.Call(uintptr(style), uintptr(width), uintptr(color))
	return h
}

func createFont(name string, size int32, weight int32) uintptr {
	namePtr, _ := syscall.UTF16PtrFromString(name)
	h, _, _ := procCreateFontW.Call(
		uintptr(-size), 0, 0, 0,
		uintptr(weight),
		0, 0, 0,
		1,
		0, 0, 0, 0,
		uintptr(unsafe.Pointer(namePtr)),
	)
	return h
}

func drawText(hdc uintptr, text string, left, top, right, bottom int32, flags uint32) {
	txtPtr, _ := syscall.UTF16PtrFromString(text)
	rc := RECT{Left: left, Top: top, Right: right, Bottom: bottom}
	procDrawTextW.Call(
		hdc,
		uintptr(unsafe.Pointer(txtPtr)),
		^uintptr(0),
		uintptr(unsafe.Pointer(&rc)),
		uintptr(flags|DT_NOPREFIX),
	)
}
