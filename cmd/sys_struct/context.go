package sys_struct

import (
	"fmt"

	"github.com/gookit/color"
)

// SprintRegister prints the given register and it's value.
func SprintRegister(register string, value uint64) string {
	return fmt.Sprintf(
		"  %42s = %s (%s)\n",
		color.Blue.Sprint(register),
		color.Yellow.Sprintf("0x%016X", value),
		color.Yellow.Sprintf("%d", value),
	)
}
