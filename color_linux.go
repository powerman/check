// +build linux

package check

import "golang.org/x/sys/unix"

const ioctlReadTermios = unix.TCGETS
