// +build darwin dragonfly freebsd netbsd openbsd

package check

import "golang.org/x/sys/unix"

const ioctlReadTermios = unix.TIOCGETA
