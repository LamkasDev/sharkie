package http

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/http"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000010D80
// __int64 __fastcall sceHttpInit(unsigned int, unsigned int, __int64)
func libSceHttp_sceHttpInit(netMemId, sslCtxId uint32, poolSize uint64) uintptr {
	if poolSize == 0 {
		logger.Printf("%-132s %s failed due to invalid pool size.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpInit"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	context := GlobalHttpHandler.CreateContext()

	logger.Printf("%-132s %s created http context %s (poolSize=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpInit"),
		color.Yellow.Sprintf("0x%X", context.Id),
		color.Yellow.Sprintf("0x%X", poolSize),
	)
	return uintptr(context.Id)
}

// 0x00000000000111B0
// __int64 __fastcall sceHttpCreateTemplate(unsigned int, __int64, unsigned int, unsigned int)
func libSceHttp_sceHttpCreateTemplate(contextId uint32, userAgent Cstring, httpVersion, isAutoProxyConf uint32) uintptr {
	context := GlobalHttpHandler.GetContext(contextId)
	if context == nil {
		logger.Printf("%-132s %s failed due to invalid context id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceHttpCreateTemplate"),
		)
		return 0x80431100
	}

	// Create template.
	template := GlobalHttpHandler.CreateTemplate()
	template.ContextId = contextId
	if userAgent != nil {
		template.UserAgent = GoString(userAgent)
	}
	template.HttpVersion = httpVersion
	if isAutoProxyConf != 0 {
		template.AutoProxyConf = true
	}

	logger.Printf("%-132s %s created http template %s (contextId=%s, userAgent=%s, httpVersion=%s, isAutoProxyConf=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceHttpCreateTemplate"),
		color.Yellow.Sprintf("0x%X", template.Id),
		color.Yellow.Sprintf("0x%X", contextId),
		color.Blue.Sprint(template.UserAgent),
		color.Green.Sprint(httpVersion),
		color.Green.Sprint(isAutoProxyConf),
	)
	return uintptr(template.Id)
}
