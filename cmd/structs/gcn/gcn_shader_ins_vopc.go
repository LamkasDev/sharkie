package gcn

func VopcMap() map[uint32]string {
	m := make(map[uint32]string)
	add16 := func(base uint32, prefix, suffix string) {
		ops := []string{
			"f", "lt", "eq", "le", "gt", "lg", "ge", "o",
			"u", "nge", "nlg", "ngt", "nle", "neq", "nlt", "tru",
		}
		for i, op := range ops {
			opcode := base + uint32(i)
			m[opcode] = "v_cmp" + prefix + "_" + op + suffix
		}
	}
	add8 := func(base uint32, prefix, suffix string) {
		ops := []string{"f", "lt", "eq", "le", "gt", "lg", "ge", "tru"}
		for i, op := range ops {
			opcode := base + uint32(i)
			m[opcode] = "v_cmp" + prefix + "_" + op + suffix
		}
	}

	// VOPC Instructions with 16 Compare Operations.
	add16(0x00, "", "_f32")  // V_CMP_{OP16}_F32
	add16(0x10, "x", "_f32") // V_CMPX_{OP16}_F32
	add16(0x20, "", "_f64")  // V_CMP_{OP16}_F64
	add16(0x30, "x", "_f64") // V_CMPX_{OP16}_F64

	add16(0x40, "s", "_f32")  // V_CMPS_{OP16}_F32
	add16(0x50, "sx", "_f32") // V_CMPSX_{OP16}_F32
	add16(0x60, "s", "_f64")  // V_CMPS_{OP16}_F64
	add16(0x70, "sx", "_f64") // V_CMPSX_{OP16}_F64

	// VOPC Instructions with Eight Compare Operations.
	add8(0x80, "", "_i32")  // V_CMP_{OP8}_I32
	add8(0x90, "x", "_i32") // V_CMPX_{OP8}_I32
	add8(0xA0, "", "_i64")  // V_CMP_{OP8}_I64
	add8(0xB0, "x", "_i64") // V_CMPX_{OP8}_I64

	add8(0xC0, "", "_u32")  // V_CMP_{OP8}_U32
	add8(0xD0, "x", "_u32") // V_CMPX_{OP8}_U32
	add8(0xE0, "", "_u64")  // V_CMP_{OP8}_U64
	add8(0xF0, "x", "_u64") // V_CMPX_{OP8}_U64

	m[0x88] = "v_cmp_class_f32"
	m[0x98] = "v_cmpx_class_f32"
	m[0xA8] = "v_cmp_class_f64"
	m[0xB8] = "v_cmpx_class_f64"

	return m
}

func (instr *Instruction) DecodeVOPC() {
	dw := instr.Dwords[0]
	instr.Details = &VectorDetails{
		Src0: dw & 0b1111_1111_1,
		Src1: (dw >> 9) & 0b1111_1111,
		Op:   (dw >> 17) & 0b1111_1111,
	}
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}
