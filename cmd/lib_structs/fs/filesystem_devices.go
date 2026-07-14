package fs

// InitializeSystemFiles maps the virtual devices and daemon files expected by the PS4 kernel.
func (shFs *SharkieFilesystem) InitializeSystemFiles() error {
	// Device files.
	if err := shFs.MkdirAll(GetUsablePath("/dev")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/dipsw")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/hmd_cmd")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/hmd_snsr")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/hmd_3da")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/hmd_dist")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/sbl_srv")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/hid")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/ajm")); err != nil {
		return err
	}
	if _, err := shFs.Create(GetUsablePath("/dev/dce")); err != nil {
		return err
	}

	// Daemon files.
	if _, err := shFs.Write(GetUsablePath("SceNpTpip"), make([]byte, 4096)); err != nil {
		panic(err)
	}

	return nil
}
