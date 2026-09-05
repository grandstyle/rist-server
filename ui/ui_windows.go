package ui

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"rist-server/assets"
	"rist-server/engine"
	"rist-server/settings"
	"rist-server/telemetry"
	"rist-server/version"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	dwmapi   = syscall.NewLazyDLL("dwmapi.dll")
	uxtheme  = syscall.NewLazyDLL("uxtheme.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	procRegisterClassExW       = user32.NewProc("RegisterClassExW")
	procCreateWindowExW        = user32.NewProc("CreateWindowExW")
	procDefWindowProcW         = user32.NewProc("DefWindowProcW")
	procShowWindow             = user32.NewProc("ShowWindow")
	procUpdateWindow           = user32.NewProc("UpdateWindow")
	procGetMessageW            = user32.NewProc("GetMessageW")
	procTranslateMessage       = user32.NewProc("TranslateMessage")
	procDispatchMessageW       = user32.NewProc("DispatchMessageW")
	procPostQuitMessage        = user32.NewProc("PostQuitMessage")
	procPostMessageW           = user32.NewProc("PostMessageW")
	procSendMessageW           = user32.NewProc("SendMessageW")
	procBeginPaint             = user32.NewProc("BeginPaint")
	procEndPaint               = user32.NewProc("EndPaint")
	procInvalidateRect         = user32.NewProc("InvalidateRect")
	procGetClientRect          = user32.NewProc("GetClientRect")
	procSetWindowTextW         = user32.NewProc("SetWindowTextW")
	procGetWindowTextW         = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW   = user32.NewProc("GetWindowTextLengthW")
	procLoadCursorW            = user32.NewProc("LoadCursorW")
	procSetCursor              = user32.NewProc("SetCursor")
	procGetCursorPos           = user32.NewProc("GetCursorPos")
	procScreenToClient         = user32.NewProc("ScreenToClient")
	procCreateIconFromResource = user32.NewProc("CreateIconFromResourceEx")
	procDrawTextW              = user32.NewProc("DrawTextW")
	procFillRect               = user32.NewProc("FillRect")
	procSetFocus               = user32.NewProc("SetFocus")
	procEnableWindow           = user32.NewProc("EnableWindow")
	procMessageBoxW            = user32.NewProc("MessageBoxW")
	procOpenClipboard          = user32.NewProc("OpenClipboard")
	procCloseClipboard         = user32.NewProc("CloseClipboard")
	procEmptyClipboard         = user32.NewProc("EmptyClipboard")
	procSetClipboardData       = user32.NewProc("SetClipboardData")
	procCreatePopupMenu        = user32.NewProc("CreatePopupMenu")
	procAppendMenuW            = user32.NewProc("AppendMenuW")
	procTrackPopupMenu         = user32.NewProc("TrackPopupMenu")
	procDestroyMenu            = user32.NewProc("DestroyMenu")
	procSetForegroundWindow    = user32.NewProc("SetForegroundWindow")
	procLoadImageW             = user32.NewProc("LoadImageW")
	procLoadIconW              = user32.NewProc("LoadIconW")
	procRegisterWindowMessageW = user32.NewProc("RegisterWindowMessageW")
	procDestroyWindow          = user32.NewProc("DestroyWindow")
	procGetWindowRect          = user32.NewProc("GetWindowRect")
	procGetWindowLongPtrW      = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW      = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW        = user32.NewProc("CallWindowProcW")
	procGetDC                  = user32.NewProc("GetDC")
	procReleaseDC              = user32.NewProc("ReleaseDC")
	procHideCaret              = user32.NewProc("HideCaret")

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
	procEllipse                = gdi32.NewProc("Ellipse")
	procGetTextExtentPoint32W  = gdi32.NewProc("GetTextExtentPoint32W")
	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procPolyline               = gdi32.NewProc("Polyline")

	procGlobalAlloc            = kernel32.NewProc("GlobalAlloc")
	procGlobalLock             = kernel32.NewProc("GlobalLock")
	procGlobalUnlock           = kernel32.NewProc("GlobalUnlock")
	procGetModuleHandleW       = kernel32.NewProc("GetModuleHandleW")

	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
	procSetWindowTheme        = uxtheme.NewProc("SetWindowTheme")
	procShellExecuteW         = shell32.NewProc("ShellExecuteW")
	procShell_NotifyIconW     = shell32.NewProc("Shell_NotifyIconW")
)

var wmTaskbarCreated uint32

const (
	WS_OVERLAPPED   = 0x00000000
	WS_POPUP        = 0x80000000
	WS_CHILD        = 0x40000000
	WS_MINIMIZE     = 0x20000000
	WS_VISIBLE      = 0x10000000
	WS_DISABLED     = 0x08000000
	WS_CLIPSIBLINGS = 0x04000000
	WS_CLIPCHILDREN = 0x02000000
	WS_CAPTION      = 0x00C00000
	WS_SYSMENU      = 0x00080000
	WS_MINIMIZEBOX  = 0x00020000
	WS_TABSTOP      = 0x00010000

	WS_EX_CLIENTEDGE = 0x00000200
	WS_EX_APPWINDOW  = 0x00040000

	ES_CENTER = 0x0001
	ES_NUMBER = 0x2000

	BS_OWNERDRAW    = 0x0000000B
	BS_AUTOCHECKBOX = 0x00000003

	BM_GETCHECK   = 0x00F0
	BM_SETCHECK   = 0x00F1
	BST_UNCHECKED = 0x0000
	BST_CHECKED   = 0x0001

	WM_CREATE         = 0x0001
	WM_DESTROY        = 0x0002
	WM_PAINT          = 0x000F
	WM_ERASEBKGND     = 0x0014
	WM_COMMAND        = 0x0111
	WM_SYSCOMMAND     = 0x0112
	WM_CTLCOLOREDIT   = 0x0133
	WM_CTLCOLORSTATIC = 0x0138
	WM_DRAWITEM       = 0x002B
	WM_LBUTTONDOWN    = 0x0201
	WM_LBUTTONUP      = 0x0202
	WM_LBUTTONDBLCLK  = 0x0203
	WM_RBUTTONUP      = 0x0205
	WM_MOUSEMOVE      = 0x0200
	WM_SETCURSOR      = 0x0020
	WM_SETFONT        = 0x0030
	WM_SETICON        = 0x0080
	WM_CLOSE          = 0x0010
	WM_SETFOCUS       = 0x0007
	WM_KILLFOCUS      = 0x0008
	WM_USER           = 0x0400

	ICON_SMALL = 0
	ICON_BIG   = 1

	WM_USER_TELEMETRY     = WM_USER + 1
	WM_USER_LOG           = WM_USER + 2
	WM_USER_CONFIG_UPDATE = WM_USER + 3
	WM_TRAY_ICON          = WM_USER + 10

	SC_MINIMIZE = 0xF020

	SW_HIDE    = 0
	SW_SHOW    = 5
	SW_RESTORE = 9

	NIM_ADD     = 0x00000000
	NIM_MODIFY  = 0x00000001
	NIM_DELETE  = 0x00000002
	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004

	MF_STRING       = 0x00000000
	MF_SEPARATOR    = 0x00000800
	TPM_RIGHTBUTTON = 0x0002

	DWMWA_USE_IMMERSIVE_DARK_MODE = 20

	SRCCOPY     = 0x00CC0020
	TRANSPARENT = 1

	DT_LEFT       = 0x00000000
	DT_CENTER     = 0x00000001
	DT_RIGHT      = 0x00000002
	DT_VCENTER    = 0x00000004
	DT_SINGLELINE = 0x00000020
	DT_WORDBREAK  = 0x00000010
	DT_NOPREFIX   = 0x00000800

	IDC_ARROW = 32512
	IDC_HAND  = 32649

	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 0x0002

	EM_SETCUEBANNER = 0x1501 // Platzhalter-Text für Edit-Felder

	GWLP_WNDPROC = -4 // SetWindowLongPtr: eigene WndProc einhängen
)

const (
	ID_EDIT_RIST_PORT  = 101
	ID_EDIT_OUT_PORT   = 102
	ID_BTN_START_STOP  = 103
	ID_BTN_COPY_LINK   = 104
	ID_CHK_AUTO_STREAM = 105
	ID_CHK_AUTO_WIN    = 106
	ID_BTN_COPY_OBS    = 107
	ID_EDIT_BUFFER     = 108
	ID_CHK_AUDIO       = 110
	ID_CHK_TRAY        = 111
	ID_BTN_COPY_REMOTE  = 112
	ID_BTN_COPY_OVERLAY = 113

	ID_BTN_BUF_500  = 120
	ID_BTN_BUF_1000 = 121
	ID_BTN_BUF_2000 = 122
	ID_BTN_BUF_3000 = 123

	ID_TRAY_OPEN   = 2001
	ID_TRAY_TOGGLE = 2002
	ID_TRAY_EXIT   = 2003

	ID_CONFIRM_EXIT   = 3001
	ID_CONFIRM_CANCEL = 3002
)

type RECT struct {
	Left, Top, Right, Bottom int32
}

type SIZE struct {
	CX, CY int32
}

type POINT struct {
	X, Y int32
}

type PAINTSTRUCT struct {
	Hdc         uintptr
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
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
	return uint32(r) | (uint32(g) << 8) | (uint32(b) << 16)
}

// Cyberpunk Farbpalette
var (
	colBg           = rgb(8, 10, 15)     // #080A0F Deep Void
	colCardBg       = rgb(14, 18, 28)    // #0E121C Dark Cyber Blue
	colCardBorder   = rgb(28, 38, 58)    // #1C263A Cyber Border
	colCyanAccent   = rgb(0, 240, 255)   // #00F0FF Neon Cyan
	colPinkAccent   = rgb(255, 0, 85)    // #FF0055 Cyber Pink / Hot Red
	colYellowAccent = rgb(255, 230, 0)   // #FFE600 Neon Yellow
	colGreenAccent  = rgb(0, 255, 128)   // #00FF80 Matrix Green
	colOrangeAccent = rgb(255, 140, 0)   // #FF8C00 Neon Orange
	colInputBg      = rgb(20, 26, 40)    // #141A28 Input Field BG
	colTextWhite    = rgb(255, 255, 255) // #FFFFFF
	colTextMuted    = rgb(120, 140, 175) // #788CAF Soft Cyber Grey
)

type NOTIFYICONDATAW struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
}

type UI struct {
	hwnd              uintptr
	hwndEditRist      uintptr
	hwndEditOut       uintptr
	hwndBtnBuf500     uintptr
	hwndBtnBuf1000    uintptr
	hwndBtnBuf2000    uintptr
	hwndBtnBuf3000    uintptr
	hwndBtnStart       uintptr
	hwndBtnCopy        uintptr
	hwndBtnCopyObs     uintptr
	hwndBtnCopyOverlay uintptr
	hwndBtnCopyRemote  uintptr
	hwndChkAutoStr     uintptr
	hwndChkAutoWin     uintptr
	hwndChkAudio       uintptr
	hwndChkTray        uintptr

	hCursorArrow uintptr
	hCursorHand  uintptr
	hIcon        uintptr
	hIconSm      uintptr
	nid          NOTIFYICONDATAW
	trayAdded    bool

	fontBrand   uintptr
	fontSection uintptr
	fontLarge   uintptr
	fontMedium  uintptr
	fontRegular uintptr
	fontSmall   uintptr
	fontIcon    uintptr

	brushBg       uintptr
	brushCard     uintptr
	brushInput    uintptr
	penCardBorder uintptr
	penCyanAccent uintptr

	engine     *engine.Engine
	telemetry  *telemetry.Parser
	apiPort    int
	lanIP      string
	config     settings.Config
	httpServer *telemetry.HTTPServer

	editRistOrigProc uintptr // Originale WndProc des RIST-Port Edit-Felds (für Subclassing)
	editOutOrigProc  uintptr // Originale WndProc des Ziel-IP Edit-Felds (für Subclassing)

	mu                 sync.Mutex
	statusMsg          string
	stats              telemetry.Stats
	copiedStatsTimer   int
	copiedObsTimer     int
	copiedOverlayTimer int
	copiedRemoteTimer  int

	updateAvailable bool
	updateVersion   string
	updateURL       string
}

