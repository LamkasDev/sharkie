package emu

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/LamkasDev/sharkie/cmd/config"
	"github.com/LamkasDev/sharkie/cmd/elf"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/module"
	"github.com/LamkasDev/sharkie/cmd/linker"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/patcher"
	"github.com/gookit/color"
)

// LoadModule loads & links module specified by name or absolute path.
func (m *ModuleManager) LoadModule(pathOrName string, force bool) (*elf.Elf, error) {
	// Check if module is loaded already.
	if module := m.GetModule(pathOrName); module != nil {
		return module, nil
	}
	logger.Println()

	// Only load the modules.
	loadedModule, err := m._RecursiveLoadModule(pathOrName, force)
	if err != nil {
		return nil, err
	}

	// Link & patch everything now.
	GlobalModuleManager.ModulesLock.RLock()
	defer GlobalModuleManager.ModulesLock.RUnlock()
	var newlyLinkedModules []*elf.Elf
	for _, module := range m.Modules {
		if module == nil || module.Linked {
			continue
		}
		logger.Printf(
			"Linking module %s from %s...\n",
			color.Blue.Sprint(module.Name),
			color.Blue.Sprint(module.Path),
		)
		if err = linker.GlobalLinker.Link(module); err != nil {
			return nil, err
		}
		if err = patcher.GlobalPatcher.Patch(module); err != nil {
			return nil, err
		}
		module.Linked = true
		newlyLinkedModules = append(newlyLinkedModules, module)
		logger.Println()
	}

	// Attempt to resolve any pending relocations now that new modules have been linked.
	linker.ResolvePendingRelocations(m.Modules)

	// Expand TLS for newly linked modules across existing threads.
	for _, module := range newlyLinkedModules {
		ExpandThreadTLS(module)
	}

	return loadedModule, nil
}

// _RecursiveLoadModule loads a module and dependencies without linking.
func (m *ModuleManager) _RecursiveLoadModule(pathOrName string, force bool) (*elf.Elf, error) {
	// Check if module is loaded already.
	if module := m.GetModule(pathOrName); module != nil {
		return module, nil
	}

	// Get module path.
	modulePath := m.GetModulePath(pathOrName)
	if modulePath == nil {
		if strings.Contains(pathOrName, "libSceNpToolkit.") {
			logger.Printf("failed to find libSceNpToolkit...")
			return nil, nil
		}
		return nil, errors.New(fmt.Sprintf("could not find module %s", pathOrName))
	}

	// Read module binary.
	moduleIndex := uint64(len(m.Modules))
	logger.Printf(
		"Loading module %s from %s...\n",
		color.Green.Sprint(moduleIndex),
		color.Blue.Sprint(*modulePath),
	)
	data, err := os.ReadFile(*modulePath)
	if err != nil {
		return nil, err
	}

	// If it's in FSELF format, convert it to ELF.
	if len(data) >= 4 && string(data[0:4]) == "\x4f\x15\x3d\x1d" {
		logger.Printf("converting to .elf using ps4_unfself.py...\n")
		err = config.RunTool("ps4_unfself.py", *modulePath)
		if err != nil {
			return nil, fmt.Errorf("failed to convert FSELF: %v", err)
		}

		elfPath := strings.TrimSuffix(*modulePath, filepath.Ext(*modulePath)) + ".elf"
		data, err = os.ReadFile(elfPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read generated ELF: %v", err)
		}
		*modulePath = elfPath
	}

	// Parse ELF and append to module list.
	module := elf.NewElf(data)
	module.ModuleIndex = moduleIndex
	module.Path = *modulePath
	GlobalModuleManager.ModulesLock.Lock()
	m.Modules = append(m.Modules, module)

	// We strip any extensions from pathOrName and module name; Sometimes module name and file name don't match.
	// If that happens, we add it separately (seems to be case for versioned libraries and game executables).
	baseName := stripExtension(filepath.Base(pathOrName))
	moduleName := stripExtension(module.Name)
	if moduleName == "" {
		moduleName = baseName
	}
	m.ModulesMap[baseName] = module
	if moduleName != baseName {
		m.ModulesMap[moduleName] = module
	}
	GlobalModuleManager.ModulesLock.Unlock()
	logger.Println()

	// Recursively load dependencies.
	for _, needed := range module.DynamicInfo.Needed {
		// These ones not are available in retail.
		neededName := stripExtension(needed)
		if neededName == "libSceGnmDriver_padebug" || neededName == "libSceDbgAddressSanitizer" ||
			neededName == "libSceDipsw" || neededName == "libSceOttvCapture" {
			continue
		}
		// TODO: these ones stubbed temporarily.
		if neededName == "libSceCamera" || neededName == "libSceVrTracker" {
			continue
		}
		// If the game doesn't explicitly load this module and it's not a boot-time module; skip it.
		if !force && !IsBootModule(neededName) {
			continue
		}
		if _, err = m._RecursiveLoadModule(needed, true); err != nil {
			return nil, err
		}
	}

	return module, nil
}

