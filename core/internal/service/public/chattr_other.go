//go:build !linux

package public

import (
	"fmt"
	"os"
)

func changeFileAttributes(_ *os.File, _ string) error {
	return fmt.Errorf("Linux file attributes are unavailable on this platform")
}