var globalUI *UI

func loadAppIcon(w, h int32) uintptr {
	hInst, _, _ := procGetModuleHandleW.Call(0)
	if hInst != 0 {
		hIcon, _, _ := procLoadImageW.Call(
			hInst,
			uintptr(1),
			1, // IMAGE_ICON
			uintptr(w),
			uintptr(h),
			0x00008000, // LR_SHARED
		)
		if hIcon != 0 {
			return hIcon
		}
	}

	if len(assets.AppIcon) >= 22 {
		count := binary.LittleEndian.Uint16(assets.AppIcon[4:6])
		if count > 0 {
			bestIdx := 0
			bestDiff := int32(99999)
			for i := 0; i < int(count); i++ {
				entryOffset := 6 + i*16
				if entryOffset+16 > len(assets.AppIcon) {
					break
				}
				ew := int32(assets.AppIcon[entryOffset])
				if ew == 0 {
					ew = 256
				}
				diff := ew - w
				if diff < 0 {
					diff = -diff
				}
				if diff < bestDiff {
					bestDiff = diff
					bestIdx = i
				}
			}

			entryOffset := 6 + bestIdx*16
			bytesInRes := binary.LittleEndian.Uint32(assets.AppIcon[entryOffset+8 : entryOffset+12])
			offset := binary.LittleEndian.Uint32(assets.AppIcon[entryOffset+12 : entryOffset+16])
			if int(offset+bytesInRes) <= len(assets.AppIcon) {
				hIcon, _, _ := procCreateIconFromResource.Call(
					uintptr(unsafe.Pointer(&assets.AppIcon[offset])),
					uintptr(bytesInRes),
					1,
					0x00030000,
					uintptr(w),
					uintptr(h),
					0,
				)
				if hIcon != 0 {
					return hIcon
				}
			}
		}
	}

	hIcon, _, _ := procLoadIconW.Call(0, 32512) // IDI_APPLICATION
	return hIcon
}

// editSubclassProc ist die eigene WndProc für die Edit-Felder.
// Sie ruft die originale Proc auf und zeichnet danach ggf. den Platzhalter-Text.
func editSubclassProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	ui := globalUI
	if ui == nil {
		r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return r
	}

	var origProc uintptr
	var placeholder string
	if hwnd == ui.hwndEditRist {
		origProc = ui.editRistOrigProc
		placeholder = "Port (z.B. 7890)"
	} else if hwnd == ui.hwndEditOut {
		origProc = ui.editOutOrigProc
		placeholder = "z.B. 192.168.1.50:10000"
	} else {
		r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return r
	}

	r, _, _ := procCallWindowProcW.Call(origProc, hwnd, uintptr(msg), wParam, lParam)

	// Platzhalter-Logik: zeichnen wenn Feld leer ist
	drawPlaceholderIfEmpty := func() {
		length, _, _ := procGetWindowTextLengthW.Call(hwnd)
		if length == 0 {
			hdc, _, _ := procGetDC.Call(hwnd)
			if hdc != 0 {
				var rc RECT
				procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
				procSetBkMode.Call(hdc, TRANSPARENT)
				procSetTextColor.Call(hdc, uintptr(rgb(55, 75, 105)))
				oldFont, _, _ := procSelectObject.Call(hdc, ui.fontSmall)
				rc.Left += 4
				rc.Top += 2
				rc.Right -= 4
				rc.Bottom -= 2
				drawText(hdc, placeholder, rc.Left, rc.Top, rc.Right, rc.Bottom, DT_CENTER|DT_VCENTER|DT_SINGLELINE)
				procSelectObject.Call(hdc, oldFont)
				procReleaseDC.Call(hwnd, hdc)
			}
		}
	}

	switch msg {
	case WM_PAINT:
		// Nach dem originalen WM_PAINT Platzhalter drüberzeichnen
		drawPlaceholderIfEmpty()

	case WM_SETFOCUS:
		// Wenn Feld leer ist: Cursor verbergen und Platzhalter anzeigen
		length, _, _ := procGetWindowTextLengthW.Call(hwnd)
		if length == 0 {
			procHideCaret.Call(hwnd)
			procInvalidateRect.Call(hwnd, 0, 1)
		}

	case WM_KILLFOCUS:
		// Beim Fokus-Verlust neu zeichnen damit Platzhalter korrekt erscheint
		procInvalidateRect.Call(hwnd, 0, 1)
	}

	return r
}

func Run(eng *engine.Engine, tel *telemetry.Parser) error {
	return RunWithPortAndIP(eng, tel, 8686, telemetry.GetLocalIP())
}

func RunWithPort(eng *engine.Engine, tel *telemetry.Parser, apiPort int) error {
	return RunWithPortAndIP(eng, tel, apiPort, telemetry.GetLocalIP())
}

func RunWithPortAndIP(eng *engine.Engine, tel *telemetry.Parser, apiPort int, lanIP string, httpSrv ...*telemetry.HTTPServer) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if lanIP == "" {
		lanIP = "127.0.0.1"
	}

	cfg := settings.LoadConfig()

	ui := &UI{
		engine:    eng,
		telemetry: tel,
		apiPort:   apiPort,
		lanIP:     lanIP,
		config:    cfg,
		statusMsg: "Bereit. Klicke auf 'STREAM EMPFANGEN', um zu starten.",
		stats:     tel.GetStats(),
	}
	globalUI = ui

	if len(httpSrv) > 0 && httpSrv[0] != nil {
		ui.httpServer = httpSrv[0]
		ui.httpServer.SetController(
			ui.ToggleServer,
			ui.StartServer,
			ui.StopServer,
			func() bool { return ui.engine.IsRunning() },
			func() map[string]interface{} {
				ui.mu.Lock()
				defer ui.mu.Unlock()
				return map[string]interface{}{
					"rist_port":        ui.config.RistPort,
					"out_host":         ui.config.OutHost,
					"out_port":         ui.config.OutPort,
					"buffer_ms":        ui.config.BufferMs,
					"audio_alerts":     ui.config.AudioAlerts,
					"minimize_to_tray": ui.config.MinimizeToTray,
					"running":          ui.engine.IsRunning(),
				}
			},
			func(payload map[string]interface{}) error {
				ui.mu.Lock()
				if val, ok := payload["buffer_ms"]; ok {
					switch v := val.(type) {
					case float64:
						ui.config.BufferMs = int(v)
					case int:
						ui.config.BufferMs = v
					}
				}
				if val, ok := payload["audio_alerts"]; ok {
					if b, ok := val.(bool); ok {
						ui.config.AudioAlerts = b
					}
				}
				_ = settings.SaveConfig(ui.config)
				ui.mu.Unlock()
				if ui.hwnd != 0 {
					procPostMessageW.Call(ui.hwnd, WM_USER_CONFIG_UPDATE, 0, 0)
				}
				return nil
			},
		)
	}

	// Brushes & Pens
	ui.brushBg = createSolidBrush(colBg)
	ui.brushCard = createSolidBrush(colCardBg)
	ui.brushInput = createSolidBrush(colInputBg)
	ui.penCardBorder = createPen(0, 1, colCardBorder)
	ui.penCyanAccent = createPen(0, 2, colCyanAccent)

	// Fonts
	ui.fontBrand = createFont("Segoe UI", 20, 800)
	ui.fontSection = createFont("Segoe UI", 11, 700)
	ui.fontLarge = createFont("Segoe UI", 22, 800)
	ui.fontMedium = createFont("Segoe UI", 13, 600)
	ui.fontRegular = createFont("Segoe UI", 11, 400)
	ui.fontSmall = createFont("Segoe UI", 9, 600)
	ui.fontIcon = createFont("Segoe UI", 14, 600)

	className, _ := syscall.UTF16PtrFromString("RISTCyberSrvClass")
	windowTitle, _ := syscall.UTF16PtrFromString(fmt.Sprintf("RIST RECEIVER // SERVER v%s", version.Version))

	taskbarMsgStr, _ := syscall.UTF16PtrFromString("TaskbarCreated")
	tbMsg, _, _ := procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(taskbarMsgStr)))
	wmTaskbarCreated = uint32(tbMsg)

	hCursorArrow, _, _ := procLoadCursorW.Call(0, uintptr(IDC_ARROW))
	hCursorHand, _, _ := procLoadCursorW.Call(0, uintptr(IDC_HAND))
	ui.hCursorArrow = hCursorArrow
	ui.hCursorHand = hCursorHand

	hIconBig := loadAppIcon(32, 32)
	hIconSm := loadAppIcon(16, 16)
	ui.hIcon = hIconBig
	ui.hIconSm = hIconSm

	wndClass := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		Style:         0x0003,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     0,
		HIcon:         hIconBig,
		HCursor:       hCursorArrow,
		HbrBackground: ui.brushBg,
		LpszClassName: className,
		HIconSm:       hIconSm,
	}

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wndClass)))

	// Fenster-Dimension: 460 x 680 Pixel
	windowWidth := int32(460)
	windowHeight := int32(680)

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
		return fmt.Errorf("failed to create main window")
	}

	ui.hwnd = hwnd

	procSendMessageW.Call(hwnd, WM_SETICON, uintptr(ICON_SMALL), hIconSm)
	procSendMessageW.Call(hwnd, WM_SETICON, uintptr(ICON_BIG), hIconBig)
	ui.initTray(hwnd)

	// Windows Dark Mode Title Bar
	darkMode := int32(1)
	procDwmSetWindowAttribute.Call(
		hwnd,
		DWMWA_USE_IMMERSIVE_DARK_MODE,
		uintptr(unsafe.Pointer(&darkMode)),
		uintptr(unsafe.Sizeof(darkMode)),
	)

	// Callback für Telemetrie-Updates
	prevState := tel.GetStats().State
	engTelCallback := func(s telemetry.Stats) {
		ui.mu.Lock()
		ui.stats = s
		prev := prevState
		if s.State != prevState {
			if s.State == telemetry.StateStreaming {
				ui.statusMsg = "✔ Stream aktiv! Daten werden empfangen..."
				if ui.config.AudioAlerts {
					go telemetry.PlayStreamConnected()
				}
			} else if s.State == telemetry.StateListening && prev == telemetry.StateStreaming {
				ui.statusMsg = "Stream unterbrochen / kein Signal. Warte auf Stream..."
				if ui.config.AudioAlerts {
					go telemetry.PlayStreamDisconnected()
				}
			}
			prevState = s.State
		}
		ui.mu.Unlock()
		procPostMessageW.Call(hwnd, WM_USER_TELEMETRY, 0, 0)
	}
	tel.SetOnUpdate(engTelCallback)

	// Falls Autostart aktiv -> Nach kurzer Verzögerung automatisch starten
	if cfg.AutoStartStream {
		go func() {
			time.Sleep(300 * time.Millisecond)
			ui.onToggleServer()
		}()
	}

	// Nach kurzer Pause im Hintergrund prüfen, ob neue Releases auf GitHub verfügbar sind
	go func() {
		time.Sleep(1500 * time.Millisecond)
		ui.checkForUpdates()
	}()

	// Message Loop
	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	return nil
}

func GetGlobalUI() *UI {
	return globalUI
}

func (ui *UI) SetLog(text string) {
	ui.mu.Lock()
	ui.statusMsg = text
	ui.mu.Unlock()
	if ui.hwnd != 0 {
		procPostMessageW.Call(ui.hwnd, WM_USER_LOG, 0, 0)
	}
}