func (m *ModuleManager) IsModuleLoaded(pathOrName string) bool {
	return m.GetModule(pathOrName) != nil
}

func (m *ModuleManager) GetModule(pathOrName string) *elf.Elf {
	GlobalModuleManager.ModulesLock.RLock()
	defer GlobalModuleManager.ModulesLock.RUnlock()
	return m.ModulesMap[stripExtension(filepath.Base(pathOrName))]
}

// RunModuleInitializers recursively executes init functions of modules.
func (m *ModuleManager) RunModuleInitializers(thread *Thread, module *elf.Elf, force, skipOwnInit bool, args ...uintptr) uintptr {
	if module.Initialized {
		return 0
	}
	module.Initialized = true

	for _, needed := range module.DynamicInfo.Needed {
		// These ones not are available in retail.
		neededName := stripExtension(needed)
		if neededName == "libSceGnmDriver_padebug" || neededName == "libSceDbgAddressSanitizer" ||
			neededName == "libSceDipsw" || neededName == "libSceOttvCapture" {
			continue
		}
		// TODO: these ones stubbed temporarily.
		if neededName == "libSceCamera" || neededName == "libSceVrTracker" {
			continue
		}
		// If the game doesn't explicitly load this module and it's not a boot-time module; skip it.
		if !force && !IsBootModule(neededName) {
			continue
		}
		if dependency := m.ModulesMap[neededName]; dependency != nil {
			m.RunModuleInitializers(thread, dependency, true, false)
		}
	}

	if skipOwnInit {
		return 0
	}
	m.CurrentModule = module

	// Call C++ initialization functions.
	var ret uintptr
	for _, funcAddr := range module.DynamicInfo.PreInitArray {
		logger.Printf(
			"Calling %s's %s function at %s (relative=%s)...\n",
			color.Blue.Sprint(module.Name),
			color.Magenta.Sprint("DT_PREINIT_ARRAY"),
			color.Yellow.Sprintf("0x%X", funcAddr),
			color.Yellow.Sprintf("0x%X", uintptr(funcAddr)-module.BaseAddress),
		)
		thread.CallAndWaitFromStub(uintptr(funcAddr), 0)
		logger.Printf(
			"Finished calling %s's %s function at %s...\n\n",
			color.Blue.Sprint(module.Name),
			color.Magenta.Sprint("DT_PREINIT_ARRAY"),
			color.Yellow.Sprintf("0x%X", funcAddr),
		)
	}
	if module.DynamicInfo.InitFunc != nil {
		logger.Printf(
			"Calling %s's %s function at %s (relative=%s)...\n",
			color.Blue.Sprint(module.Name),
			color.Magenta.Sprint("DT_INIT"),
			color.Yellow.Sprintf("0x%X", *module.DynamicInfo.InitFunc),
			color.Yellow.Sprintf("0x%X", uintptr(*module.DynamicInfo.InitFunc)-module.BaseAddress),
		)
		if len(args) > 0 {
			ret = thread.CallAndWaitFromStub(uintptr(*module.DynamicInfo.InitFunc), args...)
		} else {
			ret = thread.CallAndWaitFromStub(uintptr(*module.DynamicInfo.InitFunc), 0)
		}
		logger.Printf(
			"Finished calling %s's %s function at %s... (return=%s)\n\n",
			color.Blue.Sprint(module.Name),
			color.Magenta.Sprint("DT_INIT"),
			color.Yellow.Sprintf("0x%X", *module.DynamicInfo.InitFunc),
			color.Yellow.Sprintf("0x%X", ret),
		)
	}
	for _, funcAddr := range module.DynamicInfo.InitArray {
		logger.Printf(
			"Calling %s's %s function at %s (relative=%s)...\n",
			color.Blue.Sprint(module.Name),
			color.Magenta.Sprint("DT_INIT_ARRAY"),
			color.Yellow.Sprintf("0x%X", funcAddr),
			color.Yellow.Sprintf("0x%X", uintptr(funcAddr)-module.BaseAddress),
		)
		thread.CallAndWaitFromStub(uintptr(funcAddr), 0)
		logger.Printf(
			"Finished calling %s's %s function at %s...\n\n",
			color.Blue.Sprint(module.Name),
			color.Magenta.Sprint("DT_INIT_ARRAY"),
			color.Yellow.Sprintf("0x%X", funcAddr),
		)
	}

	return ret
}

// RunModule runs module specified by name.
func (m *ModuleManager) RunModule(pathOrName string) {
	baseName := stripExtension(filepath.Base(pathOrName))
	module := m.ModulesMap[baseName]
	if module == nil {
		log.Panicf("Module %s is not loaded!\n", pathOrName)
	}

	logger.Printf(
		"Running module %s...\n",
		color.Blue.Sprint(pathOrName),
	)
	m.MainThread = CreateThread("MainThread", StackDefaultSize)
	m.MainThread.Setup()

	m.RunModuleInitializers(m.MainThread, module, false, true)

	m.CurrentModule = module
	m.MainThread.Run(m.CurrentModule)
}
