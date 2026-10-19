//go:build linux

package public

import (
	"github.com/g0rbe/go-chattr"
	"os"
)

func changeFileAttributes(fp *os.File, attr string) error {
	flags := map[byte]int32{'i': chattr.FS_IMMUTABLE_FL, 'a': chattr.FS_APPEND_FL,
		'd': chattr.FS_NODUMP_FL, 's': chattr.FS_SYNC_FL, 'c': chattr.FS_COMPR_FL}
	if attr[0] == '+' {
		return chattr.SetAttr(fp, flags[attr[1]])
	}
	return chattr.UnsetAttr(fp, flags[attr[1]])
}