func (ui *UI) initTray(hwnd uintptr) {
	if ui.trayAdded {
		return
	}
	icon := ui.hIconSm
	if icon == 0 {
		icon = ui.hIcon
	}
	ui.nid = NOTIFYICONDATAW{
		CbSize:           uint32(unsafe.Sizeof(NOTIFYICONDATAW{})),
		HWnd:             hwnd,
		UID:              1,
		UFlags:           NIF_MESSAGE | NIF_ICON | NIF_TIP,
		UCallbackMessage: WM_TRAY_ICON,
		HIcon:            icon,
	}
	tip, _ := syscall.UTF16FromString("RIST Receiver // Server")
	copy(ui.nid.SzTip[:], tip)
	ret, _, _ := procShell_NotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&ui.nid)))
	if ret != 0 {
		ui.trayAdded = true
	}
}

func (ui *UI) cleanUpTray() {
	if ui.trayAdded {
		procShell_NotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&ui.nid)))
		ui.trayAdded = false
	}
}

func (ui *UI) showTrayContextMenu(hwnd uintptr) {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}
	defer procDestroyMenu.Call(hMenu)

	openStr, _ := syscall.UTF16PtrFromString("Öffnen")
	procAppendMenuW.Call(hMenu, MF_STRING, uintptr(ID_TRAY_OPEN), uintptr(unsafe.Pointer(openStr)))

	sepStr, _ := syscall.UTF16PtrFromString("")
	procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, uintptr(unsafe.Pointer(sepStr)))

	var toggleText string
	if ui.engine.IsRunning() {
		toggleText = "Stream stoppen"
	} else {
		toggleText = "Stream starten"
	}
	toggleStr, _ := syscall.UTF16PtrFromString(toggleText)
	procAppendMenuW.Call(hMenu, MF_STRING, uintptr(ID_TRAY_TOGGLE), uintptr(unsafe.Pointer(toggleStr)))

	procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, uintptr(unsafe.Pointer(sepStr)))

	exitStr, _ := syscall.UTF16PtrFromString("Beenden")
	procAppendMenuW.Call(hMenu, MF_STRING, uintptr(ID_TRAY_EXIT), uintptr(unsafe.Pointer(exitStr)))

	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetForegroundWindow.Call(hwnd)
	procTrackPopupMenu.Call(hMenu, TPM_RIGHTBUTTON, uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	ui := globalUI
	if ui == nil {
		r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
		return r
	}

	if wmTaskbarCreated != 0 && msg == wmTaskbarCreated {
		ui.trayAdded = false
		ui.initTray(hwnd)
		return 0
	}

	switch msg {
	case WM_CREATE:
		ui.onCreate(hwnd)
		return 0

	case WM_USER_TELEMETRY, WM_USER_LOG:
		procInvalidateRect.Call(hwnd, 0, 0)
		if ui.hwndBtnStart != 0 {
			procInvalidateRect.Call(ui.hwndBtnStart, 0, 0)
		}
		if ui.hwndBtnCopy != 0 {
			procInvalidateRect.Call(ui.hwndBtnCopy, 0, 0)
		}
		if ui.hwndBtnCopyObs != 0 {
			procInvalidateRect.Call(ui.hwndBtnCopyObs, 0, 0)
		}
		if ui.hwndBtnCopyRemote != 0 {
			procInvalidateRect.Call(ui.hwndBtnCopyRemote, 0, 0)
		}
		if ui.hwndBtnCopyOverlay != 0 {
			procInvalidateRect.Call(ui.hwndBtnCopyOverlay, 0, 0)
		}
		return 0

	case WM_USER_CONFIG_UPDATE:
		// Buffer-Pill-Buttons neu zeichnen (aktiver Button hat sich ggf. geändert)
		for _, h := range []uintptr{ui.hwndBtnBuf500, ui.hwndBtnBuf1000, ui.hwndBtnBuf2000, ui.hwndBtnBuf3000} {
			if h != 0 {
				procInvalidateRect.Call(h, 0, 0)
			}
		}
		if ui.hwndChkAudio != 0 {
			v := uintptr(BST_UNCHECKED)
			if ui.config.AudioAlerts {
				v = uintptr(BST_CHECKED)
			}
			procSendMessageW.Call(ui.hwndChkAudio, BM_SETCHECK, v, 0)
		}
		if ui.hwndChkTray != 0 {
			v := uintptr(BST_UNCHECKED)
			if ui.config.MinimizeToTray {
				v = uintptr(BST_CHECKED)
			}
			procSendMessageW.Call(ui.hwndChkTray, BM_SETCHECK, v, 0)
		}
		procInvalidateRect.Call(hwnd, 0, 0)
		if ui.hwndBtnStart != 0 {
			procInvalidateRect.Call(ui.hwndBtnStart, 0, 0)
		}
		return 0

	case WM_TRAY_ICON:
		switch lParam {
		case WM_LBUTTONDBLCLK, WM_LBUTTONUP:
			procShowWindow.Call(hwnd, SW_RESTORE)
			procSetForegroundWindow.Call(hwnd)
		case WM_RBUTTONUP:
			ui.showTrayContextMenu(hwnd)
		}
		return 0

	case WM_SYSCOMMAND:
		if (wParam&0xFFF0) == SC_MINIMIZE && ui.config.MinimizeToTray {
			procShowWindow.Call(hwnd, SW_HIDE)
			return 0
		}

	case WM_COMMAND:
		ctrlID := int(wParam & 0xFFFF)
		notif := int(wParam >> 16)
		// EN_CHANGE (0x0300): Edit-Feld hat sich geändert → Fenster neu zeichnen damit Platzhalter erscheint/verschwindet
		if notif == 0x0300 && (ctrlID == ID_EDIT_RIST_PORT || ctrlID == ID_EDIT_OUT_PORT) {
			procInvalidateRect.Call(hwnd, 0, 0)
		}
		switch ctrlID {
		case ID_BTN_START_STOP:
			ui.onToggleServer()
		case ID_BTN_COPY_LINK:
			ui.onCopyStatsLink()
		case ID_BTN_COPY_OBS:
			ui.onCopyObsLink()
		case ID_BTN_COPY_OVERLAY:
			ui.onCopyOverlayLink()
		case ID_BTN_COPY_REMOTE:
			ui.onCopyRemoteLink()
		case ID_CHK_AUTO_STREAM:
			ui.onToggleAutoStream()
		case ID_CHK_AUTO_WIN:
			ui.onToggleAutoWin()
		case ID_CHK_AUDIO:
			ui.onToggleAudio()
		case ID_CHK_TRAY:
			ui.onToggleTray()
		case ID_BTN_BUF_500:
			ui.setBufferMs(500)
		case ID_BTN_BUF_1000:
			ui.setBufferMs(1000)
		case ID_BTN_BUF_2000:
			ui.setBufferMs(2000)
		case ID_BTN_BUF_3000:
			ui.setBufferMs(3000)
		case ID_TRAY_OPEN:
			procShowWindow.Call(hwnd, SW_RESTORE)
			procSetForegroundWindow.Call(hwnd)
		case ID_TRAY_TOGGLE:
			ui.onToggleServer()
		case ID_TRAY_EXIT:
			procShowWindow.Call(hwnd, SW_RESTORE)
			procSetForegroundWindow.Call(hwnd)
			if ui.engine.IsRunning() {
				if !ui.showConfirmExitDialog() {
					return 0
				}
			}
			ui.cleanUpTray()
			if ui.engine.IsRunning() {
				ui.engine.Stop()
				time.Sleep(100 * time.Millisecond)
			}
			procPostQuitMessage.Call(0)
		}
		return 0

	case WM_CTLCOLOREDIT:
		hdc := wParam
		procSetBkColor.Call(hdc, uintptr(colInputBg))
		// Gleiche Cyan-Farbe wie aktiver Jitter-Button; beim laufenden Stream gedimmt
		if ui.engine.IsRunning() {
			procSetTextColor.Call(hdc, uintptr(colTextMuted))
		} else {
			procSetTextColor.Call(hdc, uintptr(colCyanAccent))
		}
		return ui.brushInput

	case WM_CTLCOLORSTATIC:
		hdc := wParam
		procSetBkMode.Call(hdc, TRANSPARENT)
		procSetTextColor.Call(hdc, uintptr(colTextMuted))
		return ui.brushCard

	case WM_DRAWITEM:
		dis := (*DRAWITEMSTRUCT)(unsafe.Pointer(lParam))
		if dis.CtlID == ID_BTN_START_STOP {
			ui.drawCyberButton(dis)
			return 1
		} else if dis.CtlID == ID_BTN_COPY_LINK {
			ui.mu.Lock()
			copied := ui.copiedStatsTimer > 0
			ui.mu.Unlock()
			ui.drawCopyButton(dis, copied)
			return 1
		} else if dis.CtlID == ID_BTN_COPY_OBS {
			ui.mu.Lock()
			copied := ui.copiedObsTimer > 0
			ui.mu.Unlock()
			ui.drawCopyButton(dis, copied)
			return 1
		} else if dis.CtlID == ID_BTN_COPY_OVERLAY {
			ui.mu.Lock()
			copied := ui.copiedOverlayTimer > 0
			ui.mu.Unlock()
			ui.drawCopyButton(dis, copied)
			return 1
		} else if dis.CtlID == ID_BTN_COPY_REMOTE {
			ui.mu.Lock()
			copied := ui.copiedRemoteTimer > 0
			ui.mu.Unlock()
			ui.drawCopyButton(dis, copied)
			return 1
		} else if dis.CtlID == ID_BTN_BUF_500 {
			ui.drawBufferPill(dis, 500)
			return 1
		} else if dis.CtlID == ID_BTN_BUF_1000 {
			ui.drawBufferPill(dis, 1000)
			return 1
		} else if dis.CtlID == ID_BTN_BUF_2000 {
			ui.drawBufferPill(dis, 2000)
			return 1
		} else if dis.CtlID == ID_BTN_BUF_3000 {
			ui.drawBufferPill(dis, 3000)
			return 1
		}

	case WM_LBUTTONDOWN:
		x := int32(lParam & 0xFFFF)
		y := int32((lParam >> 16) & 0xFFFF)
		// Ko-fi Support Link Klickbereich (x: 90 bis 370, y: 594 bis 624)
		if x >= 90 && x <= 370 && y >= 594 && y <= 624 {
			ui.openKofiLink()
			return 0
		}
		// Update Badge Klickbereich (x: 24 bis 180, y: 38 bis 58)
		ui.mu.Lock()
		hasUpd := ui.updateAvailable
		ui.mu.Unlock()
		if hasUpd && x >= 24 && x <= 180 && y >= 38 && y <= 58 {
			ui.openUpdateLink()
			return 0
		}

	case WM_SETCURSOR:
		var pt POINT
		procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
		procScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&pt)))
		if pt.X >= 90 && pt.X <= 370 && pt.Y >= 594 && pt.Y <= 624 {
			procSetCursor.Call(ui.hCursorHand)
			return 1
		}
		ui.mu.Lock()
		hasUpdCursor := ui.updateAvailable
		ui.mu.Unlock()
		if hasUpdCursor && pt.X >= 24 && pt.X <= 180 && pt.Y >= 38 && pt.Y <= 58 {
			procSetCursor.Call(ui.hCursorHand)
			return 1
		}

	case WM_ERASEBKGND:
		return 1

	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		ui.onPaint(hdc)
		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_CLOSE:
		if ui.engine.IsRunning() {
			if !ui.showConfirmExitDialog() {
				return 0
			}
		}
		ui.cleanUpTray()
		if ui.engine.IsRunning() {
			ui.engine.Stop()
			time.Sleep(100 * time.Millisecond)
		}
		procPostQuitMessage.Call(0)
		return 0

	case WM_DESTROY:
		ui.cleanUpTray()
		procPostQuitMessage.Call(0)
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func (ui *UI) onCreate(hwnd uintptr) {
	editClass, _ := syscall.UTF16PtrFromString("EDIT")
	btnClass, _ := syscall.UTF16PtrFromString("BUTTON")

	// RIST Port (x=307, y=87, w=98, h=22) – kein WS_EX_CLIENTEDGE, border wird custom gemalt
	txtRist, _ := syscall.UTF16PtrFromString(strconv.Itoa(ui.config.RistPort))
	ui.hwndEditRist, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(editClass)),
		uintptr(unsafe.Pointer(txtRist)),
		WS_CHILD|WS_VISIBLE|ES_CENTER|ES_NUMBER|WS_TABSTOP,
		307, 87, 98, 22,
		hwnd,
		uintptr(ID_EDIT_RIST_PORT),
		0, 0,
	)

	// Forward UDP Target IP:Port (x=241, y=115, w=164, h=22) – kein WS_EX_CLIENTEDGE
	outHost := ui.config.OutHost
	if outHost == "" {
		outHost = "127.0.0.1"
	}
	outText := fmt.Sprintf("%s:%d", outHost, ui.config.OutPort)
	txtOut, _ := syscall.UTF16PtrFromString(outText)
	ui.hwndEditOut, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(editClass)),
		uintptr(unsafe.Pointer(txtOut)),
		WS_CHILD|WS_VISIBLE|ES_CENTER|WS_TABSTOP,
		241, 115, 164, 22,
		hwnd,
		uintptr(ID_EDIT_OUT_PORT),
		0, 0,
	)

	// Beide Edit-Controls subclassen damit wir Platzhalter-Text zeichnen können
	subclassCb := syscall.NewCallback(editSubclassProc)
	const gwlpWndProc = ^uintptr(3) // == -4 als uintptr (GWLP_WNDPROC)
	origRist, _, _ := procSetWindowLongPtrW.Call(ui.hwndEditRist, gwlpWndProc, subclassCb)
	ui.editRistOrigProc = origRist
	origOut, _, _ := procSetWindowLongPtrW.Call(ui.hwndEditOut, gwlpWndProc, subclassCb)
	ui.editOutOrigProc = origOut

	// RIST Puffer (Jitter ms): 4 Auswahl-Buttons (Pills) à 54px breit, 4px Gap
	// Startposition x=240, y=142, h=24
	bufBtnLabels := []string{"500ms", "1000ms", "2000ms", "3000ms"}
	bufBtnIDs := []uintptr{ID_BTN_BUF_500, ID_BTN_BUF_1000, ID_BTN_BUF_2000, ID_BTN_BUF_3000}
	bufBtnHwnds := []*uintptr{&ui.hwndBtnBuf500, &ui.hwndBtnBuf1000, &ui.hwndBtnBuf2000, &ui.hwndBtnBuf3000}
	for i := 0; i < 4; i++ {
		lbl, _ := syscall.UTF16PtrFromString(bufBtnLabels[i])
		hBtn, _, _ := procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(btnClass)),
			uintptr(unsafe.Pointer(lbl)),
			WS_CHILD|WS_VISIBLE|BS_OWNERDRAW|WS_TABSTOP,
			uintptr(240+i*42), 142, 40, 24,
			hwnd,
			bufBtnIDs[i],
			0, 0,
		)
		*bufBtnHwnds[i] = hBtn
	}


	// Checkboxen:
	// Row 1: Stream Autostart (x=40, y=170) & Audio-Alert (x=230, y=170)
	txtAutoStr, _ := syscall.UTF16PtrFromString("Stream Autostart")
	ui.hwndChkAutoStr, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(txtAutoStr)),
		WS_CHILD|WS_VISIBLE|BS_AUTOCHECKBOX|WS_TABSTOP,
		40, 170, 180, 20,
		hwnd,
		uintptr(ID_CHK_AUTO_STREAM),
		0, 0,
	)

	txtAudio, _ := syscall.UTF16PtrFromString("Audio-Alert bei Drop")
	ui.hwndChkAudio, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(txtAudio)),
		WS_CHILD|WS_VISIBLE|BS_AUTOCHECKBOX|WS_TABSTOP,
		230, 170, 180, 20,
		hwnd,
		uintptr(ID_CHK_AUDIO),
		0, 0,
	)

	// Row 2: Windows Autostart (x=40, y=194) & In System-Tray minimieren (x=230, y=194)
	txtAutoWin, _ := syscall.UTF16PtrFromString("Windows Autostart")
	ui.hwndChkAutoWin, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(txtAutoWin)),
		WS_CHILD|WS_VISIBLE|BS_AUTOCHECKBOX|WS_TABSTOP,
		40, 194, 180, 20,
		hwnd,
		uintptr(ID_CHK_AUTO_WIN),
		0, 0,
	)

	txtTray, _ := syscall.UTF16PtrFromString("In System-Tray minimieren")
	ui.hwndChkTray, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(txtTray)),
		WS_CHILD|WS_VISIBLE|BS_AUTOCHECKBOX|WS_TABSTOP,
		230, 194, 180, 20,
		hwnd,
		uintptr(ID_CHK_TRAY),
		0, 0,
	)

	// Setze Checkbox Initialzustände
	chkVal := func(b bool) uintptr {
		if b {
			return uintptr(BST_CHECKED)
		}
		return uintptr(BST_UNCHECKED)
	}
	procSendMessageW.Call(ui.hwndChkAutoStr, BM_SETCHECK, chkVal(ui.config.AutoStartStream), 0)
	procSendMessageW.Call(ui.hwndChkAutoWin, BM_SETCHECK, chkVal(ui.config.AutoStartWindows), 0)
	procSendMessageW.Call(ui.hwndChkAudio, BM_SETCHECK, chkVal(ui.config.AudioAlerts), 0)
	procSendMessageW.Call(ui.hwndChkTray, BM_SETCHECK, chkVal(ui.config.MinimizeToTray), 0)

	// 1. OBS Stream Link Copy Button (x=368, y=339, w=32, h=32)
	copyBtnText, _ := syscall.UTF16PtrFromString("")
	ui.hwndBtnCopyObs, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(copyBtnText)),
		WS_CHILD|WS_VISIBLE|BS_OWNERDRAW|WS_TABSTOP,
		368, 339, 32, 32,
		hwnd,
		uintptr(ID_BTN_COPY_OBS),
		0, 0,
	)

	// 2. OBS Overlay Link Copy Button (x=368, y=381, w=32, h=32)
	ui.hwndBtnCopyOverlay, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(copyBtnText)),
		WS_CHILD|WS_VISIBLE|BS_OWNERDRAW|WS_TABSTOP,
		368, 381, 32, 32,
		hwnd,
		uintptr(ID_BTN_COPY_OVERLAY),
		0, 0,
	)

	// 3. Stats Link Copy Button (x=368, y=423, w=32, h=32)
	ui.hwndBtnCopy, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(copyBtnText)),
		WS_CHILD|WS_VISIBLE|BS_OWNERDRAW|WS_TABSTOP,
		368, 423, 32, 32,
		hwnd,
		uintptr(ID_BTN_COPY_LINK),
		0, 0,
	)

	// 4. Mobile Remote Copy Button (x=368, y=465, w=32, h=32)
	ui.hwndBtnCopyRemote, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(copyBtnText)),
		WS_CHILD|WS_VISIBLE|BS_OWNERDRAW|WS_TABSTOP,
		368, 465, 32, 32,
		hwnd,
		uintptr(ID_BTN_COPY_REMOTE),
		0, 0,
	)

	// Start / Stop Button (x=24, y=516, w=396, h=46)
	btnText, _ := syscall.UTF16PtrFromString("STREAM EMPFANGEN")
	ui.hwndBtnStart, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(btnText)),
		WS_CHILD|WS_VISIBLE|BS_OWNERDRAW|WS_TABSTOP,
		24, 516, 396, 46,
		hwnd,
		uintptr(ID_BTN_START_STOP),
		0, 0,
	)

	procSendMessageW.Call(ui.hwndEditRist, WM_SETFONT, ui.fontMedium, 1)
	procSendMessageW.Call(ui.hwndEditOut, WM_SETFONT, ui.fontMedium, 1)
	procSendMessageW.Call(ui.hwndChkAutoStr, WM_SETFONT, ui.fontRegular, 1)
	procSendMessageW.Call(ui.hwndChkAutoWin, WM_SETFONT, ui.fontRegular, 1)
	procSendMessageW.Call(ui.hwndChkAudio, WM_SETFONT, ui.fontRegular, 1)
	procSendMessageW.Call(ui.hwndChkTray, WM_SETFONT, ui.fontRegular, 1)

	// Platzhalter-Text für Edit-Felder (EM_SETCUEBANNER)
	cueRist, _ := syscall.UTF16PtrFromString("Port (z.B. 7890)")
	procSendMessageW.Call(ui.hwndEditRist, EM_SETCUEBANNER, 1, uintptr(unsafe.Pointer(cueRist)))
	cueOut, _ := syscall.UTF16PtrFromString("IP:Port (z.B. 192.168.1.50:10000)")
	procSendMessageW.Call(ui.hwndEditOut, EM_SETCUEBANNER, 1, uintptr(unsafe.Pointer(cueOut)))

	// System Tray Icon initialisieren
	ui.initTray(hwnd)
}

