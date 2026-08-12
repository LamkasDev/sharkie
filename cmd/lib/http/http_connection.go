package http

import (
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/http"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000113E0
// __int64 __fastcall sceHttpCreateConnectionWithURL(unsigned int, __int64, unsigned int)
func libSceHttp_sceHttpCreateConnectionWithURL(templateId uint32, url Cstring, enableKeepAlive uint32) uintptr {
	template := GlobalHttpHandler.GetTemplate(templateId)
	if template == nil {
		logger.Printf("%-132s %s failed due to invalid template id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpCreateConnectionWithURL"),
		)
		return 0x80431100
	}
	if url == nil {
		logger.Printf("%-132s %s failed due to invalid url.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpCreateConnectionWithURL"),
		)
		return 0x80433060
	}

	// Parse URI (first get required size, then parse with allocated pool).
	var required uint64
	err := libSceHttp_sceHttpUriParse(nil, url, 0, &required, 0)
	if err != 0 {
		return err
	}
	var pool []byte
	var poolPtr uintptr
	if required > 0 {
		pool = make([]byte, required)
		poolPtr = uintptr(unsafe.Pointer(&pool[0]))
	}
	var parsed HttpUriElement
	err = libSceHttp_sceHttpUriParse(&parsed, url, poolPtr, &required, required)
	if err != 0 {
		return err
	}

	// Validate scheme and hostname.
	schemeString := GoString(parsed.Scheme)
	isSecure, err := checkScheme(schemeString)
	if err != 0 {
		return err
	}
	hostString := GoString(parsed.Hostname)
	if hostString == "" {
		logger.Printf("%-132s %s failed due to invalid hostname.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpCreateConnectionWithURL"),
		)
		return 0x80433060
	}

	// Create connection.
	connection := GlobalHttpHandler.CreateConnection()
	connection.TemplateId = templateId
	connection.Url = GoString(url)
	connection.Scheme = schemeString
	connection.Host = hostString
	if enableKeepAlive != 0 {
		connection.KeepAlive = true
	}
	connection.IsSecure = isSecure
	connection.Settings = template.Settings

	logger.Printf("%-132s %s created http connection %s (templateId=%s, url=%s, enableKeepAlive=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpCreateTemplate"),
		color.Yellow.Sprintf("0x%X", connection.Id),
		color.Yellow.Sprintf("0x%X", templateId),
		color.Blue.Sprint(connection.Url),
		color.Green.Sprint(enableKeepAlive),
	)
	return uintptr(connection.Id)
}

// checkScheme validates the scheme and determines if the connection should be secure.
func checkScheme(scheme string) (isSecure bool, err uintptr) {
	if scheme == "" {
		logger.Printf("%-132s %s failed due to empty scheme.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("checkScheme"),
		)
		return false, 0x804311FE
	}
	if strings.EqualFold(scheme, "http") {
		return false, 0
	}
	if strings.EqualFold(scheme, "https") {
		return true, 0
	}

	logger.Printf("%-132s %s failed due to invalid scheme.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("checkScheme"),
	)
	return false, 0x80431061
}
