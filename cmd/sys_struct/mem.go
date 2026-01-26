package sys_struct

// GrowGoStack grows the current goroutine stack by kb kilobytes, essentially pre-allocating space.
func GrowGoStack(kb int) {
	var dummy [1024]byte
	if kb > 0 {
		GrowGoStack(kb - 1)
	}
	_ = dummy
}
