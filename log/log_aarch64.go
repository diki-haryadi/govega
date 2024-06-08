//go:build linux && arm64
// +build linux,arm64

package log

import "syscall"

func Dup(from, to int) error {
	return syscall.Dup3(from, to, 0)
}
