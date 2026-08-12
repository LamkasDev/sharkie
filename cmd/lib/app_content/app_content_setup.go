package app_content

import (
	"os"
	"path/filepath"

	"github.com/LamkasDev/sharkie/cmd/config"
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/app_content"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/psf"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/system_service"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000001610
// __int64 __fastcall sceAppContentInitialize(__int64, __int64)
func libSceAppContent_sceAppContentInitialize(initParamPtr, bootParamPtr uintptr) uintptr {
	if GlobalAppContentInstance.IsInitialized {
		logger.Printf("%-132s %s failed (already initialized).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceAppContentInitialize"),
		)
		return 0x809E0003
	}
	GlobalAppContentInstance.IsInitialized = true

	// Iterate addons and load their metadata.
	addonsDir := config.GetGameAddonsDir()
	if _, err := os.Stat(addonsDir); err == nil {
		entries, _ := os.ReadDir(addonsDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			dlcSfoPath := filepath.Join(addonsDir, entry.Name(), "sce_sys", "param.sfo")
			dlcPsf, err := psf.NewPsfFromPath(dlcSfoPath)
			if err != nil {
				logger.Printf("Additional content folder %s has no param.sfo.\n", entry.Name())
				continue
			}

			category := dlcPsf.MapStrings["CATEGORY"]
			if len(category) >= 2 && category[:2] == "ac" {
				contentId := dlcPsf.MapStrings["CONTENT_ID"]
				if contentId == "" {
					logger.Printf("Additional content %s param.sfo is missing CONTENT_ID.\n", entry.Name())
					continue
				}
				if len(contentId) <= AppContentEntitlementLabelOffset {
					logger.Printf("Additional content %s param.sfo has malformed CONTENT_ID.\n", entry.Name())
					continue
				}

				entitlementId := contentId[AppContentEntitlementLabelOffset:]
				info := AppContentAddcontInfo{
					Status: AppContentAddcontDownloadStatusInstalled,
				}
				copy(info.EntitlementLabel.Data[:], entitlementId)

				GlobalAppContentInstance.AddcontInfo = append(GlobalAppContentInstance.AddcontInfo, info)
			} else {
				logger.Printf("Additional content folder %s is not additional content.\n", entry.Name())
			}
		}
	}

	// Send event.
	if len(GlobalAppContentInstance.AddcontInfo) > 0 {
		event := system_service.SystemServiceEvent{
			Type: system_service.SystemServiceEventTypeEntitlementUpdate,
		}
		system_service.GlobalSystemService.EventQueue.Add(event)
	}

	logger.Printf("%-132s %s initialized app content.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceAppContentInitialize"),
	)
	return 0
}

// 0x0000000000001630
// __int64 __fastcall sceAppContentAppParamGetInt(unsigned int, __int64)
func libSceAppContent_sceAppContentAppParamGetInt(paramId AppContentAppParamId, outValuePtr uintptr) uintptr {
	if outValuePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceAppContentAppParamGetInt"),
		)
		return 0x809E0000
	}
	if GlobalAppContentInstance.ParamSfo == nil {
		logger.Printf("%-132s %s failed due to missing param.sfo.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceAppContentAppParamGetInt"),
		)
		return 0x809E0000
	}

	// Read parameter.
	var value int32
	var ok bool
	switch paramId {
	case AppContentAppParamIdSkuFlag:
		value = AppContentAppParamSkuFlagFull
		ok = true
	case AppContentAppParamIdUserDefinedParam1:
		value, ok = GlobalAppContentInstance.ParamSfo.MapIntegers["USER_DEFINED_PARAM_1"]
	case AppContentAppParamIdUserDefinedParam2:
		value, ok = GlobalAppContentInstance.ParamSfo.MapIntegers["USER_DEFINED_PARAM_2"]
	case AppContentAppParamIdUserDefinedParam3:
		value, ok = GlobalAppContentInstance.ParamSfo.MapIntegers["USER_DEFINED_PARAM_3"]
	case AppContentAppParamIdUserDefinedParam4:
		value, ok = GlobalAppContentInstance.ParamSfo.MapIntegers["USER_DEFINED_PARAM_4"]
	default:
		logger.Printf("%-132s %s failed due to invalid parameter id %d.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceAppContentAppParamGetInt"),
			paramId,
		)
		return 0x809E0000
	}
	if !ok {
		value = 0
	}
	WriteResult(outValuePtr, uint32(value))

	logger.Printf("%-132s %s read parameter %d as %d.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceAppContentAppParamGetInt"),
		paramId, value,
	)
	return 0
}
