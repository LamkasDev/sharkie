package emu

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/linker"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/patcher"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// LoadModule loads & links module specified by name.
func (m *ModuleManager) LoadModule(name string) error {
	logger.Println()

	// Only load the modules.
	if err := m._RecursiveLoadModule(name); err != nil {
		return err
	}

	// Link & patch everything now.
	GlobalModuleManager.ModulesLock.RLock()
	defer GlobalModuleManager.ModulesLock.RUnlock()
	for _, module := range m.Modules {
		if module == nil || module.Linked {
			continue
		}
		logger.Printf(
			"Linking module %s from %s...\n",
			color.Blue.Sprint(module.Name),
			color.Blue.Sprint(module.Path),
		)
		if err := linker.GlobalLinker.Link(module); err != nil {
			return err
		}
		if err := patcher.GlobalPatcher.Patch(module); err != nil {
			return err
		}
		module.Linked = true
		logger.Println()
	}

	return nil
}

// _RecursiveLoadModule loads a module and dependencies without linking.
func (m *ModuleManager) _RecursiveLoadModule(name string) error {
	GlobalModuleManager.ModulesLock.RLock()
	if m.ModulesMap[name] != nil {
		GlobalModuleManager.ModulesLock.RUnlock()
		return nil
	}
	GlobalModuleManager.ModulesLock.RUnlock()

	modulePath := m.GetModulePath(name)
	if modulePath == nil {
		return errors.New(fmt.Sprintf("could not find module %s", name))
	}

	moduleIndex := uint64(len(m.Modules))
	logger.Printf(
		"Loading module %s from %s...\n",
		color.Green.Sprint(moduleIndex),
		color.Blue.Sprint(*modulePath),
	)
	data, err := os.ReadFile(*modulePath)
	if err != nil {
		return err
	}

	module := elf.NewElf(data)
	module.ModuleIndex = moduleIndex
	module.Path = *modulePath
	GlobalModuleManager.ModulesLock.Lock()
	m.Modules = append(m.Modules, module)
	m.ModulesMap[name] = module
	GlobalModuleManager.ModulesLock.Unlock()
	if module.Name == "libSceFios2.sprx" {
		RegisterFiosStubs()
	}
	logger.Println()

	for _, needed := range module.DynamicInfo.Needed {
		needed = strings.ReplaceAll(needed, ".prx", ".sprx")
		if needed == "libSceGnmDriver_padebug.sprx" ||
			needed == "libSceDbgAddressSanitizer.sprx" ||
			needed == "libSceDipsw.sprx" {
			continue
		}
		if err = m._RecursiveLoadModule(needed); err != nil {
			return err
		}
	}

	return nil
}

// RunModuleInitializers recursively executes init functions of modules.
func (m *ModuleManager) RunModuleInitializers(module *elf.Elf, visited map[string]bool, skipOwnInit bool) {
	if visited[module.Name] {
		return
	}
	visited[module.Name] = true

	for _, needed := range module.DynamicInfo.Needed {
		needed = strings.ReplaceAll(needed, ".prx", ".sprx")
		if needed == "libSceGnmDriver_padebug.sprx" ||
			needed == "libSceDbgAddressSanitizer.sprx" ||
			needed == "libSceDipsw.sprx" {
			continue
		}
		if dependency := m.ModulesMap[needed]; dependency != nil {
			m.RunModuleInitializers(dependency, visited, false)
		}
	}

	isSelfContained := module.Name == "libSceLibcInternal.sprx"
	if skipOwnInit {
		return
	}

	// Call initialization functions.
	if !isSelfContained {
		for _, funcAddr := range module.DynamicInfo.PreInitArray {
			logger.Printf(
				"Calling %s's %s function at %s (relative=%s)...\n",
				color.Blue.Sprint(module.Name),
				color.Magenta.Sprint("DT_PREINIT_ARRAY"),
				color.Yellow.Sprintf("0x%X", funcAddr),
				color.Yellow.Sprintf("0x%X", uintptr(funcAddr)-module.BaseAddress),
			)
			m.MainThread.CallAndWait(uintptr(funcAddr), 0)
		}
	}
	if module.DynamicInfo.InitFunc != nil {
		logger.Printf(
			"Calling %s's %s function at %s (relative=%s)...\n",
			color.Blue.Sprint(module.Name),
			color.Magenta.Sprint("DT_INIT"),
			color.Yellow.Sprintf("0x%X", *module.DynamicInfo.InitFunc),
			color.Yellow.Sprintf("0x%X", uintptr(*module.DynamicInfo.InitFunc)-module.BaseAddress),
		)
		m.MainThread.CallAndWait(uintptr(*module.DynamicInfo.InitFunc), 0)
	}
	if !isSelfContained {
		for _, funcAddr := range module.DynamicInfo.InitArray {
			logger.Printf(
				"Calling %s's %s function at %s (relative=%s)...\n",
				color.Blue.Sprint(module.Name),
				color.Magenta.Sprint("DT_INIT_ARRAY"),
				color.Yellow.Sprintf("0x%X", funcAddr),
				color.Yellow.Sprintf("0x%X", uintptr(funcAddr)-module.BaseAddress),
			)
			m.MainThread.CallAndWait(uintptr(funcAddr), 0)
		}
	}
}

// RunModule runs module specified by name.
func (m *ModuleManager) RunModule(name string) {
	m.CurrentModule = m.ModulesMap[name]
	if m.CurrentModule == nil {
		log.Panicf("Module %s is not loaded!\n", name)
	}

	logger.Printf(
		"Running module %s...\n",
		color.Blue.Sprint(name),
	)
	m.MainThread = CreateThread(0, structs.StackDefaultSize)
	m.MainThread.Setup()

	visited := make(map[string]bool)
	m.RunModuleInitializers(m.CurrentModule, visited, true)
	m.MainThread.Run(m.CurrentModule)
}
