//go:build !windows

package platform_gateway

import "os/exec"

type processGroup struct{}

func newProcessGroup() (*processGroup, error) { return &processGroup{}, nil }
func (*processGroup) attach(*exec.Cmd) error  { return nil }
func (*processGroup) close()                  {}
func hideProcess(*exec.Cmd)                   {}
