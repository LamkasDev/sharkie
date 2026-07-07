package lib

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterMinecraftStubs() {
	if logger.GameDebugMode {
		elf.RegisterStub("Minecraft.Client.sprx", "sub_20280", sub_20280)
		elf.RegisterStub("Minecraft.Client.sprx", "sub_7510F0", sub_7510F0)
		elf.RegisterStub("Minecraft.Client.sprx", "sub_133DD0", sub_133DD0)
	}
}

var cFormatRegex = regexp.MustCompile(`%([0-9#+\- ]*)l{0,2}([a-zA-Z%])`)
var verbCountRegex = regexp.MustCompile(`%[^%]|%[a-zA-Z]`)

func sub_20280(_, stringPtr, a, b, c, d, e, f, g, h uintptr) uintptr {
	rawFormat := GoString(Cstring(stringPtr))
	cleanFormat := cFormatRegex.ReplaceAllString(rawFormat, "%$1$2")
	cleanFormat = strings.ReplaceAll(cleanFormat, "\n", "")

	args := []interface{}{a, b, c, d, e, f, g, h}
	matches := verbCountRegex.FindAllStringIndex(cleanFormat, -1)
	argCount := len(matches)
	if argCount > len(args) {
		argCount = len(args)
	}

	formattedContent := fmt.Sprintf(cleanFormat, args[:argCount]...)
	logger.Printf("[%s/DEBUG] %s\n", emu.GetCurrentThread().Name, color.Cyan.Sprint(formattedContent))

	return 0
}

func sub_7510F0(stringPtr, a, b, c, d, e, f, g, h, i uintptr) uintptr {
	rawFormat := GoString(Cstring(stringPtr))
	cleanFormat := cFormatRegex.ReplaceAllString(rawFormat, "%$1$2")
	cleanFormat = strings.ReplaceAll(cleanFormat, "\n", "")

	args := []interface{}{a, b, c, d, e, f, g, h}
	matches := verbCountRegex.FindAllStringIndex(cleanFormat, -1)
	argCount := len(matches)
	if argCount > len(args) {
		argCount = len(args)
	}

	formattedContent := fmt.Sprintf(cleanFormat, args[:argCount]...)
	logger.Printf("[%s/DEBUG] %s\n", emu.GetCurrentThread().Name, color.Cyan.Sprint(formattedContent))

	return 0
}

func sub_133DD0(_, stringPtr uintptr) uintptr {
	logger.Printf("[%s/DEBUG] %s\n", emu.GetCurrentThread().Name, color.Cyan.Sprint(GoString(Cstring(stringPtr))))
	return 0
}