func (ui *UI) onToggleAutoStream() {
	checked, _, _ := procSendMessageW.Call(ui.hwndChkAutoStr, BM_GETCHECK, 0, 0)
	ui.config.AutoStartStream = (checked == BST_CHECKED)
	_ = settings.SaveConfig(ui.config)
	if ui.config.AutoStartStream {
		ui.SetLog("✔ Stream-Autostart bei App-Start aktiviert.")
	} else {
		ui.SetLog("Stream-Autostart deaktiviert.")
	}
}

func (ui *UI) onToggleAutoWin() {
	checked, _, _ := procSendMessageW.Call(ui.hwndChkAutoWin, BM_GETCHECK, 0, 0)
	isWin := (checked == BST_CHECKED)
	ui.config.AutoStartWindows = isWin
	if err := settings.SetWindowsAutoStart(isWin); err != nil {
		ui.SetLog(fmt.Sprintf("Fehler bei Windows-Autostart: %v", err))
		showError(ui.hwnd, "Autostart Fehler", fmt.Sprintf("Konnte Windows Autostart nicht anpassen:\n%v", err))
	} else {
		_ = settings.SaveConfig(ui.config)
		if isWin {
			ui.SetLog("✔ Windows-Autostart erfolgreich eingerichtet!")
		} else {
			ui.SetLog("Windows-Autostart deaktiviert.")
		}
	}
}

func (ui *UI) onToggleAudio() {
	checked, _, _ := procSendMessageW.Call(ui.hwndChkAudio, BM_GETCHECK, 0, 0)
	ui.config.AudioAlerts = (checked == BST_CHECKED)
	_ = settings.SaveConfig(ui.config)
	if ui.config.AudioAlerts {
		ui.SetLog("✔ Audio-Alerts bei Verbindung/Drop aktiviert.")
	} else {
		ui.SetLog("Audio-Alerts deaktiviert.")
	}
}

func (ui *UI) onToggleTray() {
	checked, _, _ := procSendMessageW.Call(ui.hwndChkTray, BM_GETCHECK, 0, 0)
	ui.config.MinimizeToTray = (checked == BST_CHECKED)
	_ = settings.SaveConfig(ui.config)
	if ui.config.MinimizeToTray {
		ui.SetLog("✔ Minimieren ins System-Tray aktiviert.")
	} else {
		ui.SetLog("Minimieren ins System-Tray deaktiviert.")
	}
}

