package gpu

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
	"github.com/gookit/color"
)

// DumpShaderOnce scans the GCN shader at addr and prints its bytecode to the log.
func (l *Liverpool) DumpShaderOnce(address uintptr, stage string, rsrc1, rsrc2 uint32) error {
	if address == 0 {
		return fmt.Errorf("invalid address")
	}
	if _, loaded := l.SeenShaders.LoadOrStore(address, struct{}{}); loaded {
		return nil
	}

	// Decode inputs.
	vgprs, sgprs := decodeRsrc1(rsrc1)
	scratchEnable := rsrc2 & 1
	userDataCount := (rsrc2 >> 1) & 0x1F

	// Create a new shader from shader byte-code.
	shader, err := gcn.NewGcnShader(address)
	if err != nil {
		panic(err)
	}
	logger.Printf("[%s] Scanned %s shader %s of %s bytes (vgprs=%s, sgprs=%s, scratchEnable=%s, userDataCount=%s, rsrc1=%s, rsrc2=%s)...\n",
		color.Blue.Sprint("SHADER"),
		color.Blue.Sprint(stage),
		color.Yellow.Sprintf("0x%X", address),
		color.Green.Sprint(shader.DwordLength*4),
		color.Yellow.Sprintf("0x%X", vgprs),
		color.Yellow.Sprintf("0x%X", sgprs),
		color.Yellow.Sprintf("0x%X", scratchEnable),
		color.Yellow.Sprintf("0x%X", userDataCount),
		color.Yellow.Sprintf("0x%X", rsrc1),
		color.Yellow.Sprintf("0x%X", rsrc2),
	)

	// Print the disassembly.
	var sb strings.Builder
	for _, block := range shader.Cfg.Blocks {
		fmt.Fprintf(&sb, "[%s] Block %s (",
			color.Blue.Sprint("SHADER"),
			color.Green.Sprint(block.Id),
		)
		if block.IsLoopHeader {
			fmt.Fprintf(&sb, "loop continue=%s ",
				color.Green.Sprint(block.ContinueBlockId),
			)
		}
		if block.MergeBlockId >= 0 {
			fmt.Fprintf(&sb, "merge=%s",
				color.Green.Sprint(block.MergeBlockId),
			)
		} else {
			fmt.Fprintf(&sb, "no merge")
		}
		fmt.Fprint(&sb, "):\n")

		for _, instr := range block.Instructions {
			fmt.Fprintf(&sb, "[%s] %s: %s\n",
				color.Blue.Sprint("SHADER"),
				color.Yellow.Sprintf("0x%04X", instr.DwordOffset),
				color.Cyan.Sprint(instr.String()),
			)
		}

		fmt.Fprintf(&sb, "[%s] Branches (%s)",
			color.Blue.Sprint("SHADER"),
			color.Blue.Sprint(block.Term),
		)
		if block.BranchCond != gcn.CondNone {
			fmt.Fprintf(&sb, " on %s",
				color.Magenta.Sprint(block.BranchCond),
			)
		}
		switch len(block.Successors) {
		case 0:
			fmt.Fprintf(&sb, " to %s.\n", color.Red.Sprint("Exit"))
		case 1:
			fmt.Fprintf(&sb, " to %s.\n", color.Green.Sprint(block.Successors[0]))
		case 2:
			fmt.Fprintf(&sb, " to %s (fallthrough=%s).\n",
				color.Green.Sprint(block.Successors[1]),
				color.Green.Sprint(block.Successors[0]),
			)
		}
	}
	logger.Print(sb.String())

	// Dump the disassembled shader.
	textFilename := path.Join("temp", "shaders", fmt.Sprintf("shader_0x%X_%s.txt", address, stage))
	if err = os.WriteFile(textFilename, []byte(sb.String()), 0777); err != nil {
		return err
	}
	logger.Printf("[%s] Dumped shader to %s...\n",
		color.Blue.Sprint("SHADER"),
		color.Blue.Sprint(textFilename),
	)

	return nil
}

// decodeRsrc1 extracts VGPR and SGPR counts from a SPI_SHADER_PGM_RSRC1 or COMPUTE_PGM_RSRC1 value.
func decodeRsrc1(rsrc1 uint32) (vgprs, sgprs uint32) {
	vgprs = ((rsrc1 & 0x3F) + 1) * 4
	sgprs = (((rsrc1 >> 6) & 0xF) + 1) * 8

	return
}
