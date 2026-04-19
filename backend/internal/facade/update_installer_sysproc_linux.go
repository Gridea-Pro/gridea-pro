//go:build linux

package facade

import "syscall"

func detachAttrsOther() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