func (ui *UI) StartServer() error {
	if ui.engine.IsRunning() {
		return nil
	}

	ristPort := ui.config.RistPort
	outHost := ui.config.OutHost
	if outHost == "" {
		outHost = "127.0.0.1"
	}
	outPort := ui.config.OutPort
	bufMs := ui.config.BufferMs

	if ui.hwndEditRist != 0 {
		if p, err := strconv.Atoi(getWindowText(ui.hwndEditRist)); err == nil && p > 0 {
			ristPort = p
		}
	}
	if ui.hwndEditOut != 0 {
		if h, p, err := settings.ParseOutTarget(getWindowText(ui.hwndEditOut), outPort); err == nil {
			outHost = h
			outPort = p
		}
	}
	// bufMs kommt direkt aus config (wird von setBufferMs gesetzt)
	if bufMs <= 0 {
		bufMs = 1000
	}

	ui.config.RistPort = ristPort
	ui.config.OutHost = outHost
	ui.config.OutPort = outPort
	ui.config.BufferMs = bufMs
	_ = settings.SaveConfig(ui.config)

	cfg := engine.ServerConfig{
		RistPort:    ristPort,
		OutHost:     outHost,
		OutPort:     outPort,
		BufferMs:    bufMs,
		PacketSize:  1316,
		StatsPeriod: 0.5,
	}

	if ui.hwndEditRist != 0 {
		procEnableWindow.Call(ui.hwndEditRist, 0)
		procEnableWindow.Call(ui.hwndEditOut, 0)
	}
	// Buffer-Pill-Buttons beim laufenden Server deaktivieren
	for _, h := range []uintptr{ui.hwndBtnBuf500, ui.hwndBtnBuf1000, ui.hwndBtnBuf2000, ui.hwndBtnBuf3000} {
		if h != 0 {
			procEnableWindow.Call(h, 0)
		}
	}

	ui.SetLog(fmt.Sprintf("Warte auf Stream auf Port %d... (Puffer: %dms)", ristPort, bufMs))

	go func() {
		if err := ui.engine.Start(cfg); err != nil {
			if ui.hwndEditRist != 0 {
				procEnableWindow.Call(ui.hwndEditRist, 1)
				procEnableWindow.Call(ui.hwndEditOut, 1)
			}
			for _, h := range []uintptr{ui.hwndBtnBuf500, ui.hwndBtnBuf1000, ui.hwndBtnBuf2000, ui.hwndBtnBuf3000} {
				if h != 0 {
					procEnableWindow.Call(h, 1)
				}
			}
			ui.SetLog(fmt.Sprintf("Fehler: %v", err))
			showError(ui.hwnd, "Startfehler", fmt.Sprintf("Konnte RIST Receiver nicht starten:\n%v", err))
		}
	}()

	return nil
}

func (ui *UI) StopServer() {
	if ui.engine.IsRunning() {
		ui.engine.Stop()
		if ui.hwndEditRist != 0 {
			procEnableWindow.Call(ui.hwndEditRist, 1)
			procEnableWindow.Call(ui.hwndEditOut, 1)
		}
		// Buffer-Pill-Buttons wieder aktivieren
		for _, h := range []uintptr{ui.hwndBtnBuf500, ui.hwndBtnBuf1000, ui.hwndBtnBuf2000, ui.hwndBtnBuf3000} {
			if h != 0 {
				procEnableWindow.Call(h, 1)
			}
		}
		ui.SetLog("Server gestoppt. Bereit zum Starten.")
		if ui.hwnd != 0 {
			procInvalidateRect.Call(ui.hwnd, 0, 0)
		}
		if ui.hwndBtnStart != 0 {
			procInvalidateRect.Call(ui.hwndBtnStart, 0, 0)
		}
	}
}

func (ui *UI) ToggleServer() (bool, error) {
	if ui.engine.IsRunning() {
		ui.StopServer()
		return false, nil
	}
	err := ui.StartServer()
	return ui.engine.IsRunning(), err
}

func (ui *UI) onToggleServer() {
	if ui.engine.IsRunning() {
		ui.StopServer()
		return
	}

	ristPortStr := getWindowText(ui.hwndEditRist)
	outTargetStr := getWindowText(ui.hwndEditOut)

	ristPort, err1 := strconv.Atoi(ristPortStr)
	outHost, outPort, err2 := settings.ParseOutTarget(outTargetStr, ui.config.OutPort)

	if err1 != nil || ristPort <= 0 || ristPort > 65535 {
		showError(ui.hwnd, "Ungültiger Port", "Bitte gib einen gültigen RIST Port zwischen 1 und 65535 ein.")
		return
	}
	if err2 != nil || outPort <= 0 || outPort > 65535 {
		showError(ui.hwnd, "Ungültiges Weiterleitungsziel", fmt.Sprintf("Bitte gib ein gültiges Ziel ein (z. B. '10000' oder '192.168.1.50:10000'):\n%v", err2))
		return
	}

	ui.config.RistPort = ristPort
	ui.config.OutHost = outHost
	ui.config.OutPort = outPort
	// ui.config.BufferMs wird bereits von setBufferMs aktuell gehalten
	_ = settings.SaveConfig(ui.config)

	_ = ui.StartServer()
}

func (ui *UI) onCopyStatsLink() {
	link := fmt.Sprintf("http://%s:%d/stats", ui.lanIP, ui.apiPort)
	if setClipboardText(ui.hwnd, link) {
		ui.mu.Lock()
		ui.copiedStatsTimer = 3
		ui.mu.Unlock()
		ui.SetLog("✔ Stats-Link in die Zwischenablage kopiert!")
		if ui.hwndBtnCopy != 0 {
			procInvalidateRect.Call(ui.hwndBtnCopy, 0, 0)
		}
		go func() {
			time.Sleep(2 * time.Second)
			ui.mu.Lock()
			ui.copiedStatsTimer = 0
			ui.mu.Unlock()
			if ui.hwndBtnCopy != 0 {
				procInvalidateRect.Call(ui.hwndBtnCopy, 0, 0)
			}
		}()
	}
}

func (ui *UI) onCopyObsLink() {
	outHost := ui.config.OutHost
	outPort := ui.config.OutPort
	if ui.hwndEditOut != 0 {
		if h, p, err := settings.ParseOutTarget(getWindowText(ui.hwndEditOut), outPort); err == nil {
			outHost = h
			outPort = p
		}
	}
	isLocal := outHost == "" || outHost == "127.0.0.1" || outHost == "localhost"

	var link string
	if isLocal {
		link = fmt.Sprintf("udp://127.0.0.1:%d?pkt_size=1316", outPort)
	} else {
		link = fmt.Sprintf("udp://@:%d?pkt_size=1316", outPort)
	}

	if setClipboardText(ui.hwnd, link) {
		ui.mu.Lock()
		ui.copiedObsTimer = 3
		ui.mu.Unlock()
		if isLocal {
			ui.SetLog("✔ OBS-Link kopiert! (Single-PC: 127.0.0.1)")
		} else {
			ui.SetLog(fmt.Sprintf("✔ 2-PC Link kopiert! Auf OBS-PC (%s) einfügen.", outHost))
		}
		if ui.hwndBtnCopyObs != 0 {
			procInvalidateRect.Call(ui.hwndBtnCopyObs, 0, 0)
		}
		go func() {
			time.Sleep(2 * time.Second)
			ui.mu.Lock()
			ui.copiedObsTimer = 0
			ui.mu.Unlock()
			if ui.hwndBtnCopyObs != 0 {
				procInvalidateRect.Call(ui.hwndBtnCopyObs, 0, 0)
			}
		}()
	}
}

func (ui *UI) onCopyRemoteLink() {
	link := fmt.Sprintf("http://%s:%d/remote", ui.lanIP, ui.apiPort)
	if setClipboardText(ui.hwnd, link) {
		ui.mu.Lock()
		ui.copiedRemoteTimer = 3
		ui.mu.Unlock()
		ui.SetLog("✔ Remote-Link kopiert! Auf dem Smartphone aufrufen.")
		if ui.hwndBtnCopyRemote != 0 {
			procInvalidateRect.Call(ui.hwndBtnCopyRemote, 0, 0)
		}
		go func() {
			time.Sleep(2 * time.Second)
			ui.mu.Lock()
			ui.copiedRemoteTimer = 0
			ui.mu.Unlock()
			if ui.hwndBtnCopyRemote != 0 {
				procInvalidateRect.Call(ui.hwndBtnCopyRemote, 0, 0)
			}
		}()
	}
}

func (ui *UI) onCopyOverlayLink() {
	link := fmt.Sprintf("http://%s:%d/overlay", ui.lanIP, ui.apiPort)
	if setClipboardText(ui.hwnd, link) {
		ui.mu.Lock()
		ui.copiedOverlayTimer = 3
		ui.mu.Unlock()
		ui.SetLog("✔ OBS-Overlay Link kopiert! Als Browser-Quelle in OBS einfügen.")
		if ui.hwndBtnCopyOverlay != 0 {
			procInvalidateRect.Call(ui.hwndBtnCopyOverlay, 0, 0)
		}
		go func() {
			time.Sleep(2 * time.Second)
			ui.mu.Lock()
			ui.copiedOverlayTimer = 0
			ui.mu.Unlock()
			if ui.hwndBtnCopyOverlay != 0 {
				procInvalidateRect.Call(ui.hwndBtnCopyOverlay, 0, 0)
			}
		}()
	}
}

func (ui *UI) openKofiLink() {
	kofiURL, _ := syscall.UTF16PtrFromString("https://ko-fi.com/jumpgeil")
	op, _ := syscall.UTF16PtrFromString("open")
	procShellExecuteW.Call(0, uintptr(unsafe.Pointer(op)), uintptr(unsafe.Pointer(kofiURL)), 0, 0, 1)
}

func (ui *UI) openUpdateLink() {
	ui.mu.Lock()
	url := ui.updateURL
	ui.mu.Unlock()
	if url == "" {
		url = "https://github.com/grandstyle/rist-server/releases"
	}
	urlPtr, _ := syscall.UTF16PtrFromString(url)
	op, _ := syscall.UTF16PtrFromString("open")
	procShellExecuteW.Call(0, uintptr(unsafe.Pointer(op)), uintptr(unsafe.Pointer(urlPtr)), 0, 0, 1)
}

// checkForUpdates fragt im Hintergrund die GitHub API nach Releases ab
func (ui *UI) checkForUpdates() {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/grandstyle/rist-server/releases/latest", nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "RIST-Server/"+version.Version)

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var rel struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return
	}

	remoteVer := strings.TrimPrefix(rel.TagName, "v")
	currentVer := strings.TrimPrefix(version.Version, "v")

	if remoteVer != "" && remoteVer != currentVer {
		// Neuere Version vorhanden
		ui.mu.Lock()
		ui.updateAvailable = true
		ui.updateVersion = rel.TagName
		ui.updateURL = rel.HTMLURL
		ui.statusMsg = fmt.Sprintf("★ Neues Update verfügbar (%s)! Klicke oben zum Herunterladen.", rel.TagName)
		ui.mu.Unlock()

		if ui.hwnd != 0 {
			procPostMessageW.Call(ui.hwnd, WM_USER_LOG, 0, 0)
			procInvalidateRect.Call(ui.hwnd, 0, 0)
		}
	}
}

