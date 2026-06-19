package core

import "syscall"

type FDcomm struct {
	Fd int
}

func (f FDcomm) Write(b []byte) (int, error) {
	return syscall.Write(f.Fd, b)
}

func (f FDcomm) Read(b []byte) (int, error) {
	return syscall.Read(f.Fd, b)
}