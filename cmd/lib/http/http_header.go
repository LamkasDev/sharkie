package http

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/http"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000012680
// __int64 __fastcall sceHttpAddRequestHeader(unsigned int, __int64, __int64, unsigned int)
func libSceHttp_sceHttpAddRequestHeader(id uint32, name, value Cstring, mode HttpHeaderMode) uintptr {
	if name == nil || value == nil {
		logger.Printf("%-132s %s failed due to invalid name or value pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpAddRequestHeader"),
		)
		return 0x804311FE
	}
	if mode < HttpHeaderModeOverwrite || mode > HttpHeaderModeAdd {
		logger.Printf("%-132s %s failed due to invalid mode.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpAddRequestHeader"),
		)
		return 0x804311FE
	}

	rawObject := GlobalHttpHandler.GetHeaderObject(id)
	if rawObject == nil {
		logger.Printf("%-132s %s failed due to invalid object.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpAddRequestHeader"),
		)
		return 0x80431100
	}
	nameString := GoString(name)
	valueString := GoString(value)
	switch object := rawObject.(type) {
	case *HttpTemplate:
		object.Headers[nameString] = valueString
	case *HttpConnection:
		object.Headers[nameString] = valueString
	case *HttpRequest:
		object.Headers[nameString] = valueString
	}

	logger.Printf("%-132s %s added request header to object %s (name=%s, value=%s, mode=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpCreateTemplate"),
		color.Yellow.Sprintf("0x%X", id),
		color.Blue.Sprint(nameString),
		color.Blue.Sprint(valueString),
		color.Green.Sprint(mode),
	)
	return 0
}