func (ui *UI) onPaint(hdc uintptr) {
	var rc RECT
	procGetClientRect.Call(ui.hwnd, uintptr(unsafe.Pointer(&rc)))
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top

	// Double Buffering
	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	memBmp, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
	oldBmp, _, _ := procSelectObject.Call(memDC, memBmp)

	// 1. Cyber Background füllen
	procFillRect.Call(memDC, uintptr(unsafe.Pointer(&rc)), ui.brushBg)

	// 2. Header Bereich: Neon Cyberpunk Title
	procSetBkMode.Call(memDC, TRANSPARENT)

	// Cyber Tag & Version
	procSetTextColor.Call(memDC, uintptr(colCyanAccent))
	oldFont, _, _ := procSelectObject.Call(memDC, ui.fontBrand)
	drawText(memDC, "// RIST RECEIVER", 24, 16, 235, 46, DT_LEFT|DT_SINGLELINE)

	// Falls Update verfügbar -> blinkender/auffälliger Update-Badge
	ui.mu.Lock()
	hasUpdate := ui.updateAvailable
	updVer := ui.updateVersion
	ui.mu.Unlock()

	if hasUpdate {
		procSetTextColor.Call(memDC, uintptr(colYellowAccent))
		procSelectObject.Call(memDC, ui.fontSmall)
		drawText(memDC, fmt.Sprintf("▲ UPDATE %s", updVer), 24, 42, 235, 56, DT_LEFT|DT_SINGLELINE)
	}

	// 3. Status Pill oben rechts (Exakt zentrierter Punkt & Text)
	ui.drawCyberStatusBadge(memDC, 246, 14, 174, 34)

	// 4. Port Settings & Options Card (y=58, h=186)
	ui.drawCyberCard(memDC, 24, 58, 396, 186, "STREAM PORTS & OPTIONEN")
	procSetTextColor.Call(memDC, uintptr(colTextWhite))
	procSelectObject.Call(memDC, ui.fontRegular)
	drawText(memDC, "RIST Empfangsport (Sender):", 40, 90, 295, 110, DT_LEFT|DT_SINGLELINE)
	drawText(memDC, "Weiterleitung (Ziel-IP:Port):", 40, 118, 235, 138, DT_LEFT|DT_SINGLELINE)
	drawText(memDC, "RIST Puffer (Jitter/Latenz ms):", 40, 146, 295, 166, DT_LEFT|DT_SINGLELINE)

	// Custom Rahmen für die Edit-Felder (1px außen um das Feld)
	// Die Felder sind ohne WS_EX_CLIENTEDGE, daher malen wir den Border im Parent-DC.
	editRunning := ui.engine.IsRunning()
	editBorderCol := colCyanAccent
	if editRunning {
		editBorderCol = colCardBorder
	}
	editPen := createPen(0, 1, editBorderCol)
	editBgBrush := createSolidBrush(colInputBg)
	oldEditBrush, _, _ := procSelectObject.Call(memDC, editBgBrush)
	oldEditPen, _, _ := procSelectObject.Call(memDC, editPen)
	// RIST Port Rahmen: 1px außen um x=307,y=87,w=98,h=22
	procRoundRect.Call(memDC, 305, 85, 407, 111, 4, 4)
	// Ziel IP:Port Rahmen: 1px außen um x=241,y=115,w=164,h=22
	procRoundRect.Call(memDC, 239, 113, 407, 139, 4, 4)
	procSelectObject.Call(memDC, oldEditBrush)
	procSelectObject.Call(memDC, oldEditPen)
	procDeleteObject.Call(editPen)
	procDeleteObject.Call(editBgBrush)

	// 5. Live Telemetrie Card (y=254, h=252)
	ui.drawCyberCard(memDC, 24, 254, 396, 252, "LIVE DURCHSATZ & TELEMETRIE")
	ui.drawCyberTelemetry(memDC, 24, 254, 396, 252)

	// 6. User-Friendly Live Status Line (y=572)
	ui.mu.Lock()
	statusText := ui.statusMsg
	ui.mu.Unlock()

	procSetBkMode.Call(memDC, TRANSPARENT)
	procSelectObject.Call(memDC, ui.fontRegular)

	// "STATUS:" in Cyan
	procSetTextColor.Call(memDC, uintptr(colCyanAccent))
	drawText(memDC, "STATUS:", 24, 572, 78, 592, DT_LEFT|DT_VCENTER|DT_SINGLELINE)

	// Status Message in Weiß
	procSetTextColor.Call(memDC, uintptr(colTextWhite))
	drawText(memDC, statusText, 82, 572, 420, 592, DT_LEFT|DT_VCENTER|DT_SINGLELINE)

	// 7. Ko-fi Support Link im Footer (y=598)
	ui.drawKofiFooterLink(memDC, 24, 598, 396, 22)

	// Blit auf Bildschirm DC
	procBitBlt.Call(hdc, 0, 0, uintptr(w), uintptr(h), memDC, 0, 0, SRCCOPY)

	// Cleanup
	procSelectObject.Call(memDC, oldFont)
	procSelectObject.Call(memDC, oldBmp)
	procDeleteObject.Call(memBmp)
	procDeleteDC.Call(memDC)
}

func (ui *UI) drawCyberCard(hdc uintptr, x, y, w, h int32, title string) {
	oldBrush, _, _ := procSelectObject.Call(hdc, ui.brushCard)
	oldPen, _, _ := procSelectObject.Call(hdc, ui.penCardBorder)

	procRoundRect.Call(hdc, uintptr(x), uintptr(y), uintptr(x+w), uintptr(y+h), 8, 8)

	// Cyan Top-Accent Line
	cyanPen := createPen(0, 2, colCyanAccent)
	procSelectObject.Call(hdc, cyanPen)
	procRectangle.Call(hdc, uintptr(x+14), uintptr(y), uintptr(x+60), uintptr(y+2))

	// Section Title
	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(colCyanAccent))
	procSelectObject.Call(hdc, ui.fontSection)
	drawText(hdc, title, x+16, y+8, x+w-16, y+24, DT_LEFT|DT_SINGLELINE)

	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(cyanPen)
}

func (ui *UI) drawCyberStatusBadge(hdc uintptr, x, y, w, h int32) {
	ui.mu.Lock()
	st := ui.stats.State
	ui.mu.Unlock()

	var bgCol, borderCol, dotCol, textCol uint32
	var statusText string

	switch st {
	case telemetry.StateStreaming:
		bgCol = rgb(4, 38, 20)
		borderCol = colGreenAccent
		dotCol = colGreenAccent
		textCol = colGreenAccent
		statusText = "LIVE STREAMING"
	case telemetry.StateListening:
		bgCol = rgb(40, 34, 0)
		borderCol = colYellowAccent
		dotCol = colYellowAccent
		textCol = colYellowAccent
		statusText = "WARTE AUF STREAM..."
	default:
		bgCol = rgb(38, 4, 16)
		borderCol = colPinkAccent
		dotCol = colPinkAccent
		textCol = colPinkAccent
		statusText = "OFFLINE"
	}

	pillBrush := createSolidBrush(bgCol)
	pillPen := createPen(0, 2, borderCol)

	oldBrush, _, _ := procSelectObject.Call(hdc, pillBrush)
	oldPen, _, _ := procSelectObject.Call(hdc, pillPen)
	procRoundRect.Call(hdc, uintptr(x), uintptr(y), uintptr(x+w), uintptr(y+h), 6, 6)

	// Berechne exakte Zentrierung von Punkt + Text
	procSelectObject.Call(hdc, ui.fontSection)
	txtPtr, _ := syscall.UTF16PtrFromString(statusText)
	var textSize SIZE
	procGetTextExtentPoint32W.Call(
		hdc,
		uintptr(unsafe.Pointer(txtPtr)),
		uintptr(len([]rune(statusText))),
		uintptr(unsafe.Pointer(&textSize)),
	)

	dotDiameter := int32(8)
	gap := int32(8)
	totalContentW := dotDiameter + gap + textSize.CX

	contentStartX := x + (w-totalContentW)/2
	dotY := y + (h-dotDiameter)/2

	// Glowing Dot
	dotBrush := createSolidBrush(dotCol)
	dotPen := createPen(0, 1, dotCol)
	procSelectObject.Call(hdc, dotBrush)
	procSelectObject.Call(hdc, dotPen)
	procEllipse.Call(hdc, uintptr(contentStartX), uintptr(dotY), uintptr(contentStartX+dotDiameter), uintptr(dotY+dotDiameter))

	// Status Text
	textX := contentStartX + dotDiameter + gap
	textY := y + (h-textSize.CY)/2
	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(textCol))
	drawText(hdc, statusText, textX, textY, textX+textSize.CX+4, textY+textSize.CY+2, DT_LEFT|DT_SINGLELINE)

	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(pillBrush)
	procDeleteObject.Call(pillPen)
	procDeleteObject.Call(dotBrush)
	procDeleteObject.Call(dotPen)
}

func (ui *UI) drawCyberTelemetry(hdc uintptr, x, y, w, h int32) {
	ui.mu.Lock()
	s := ui.stats
	ui.mu.Unlock()

	colW := (w - 32) / 2

	rowLabelY := y + 24
	rowValueY := y + 38
	rowSubY := y + 64

	// Column 1: Live Bitrate
	c1X := x + 16

	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(colTextMuted))
	procSelectObject.Call(hdc, ui.fontSmall)
	drawText(hdc, "AKTUELLE BITRATE", c1X, rowLabelY, c1X+colW, rowLabelY+14, DT_LEFT|DT_SINGLELINE)

	bitrateColor := colTextMuted
	if s.State == telemetry.StateStreaming && s.BitrateKbps > 0 {
		bitrateColor = colGreenAccent
	}
	procSetTextColor.Call(hdc, uintptr(bitrateColor))
	procSelectObject.Call(hdc, ui.fontLarge)
	drawText(hdc, s.BitrateStr, c1X, rowValueY, c1X+colW, rowValueY+26, DT_LEFT|DT_SINGLELINE)

	procSetTextColor.Call(hdc, uintptr(colTextMuted))
	procSelectObject.Call(hdc, ui.fontSmall)
	fpsText := fmt.Sprintf("Framerate: %s FPS", s.Fps)
	drawText(hdc, fpsText, c1X, rowSubY, c1X+colW, rowSubY+14, DT_LEFT|DT_SINGLELINE)

	// Column 2: Total Received
	c2X := x + 16 + colW + 8

	procSetTextColor.Call(hdc, uintptr(colTextMuted))
	procSelectObject.Call(hdc, ui.fontSmall)
	drawText(hdc, "GESAMT EMPFANGEN", c2X, rowLabelY, c2X+colW, rowLabelY+14, DT_LEFT|DT_SINGLELINE)

	procSetTextColor.Call(hdc, uintptr(colCyanAccent))
	procSelectObject.Call(hdc, ui.fontLarge)
	drawText(hdc, s.TotalFormatted, c2X, rowValueY, c2X+colW, rowValueY+26, DT_LEFT|DT_SINGLELINE)

	procSetTextColor.Call(hdc, uintptr(colTextMuted))
	procSelectObject.Call(hdc, ui.fontSmall)
	durationText := fmt.Sprintf("Laufzeit: %s", s.Duration)
	drawText(hdc, durationText, c2X, rowSubY, c2X+colW, rowSubY+14, DT_LEFT|DT_SINGLELINE)

	// Row 3: Container 1 - Stats Link (NOALBS / Browser)
	boxX := x + 12
	boxW := w - 24
	boxH := int32(38)

	// Container 1: OBS Stream Link (Medienquelle)
	box1Y := y + 82
	boxBg1 := createSolidBrush(rgb(7, 9, 15))
	boxBorder1 := createPen(0, 1, rgb(30, 42, 64))
	oldBrush1, _, _ := procSelectObject.Call(hdc, boxBg1)
	oldPen1, _, _ := procSelectObject.Call(hdc, boxBorder1)
	procRoundRect.Call(hdc, uintptr(boxX), uintptr(box1Y), uintptr(boxX+boxW), uintptr(box1Y+boxH), 6, 6)
	procSelectObject.Call(hdc, oldBrush1)
	procSelectObject.Call(hdc, oldPen1)
	procDeleteObject.Call(boxBg1)
	procDeleteObject.Call(boxBorder1)

	outHost := ui.config.OutHost
	outPort := ui.config.OutPort
	if ui.hwndEditOut != 0 {
		if h, p, err := settings.ParseOutTarget(getWindowText(ui.hwndEditOut), outPort); err == nil {
			outHost = h
			outPort = p
		}
	}
	isLocal := outHost == "" || outHost == "127.0.0.1" || outHost == "localhost"

	var headerText, obsLink string
	if isLocal {
		headerText = "🎥 OBS STREAM LINK (GLEICHER PC)"
		obsLink = fmt.Sprintf("udp://127.0.0.1:%d?pkt_size=1316", outPort)
	} else {
		headerText = "🎥 OBS STREAM LINK (ZIEL-PC)"
		obsLink = fmt.Sprintf("udp://@:%d?pkt_size=1316", outPort)
	}

	procSetTextColor.Call(hdc, uintptr(colGreenAccent))
	procSelectObject.Call(hdc, ui.fontSmall)
	drawText(hdc, headerText, boxX+8, box1Y+3, boxX+boxW-44, box1Y+17, DT_CENTER|DT_SINGLELINE)

	procSetTextColor.Call(hdc, uintptr(colCyanAccent))
	procSelectObject.Call(hdc, ui.fontMedium)
	drawText(hdc, obsLink, boxX+8, box1Y+18, boxX+boxW-44, box1Y+35, DT_CENTER|DT_SINGLELINE)

	// Container 2: OBS Overlay (Browserquelle)
	box2Y := y + 124
	boxBg2 := createSolidBrush(rgb(7, 9, 15))
	boxBorder2 := createPen(0, 1, rgb(30, 42, 64))
	oldBrush2, _, _ := procSelectObject.Call(hdc, boxBg2)
	oldPen2, _, _ := procSelectObject.Call(hdc, boxBorder2)
	procRoundRect.Call(hdc, uintptr(boxX), uintptr(box2Y), uintptr(boxX+boxW), uintptr(box2Y+boxH), 6, 6)
	procSelectObject.Call(hdc, oldBrush2)
	procSelectObject.Call(hdc, oldPen2)
	procDeleteObject.Call(boxBg2)
	procDeleteObject.Call(boxBorder2)

	procSetTextColor.Call(hdc, uintptr(colOrangeAccent))
	procSelectObject.Call(hdc, ui.fontSmall)
	drawText(hdc, "🖥️ OBS OVERLAY (BROWSERQUELLE)", boxX+8, box2Y+3, boxX+boxW-44, box2Y+17, DT_CENTER|DT_SINGLELINE)

	overlayLink := fmt.Sprintf("http://%s:%d/overlay", ui.lanIP, ui.apiPort)
	procSetTextColor.Call(hdc, uintptr(colCyanAccent))
	procSelectObject.Call(hdc, ui.fontMedium)
	drawText(hdc, overlayLink, boxX+8, box2Y+18, boxX+boxW-44, box2Y+35, DT_CENTER|DT_SINGLELINE)

	// Container 3: Stats Link (NOALBS / Browser)
	box3Y := y + 166
	boxBg3 := createSolidBrush(rgb(7, 9, 15))
	boxBorder3 := createPen(0, 1, rgb(30, 42, 64))
	oldBrush3, _, _ := procSelectObject.Call(hdc, boxBg3)
	oldPen3, _, _ := procSelectObject.Call(hdc, boxBorder3)
	procRoundRect.Call(hdc, uintptr(boxX), uintptr(box3Y), uintptr(boxX+boxW), uintptr(box3Y+boxH), 6, 6)
	procSelectObject.Call(hdc, oldBrush3)
	procSelectObject.Call(hdc, oldPen3)
	procDeleteObject.Call(boxBg3)
	procDeleteObject.Call(boxBorder3)

	procSetTextColor.Call(hdc, uintptr(colYellowAccent))
	procSelectObject.Call(hdc, ui.fontSmall)
	drawText(hdc, "⚡ STATS LINK (NOALBS / BROWSER)", boxX+8, box3Y+3, boxX+boxW-44, box3Y+17, DT_CENTER|DT_SINGLELINE)

	apiLink := fmt.Sprintf("http://%s:%d/stats", ui.lanIP, ui.apiPort)
	procSetTextColor.Call(hdc, uintptr(colCyanAccent))
	procSelectObject.Call(hdc, ui.fontMedium)
	drawText(hdc, apiLink, boxX+8, box3Y+18, boxX+boxW-44, box3Y+35, DT_CENTER|DT_SINGLELINE)

	// Container 4: Mobile Remote (Smartphone)
	box4Y := y + 208
	boxBg4 := createSolidBrush(rgb(7, 9, 15))
	boxBorder4 := createPen(0, 1, rgb(30, 42, 64))
	oldBrush4, _, _ := procSelectObject.Call(hdc, boxBg4)
	oldPen4, _, _ := procSelectObject.Call(hdc, boxBorder4)
	procRoundRect.Call(hdc, uintptr(boxX), uintptr(box4Y), uintptr(boxX+boxW), uintptr(box4Y+boxH), 6, 6)
	procSelectObject.Call(hdc, oldBrush4)
	procSelectObject.Call(hdc, oldPen4)
	procDeleteObject.Call(boxBg4)
	procDeleteObject.Call(boxBorder4)

	procSetTextColor.Call(hdc, uintptr(colPinkAccent))
	procSelectObject.Call(hdc, ui.fontSmall)
	drawText(hdc, "📱 MOBILE FERNBEDIENUNG", boxX+8, box4Y+3, boxX+boxW-44, box4Y+17, DT_CENTER|DT_SINGLELINE)

	remoteLink := fmt.Sprintf("http://%s:%d/remote", ui.lanIP, ui.apiPort)
	procSetTextColor.Call(hdc, uintptr(colCyanAccent))
	procSelectObject.Call(hdc, ui.fontMedium)
	drawText(hdc, remoteLink, boxX+8, box4Y+18, boxX+boxW-44, box4Y+35, DT_CENTER|DT_SINGLELINE)
}

