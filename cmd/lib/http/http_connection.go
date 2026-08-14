package http

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/http"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000114F0
// __int64 __fastcall sceHttpCreateConnection(unsigned int, __int64, __int64, unsigned __int16, unsigned int)
func libSceHttp_sceHttpCreateConnection(templateId uint32, serverName, scheme Cstring, port uint16, enableKeepAlive uint32) uintptr {
	template := GlobalHttpHandler.GetTemplate(templateId)
	if template == nil {
		logger.Printf("%-132s %s failed due to invalid template id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpCreateConnection"),
		)
		return 0x80431100
	}
	if serverName == nil {
		logger.Printf("%-132s %s failed due to invalid url.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpCreateConnection"),
		)
		return 0x804311FE
	}

	// Validate scheme.
	schemeString := GoString(scheme)
	isSecure, err := checkScheme(schemeString)
	if err != 0 {
		return err
	}
	hostString := GoString(serverName)

	// Create connection.
	connection := GlobalHttpHandler.CreateConnection()
	connection.TemplateId = templateId
	connection.Url = fmt.Sprintf("%s://%s:%d", schemeString, hostString, port)
	connection.Scheme = schemeString
	connection.Host = hostString
	if enableKeepAlive != 0 {
		connection.KeepAlive = true
	}
	connection.IsSecure = isSecure
	connection.Settings = template.Settings

	logger.Printf("%-132s %s created http connection %s (templateId=%s, url=%s, enableKeepAlive=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpCreateConnection"),
		color.Yellow.Sprintf("0x%X", connection.Id),
		color.Yellow.Sprintf("0x%X", templateId),
		color.Blue.Sprint(connection.Url),
		color.Green.Sprint(enableKeepAlive),
	)
	return uintptr(connection.Id)
}

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
	urlString := GoString(url)

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
	connection.Url = urlString
	connection.Scheme = schemeString
	connection.Host = hostString
	if enableKeepAlive != 0 {
		connection.KeepAlive = true
	}
	connection.IsSecure = isSecure
	connection.Settings = template.Settings

	logger.Printf("%-132s %s created http connection %s (templateId=%s, url=%s, enableKeepAlive=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpCreateConnectionWithURL"),
		color.Yellow.Sprintf("0x%X", connection.Id),
		color.Yellow.Sprintf("0x%X", templateId),
		color.Blue.Sprint(connection.Url),
		color.Green.Sprint(enableKeepAlive),
	)
	return uintptr(connection.Id)
}

// 0x0000000000011620
// __int64 __fastcall sceHttpDeleteConnection(unsigned int)
func libSceHttp_sceHttpDeleteConnection(connectionId uint32) uintptr {
	connection := GlobalHttpHandler.GetConnection(connectionId)
	if connection == nil {
		logger.Printf("%-132s %s failed due to invalid connection id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpDeleteConnection"),
		)
		return 0x80431100
	}

	// Delete connection.
	GlobalHttpHandler.DeleteConnection(connectionId)

	logger.Printf("%-132s %s deleted http connection %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpDeleteConnection"),
		color.Yellow.Sprintf("0x%X", connection.Id),
	)
	return 0
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
