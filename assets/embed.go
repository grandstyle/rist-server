package assets

import (
	_ "embed"
)

//go:embed ffmpeg.exe
var FFmpegBinary []byte

//go:embed app.ico
var AppIcon []byte

//go:embed icon-192.png
var Icon192 []byte

//go:embed icon-512.png
var Icon512 []byte

//go:embed icon-maskable-192.png
var IconMaskable192 []byte

//go:embed icon-maskable-512.png
var IconMaskable512 []byte