func (ui *UI) drawCopyButton(dis *DRAWITEMSTRUCT, copied bool) {
	hdc := dis.HDC
	rc := dis.RcItem

	var bgCol, borderCol uint32

	if copied {
		bgCol = rgb(0, 100, 50)
		borderCol = colGreenAccent
	} else {
		bgCol = rgb(20, 28, 44)
		borderCol = colCyanAccent
	}

	// Ecken mit Card-Hintergrund füllen (verhindert weiße Artefakte bei RoundRect)
	bgFillBrush := createSolidBrush(colCardBg)
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), bgFillBrush)
	procDeleteObject.Call(bgFillBrush)

	btnBrush := createSolidBrush(bgCol)
	btnPen := createPen(0, 1, borderCol)

	oldBrush, _, _ := procSelectObject.Call(hdc, btnBrush)
	oldPen, _, _ := procSelectObject.Call(hdc, btnPen)
	procRoundRect.Call(hdc, uintptr(rc.Left), uintptr(rc.Top), uintptr(rc.Right), uintptr(rc.Bottom), 6, 6)

	cx := (rc.Left + rc.Right) / 2
	cy := (rc.Top + rc.Bottom) / 2

	if copied {
		// Grüner Vektor-Häkchen (Checkmark)
		checkPen := createPen(0, 2, colGreenAccent)
		procSelectObject.Call(hdc, checkPen)
		pts := [3]POINT{
			{X: cx - 5, Y: cy},
			{X: cx - 1, Y: cy + 4},
			{X: cx + 6, Y: cy - 4},
		}
		procPolyline.Call(hdc, uintptr(unsafe.Pointer(&pts[0])), 3)
		procDeleteObject.Call(checkPen)
	} else {
		// Vektor Copy-Icon (Zwei überlappende Dokumente)
		iconPen := createPen(0, 1, colCyanAccent)
		procSelectObject.Call(hdc, iconPen)

		// 1. Hinteres Dokument (oben rechts)
		procRoundRect.Call(hdc, uintptr(cx-2), uintptr(cy-7), uintptr(cx+8), uintptr(cy+4), 2, 2)

		// 2. Vorderes Dokument (unten links mit Button-Hintergrund gefüllt)
		procRoundRect.Call(hdc, uintptr(cx-7), uintptr(cy-3), uintptr(cx+3), uintptr(cy+8), 2, 2)

		// 3. Subtile Content-Linien auf dem vorderen Dokument
		linePen := createPen(0, 1, rgb(0, 180, 210))
		procSelectObject.Call(hdc, linePen)
		l1 := [2]POINT{{X: cx - 4, Y: cy + 1}, {X: cx, Y: cy + 1}}
		procPolyline.Call(hdc, uintptr(unsafe.Pointer(&l1[0])), 2)
		l2 := [2]POINT{{X: cx - 4, Y: cy + 4}, {X: cx - 1, Y: cy + 4}}
		procPolyline.Call(hdc, uintptr(unsafe.Pointer(&l2[0])), 2)

		procDeleteObject.Call(iconPen)
		procDeleteObject.Call(linePen)
	}

	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(btnBrush)
	procDeleteObject.Call(btnPen)
}

func (ui *UI) drawKofiFooterLink(hdc uintptr, x, y, w, h int32) {
	procSetBkMode.Call(hdc, TRANSPARENT)

	// Zentrierter Ko-fi Support Link
	procSetTextColor.Call(hdc, uintptr(colPinkAccent))
	oldFont, _, _ := procSelectObject.Call(hdc, ui.fontSection)

	drawText(hdc, "♥ SUPPORT:  KO-FI.COM/JUMPGEIL", x, y, x+w, y+h, DT_CENTER|DT_VCENTER|DT_SINGLELINE)

	procSelectObject.Call(hdc, oldFont)
}

func (ui *UI) drawCyberButton(dis *DRAWITEMSTRUCT) {
	hdc := dis.HDC
	rc := dis.RcItem

	running := ui.engine.IsRunning()

	var bgCol, borderCol, textCol uint32
	var label string

	if running {
		bgCol = rgb(180, 0, 60)
		borderCol = colPinkAccent
		textCol = rgb(255, 255, 255)
		label = "⏹   STREAM STOPPEN"
	} else {
		bgCol = rgb(0, 140, 75)
		borderCol = colGreenAccent
		textCol = rgb(255, 255, 255)
		label = "▶   STREAM EMPFANGEN"
	}

	btnBrush := createSolidBrush(bgCol)
	btnPen := createPen(0, 2, borderCol)

	oldBrush, _, _ := procSelectObject.Call(hdc, btnBrush)
	oldPen, _, _ := procSelectObject.Call(hdc, btnPen)

	// Hintergrund-Fill gegen weiße Ecken beim RoundRect
	bgFillBrush := createSolidBrush(colBg)
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), bgFillBrush)
	procDeleteObject.Call(bgFillBrush)

	procRoundRect.Call(hdc, uintptr(rc.Left), uintptr(rc.Top), uintptr(rc.Right), uintptr(rc.Bottom), 8, 8)

	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(textCol))
	oldFont, _, _ := procSelectObject.Call(hdc, ui.fontMedium)

	drawText(hdc, label, rc.Left, rc.Top, rc.Right, rc.Bottom, DT_CENTER|DT_VCENTER|DT_SINGLELINE)

	procSelectObject.Call(hdc, oldFont)
	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(btnBrush)
	procDeleteObject.Call(btnPen)
}

// setBufferMs setzt den Buffer-Wert, speichert ihn und aktualisiert die UI.
func (ui *UI) setBufferMs(ms int) {
	ui.mu.Lock()
	ui.config.BufferMs = ms
	ui.mu.Unlock()
	_ = settings.SaveConfig(ui.config)
	// Alle Pill-Buttons neu zeichnen (aktive Auswahl wechselt)
	for _, h := range []uintptr{ui.hwndBtnBuf500, ui.hwndBtnBuf1000, ui.hwndBtnBuf2000, ui.hwndBtnBuf3000} {
		if h != 0 {
			procInvalidateRect.Call(h, 0, 0)
		}
	}
	ui.SetLog(fmt.Sprintf("✔ RIST Puffer auf %dms gesetzt.", ms))
}

