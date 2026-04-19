//go:build windows

package facade

import "syscall"

func detachAttrsOther() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: 0x00000008, // DETACHED_PROCESS
	}
}