// drawBufferPill zeichnet einen einzelnen Buffer-Auswahl-Button (Pill-Stil).
func (ui *UI) drawBufferPill(dis *DRAWITEMSTRUCT, ms int) {
	hdc := dis.HDC
	rc := dis.RcItem

	ui.mu.Lock()
	active := ui.config.BufferMs == ms
	ui.mu.Unlock()

	disabled := (dis.ItemState & 0x0004) != 0 // ODS_DISABLED

	var bgCol, borderCol, textCol uint32
	if disabled {
		// Beim laufenden Stream: alles gedimmt, aktiver Button nur minimal heller
		if active {
			bgCol = rgb(10, 28, 36)
			borderCol = rgb(30, 65, 80)
			textCol = rgb(60, 100, 120)
		} else {
			bgCol = rgb(10, 14, 20)
			borderCol = rgb(20, 28, 40)
			textCol = rgb(40, 55, 75)
		}
	} else if active {
		bgCol = rgb(0, 60, 80)
		borderCol = colCyanAccent
		textCol = colCyanAccent
	} else {
		if (dis.ItemState & 0x0001) != 0 { // ODS_SELECTED (hover/pressed)
			bgCol = rgb(20, 30, 46)
			borderCol = colCyanAccent
		} else {
			bgCol = rgb(14, 18, 28)
			borderCol = colCardBorder
		}
		textCol = colTextMuted
	}

	// Erst den gesamten Rect mit Card-Hintergrundfarbe füllen,
	// damit keine weißen Ecken hinter dem RoundRect sichtbar sind.
	bgFillBrush := createSolidBrush(colCardBg)
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), bgFillBrush)
	procDeleteObject.Call(bgFillBrush)

	brush := createSolidBrush(bgCol)
	pen := createPen(0, 1, borderCol)

	oldBrush, _, _ := procSelectObject.Call(hdc, brush)
	oldPen, _, _ := procSelectObject.Call(hdc, pen)
	procRoundRect.Call(hdc, uintptr(rc.Left), uintptr(rc.Top), uintptr(rc.Right), uintptr(rc.Bottom), 6, 6)

	label := fmt.Sprintf("%d", ms)
	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(textCol))
	oldFont, _, _ := procSelectObject.Call(hdc, ui.fontSmall)
	drawText(hdc, label, rc.Left, rc.Top, rc.Right, rc.Bottom, DT_CENTER|DT_VCENTER|DT_SINGLELINE)

	procSelectObject.Call(hdc, oldFont)
	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(brush)
	procDeleteObject.Call(pen)
}


// Clipboard Helper für Windows
func setClipboardText(hwnd uintptr, text string) bool {
	r, _, _ := procOpenClipboard.Call(hwnd)
	if r == 0 {
		return false
	}
	defer procCloseClipboard.Call()

	procEmptyClipboard.Call()

	utf16Data, err := syscall.UTF16FromString(text)
	if err != nil {
		return false
	}

	byteLen := uintptr(len(utf16Data) * 2)
	hMem, _, _ := procGlobalAlloc.Call(GMEM_MOVEABLE, byteLen)
	if hMem == 0 {
		return false
	}

	ptr, _, _ := procGlobalLock.Call(hMem)
	if ptr == 0 {
		return false
	}

	copy((*[1 << 30]byte)(unsafe.Pointer(ptr))[:byteLen], (*[1 << 30]byte)(unsafe.Pointer(&utf16Data[0]))[:byteLen])
	procGlobalUnlock.Call(hMem)

	res, _, _ := procSetClipboardData.Call(CF_UNICODETEXT, hMem)
	return res != 0
}

// Hilfsfunktionen
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

func getWindowText(hwnd uintptr) string {
	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(length+1))
	return syscall.UTF16ToString(buf)
}

func showError(hwnd uintptr, title, msg string) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	msgPtr, _ := syscall.UTF16PtrFromString(msg)
	procMessageBoxW.Call(hwnd, uintptr(unsafe.Pointer(msgPtr)), uintptr(unsafe.Pointer(titlePtr)), 0x00000010)
}

type confirmDialogState struct {
	confirmed    bool
	done         bool
	isStreamLive bool
	btnExit      uintptr
	btnCancel    uintptr
}

var (
	confirmClassRegistered bool
	activeConfirmState     *confirmDialogState
)

func confirmDlgWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	ui := globalUI

	switch msg {
	case WM_CREATE:
		btnClass, _ := syscall.UTF16PtrFromString("BUTTON")
		btnExitText, _ := syscall.UTF16PtrFromString("BEENDEN")
		btnCancelText, _ := syscall.UTF16PtrFromString("ABBRECHEN")

		hExit, _, _ := procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(btnClass)),
			uintptr(unsafe.Pointer(btnExitText)),
			WS_CHILD|WS_VISIBLE|BS_OWNERDRAW|WS_TABSTOP,
			24, 86, 160, 38,
			hwnd,
			uintptr(ID_CONFIRM_EXIT),
			0, 0,
		)

		hCancel, _, _ := procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(btnClass)),
			uintptr(unsafe.Pointer(btnCancelText)),
			WS_CHILD|WS_VISIBLE|BS_OWNERDRAW|WS_TABSTOP,
			206, 86, 160, 38,
			hwnd,
			uintptr(ID_CONFIRM_CANCEL),
			0, 0,
		)

		if activeConfirmState != nil {
			activeConfirmState.btnExit = hExit
			activeConfirmState.btnCancel = hCancel
		}
		return 0

	case WM_DRAWITEM:
		dis := (*DRAWITEMSTRUCT)(unsafe.Pointer(lParam))
		hdc := dis.HDC
		rc := dis.RcItem

		var bgCol, borderCol, textCol uint32
		var label string

		if dis.CtlID == ID_CONFIRM_EXIT {
			label = "BEENDEN"
			if (dis.ItemState & 0x0001) != 0 { // ODS_SELECTED
				bgCol = rgb(90, 10, 30)
				borderCol = colPinkAccent
			} else {
				bgCol = rgb(50, 8, 22)
				borderCol = colPinkAccent
			}
			textCol = colTextWhite
		} else {
			label = "ABBRECHEN"
			if (dis.ItemState & 0x0001) != 0 {
				bgCol = rgb(30, 42, 64)
				borderCol = colCyanAccent
			} else {
				bgCol = rgb(14, 20, 32)
				borderCol = colCardBorder
			}
			textCol = colCyanAccent
		}

		btnBrush := createSolidBrush(bgCol)
		btnPen := createPen(0, 2, borderCol)
		oldBrush, _, _ := procSelectObject.Call(hdc, btnBrush)
		oldPen, _, _ := procSelectObject.Call(hdc, btnPen)
		procRoundRect.Call(hdc, uintptr(rc.Left), uintptr(rc.Top), uintptr(rc.Right), uintptr(rc.Bottom), 8, 8)
		procSelectObject.Call(hdc, oldBrush)
		procSelectObject.Call(hdc, oldPen)
		procDeleteObject.Call(btnBrush)
		procDeleteObject.Call(btnPen)

		procSetBkMode.Call(hdc, TRANSPARENT)
		procSetTextColor.Call(hdc, uintptr(textCol))
		var oldFont uintptr
		if ui != nil {
			oldFont, _, _ = procSelectObject.Call(hdc, ui.fontMedium)
		}
		drawText(hdc, label, rc.Left, rc.Top, rc.Right, rc.Bottom, DT_CENTER|DT_VCENTER|DT_SINGLELINE)
		if oldFont != 0 {
			procSelectObject.Call(hdc, oldFont)
		}
		return 1

	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

		var rc RECT
		procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))

		bgBrush := createSolidBrush(colCardBg)
		procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), bgBrush)
		procDeleteObject.Call(bgBrush)

		procSetBkMode.Call(hdc, TRANSPARENT)

		// Warning Title & Text
		if activeConfirmState != nil && activeConfirmState.isStreamLive {
			procSetTextColor.Call(hdc, uintptr(colPinkAccent))
			if ui != nil {
				oldFont, _, _ := procSelectObject.Call(hdc, ui.fontSection)
				drawText(hdc, "⚠️  STREAM AKTIV", 24, 18, rc.Right-24, 40, DT_LEFT|DT_SINGLELINE)
				procSelectObject.Call(hdc, oldFont)

				procSetTextColor.Call(hdc, uintptr(colTextWhite))
				oldFont2, _, _ := procSelectObject.Call(hdc, ui.fontRegular)
				drawText(hdc, "Der Stream läuft noch. Wirklich trennen und beenden?", 24, 46, rc.Right-24, 76, DT_LEFT|DT_SINGLELINE)
				procSelectObject.Call(hdc, oldFont2)
			}
		} else {
			procSetTextColor.Call(hdc, uintptr(colYellowAccent))
			if ui != nil {
				oldFont, _, _ := procSelectObject.Call(hdc, ui.fontSection)
				drawText(hdc, "⚠️  SERVER AKTIV", 24, 18, rc.Right-24, 40, DT_LEFT|DT_SINGLELINE)
				procSelectObject.Call(hdc, oldFont)

				procSetTextColor.Call(hdc, uintptr(colTextWhite))
				oldFont2, _, _ := procSelectObject.Call(hdc, ui.fontRegular)
				drawText(hdc, "Der Server läuft noch. Wirklich beenden?", 24, 46, rc.Right-24, 76, DT_LEFT|DT_SINGLELINE)
				procSelectObject.Call(hdc, oldFont2)
			}
		}

		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_COMMAND:
		ctrlID := int(wParam & 0xFFFF)
		if ctrlID == ID_CONFIRM_EXIT {
			if activeConfirmState != nil {
				activeConfirmState.confirmed = true
				activeConfirmState.done = true
			}
			procDestroyWindow.Call(hwnd)
			return 0
		} else if ctrlID == ID_CONFIRM_CANCEL {
			if activeConfirmState != nil {
				activeConfirmState.confirmed = false
				activeConfirmState.done = true
			}
			procDestroyWindow.Call(hwnd)
			return 0
		}

	case WM_CLOSE:
		if activeConfirmState != nil {
			activeConfirmState.confirmed = false
			activeConfirmState.done = true
		}
		procDestroyWindow.Call(hwnd)
		return 0

	case WM_DESTROY:
		if activeConfirmState != nil {
			activeConfirmState.done = true
		}
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func (ui *UI) showConfirmExitDialog() bool {
	dlgClassName, _ := syscall.UTF16PtrFromString("RISTConfirmExitClass")
	if !confirmClassRegistered {
		wcex := WNDCLASSEXW{
			CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
			Style:         0,
			LpfnWndProc:   syscall.NewCallback(confirmDlgWndProc),
			HInstance:     0,
			HCursor:       ui.hCursorArrow,
			HbrBackground: ui.brushCard,
			LpszClassName: dlgClassName,
			HIcon:         ui.hIcon,
			HIconSm:       ui.hIconSm,
		}
		procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wcex)))
		confirmClassRegistered = true
	}

	stats := ui.telemetry.GetStats()
	isStreamLive := (stats.State == telemetry.StateStreaming && stats.BitrateKbps > 0)

	activeConfirmState = &confirmDialogState{
		isStreamLive: isStreamLive,
	}

	// Center dialog on main window
	var parentRc RECT
	procGetWindowRect.Call(ui.hwnd, uintptr(unsafe.Pointer(&parentRc)))
	dlgW := int32(390)
	dlgH := int32(175)
	dlgX := parentRc.Left + (parentRc.Right-parentRc.Left-dlgW)/2
	dlgY := parentRc.Top + (parentRc.Bottom-parentRc.Top-dlgH)/2
	if dlgX < 50 {
		dlgX = 50
	}
	if dlgY < 50 {
		dlgY = 50
	}

	titleText := "Server beenden?"
	if isStreamLive {
		titleText = "Stream beenden?"
	}
	dlgTitle, _ := syscall.UTF16PtrFromString(titleText)
	hDlg, _, _ := procCreateWindowExW.Call(
		0x00000001, // WS_EX_DLGMODALFRAME
		uintptr(unsafe.Pointer(dlgClassName)),
		uintptr(unsafe.Pointer(dlgTitle)),
		WS_POPUP|WS_CAPTION|WS_SYSMENU|WS_VISIBLE,
		uintptr(dlgX), uintptr(dlgY), uintptr(dlgW), uintptr(dlgH),
		ui.hwnd,
		0, 0, 0,
	)

	if hDlg == 0 {
		return true
	}

	// Dark Mode für Titelleiste
	darkMode := int32(1)
	procDwmSetWindowAttribute.Call(
		hDlg,
		DWMWA_USE_IMMERSIVE_DARK_MODE,
		uintptr(unsafe.Pointer(&darkMode)),
		4,
	)

	procEnableWindow.Call(ui.hwnd, 0)
	defer func() {
		procEnableWindow.Call(ui.hwnd, 1)
		procSetForegroundWindow.Call(ui.hwnd)
	}()

	procShowWindow.Call(hDlg, SW_SHOW)
	procSetForegroundWindow.Call(hDlg)

	// Modale Ereignisschleife
	var msg MSG
	for !activeConfirmState.done {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if r == 0 || int32(r) == -1 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	return activeConfirmState.confirmed
}
