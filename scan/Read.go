package scan

import (
	"errors"
	"sync"
	"syscall"
	"unsafe"
)

const (
	READER_IMPL_RPM = iota
)

type ReaderImpl interface {
	Read(addr, size uint) ([]byte, error)
	ReadOneByte(addr uint) (byte, error) //code analysis give a warning for name "ReadByte" , use "ReadOneByte" instead.
	ReadWord(addr uint) (int, error)
	ReadDword(addr uint) (int, error)
	ReadQword(addr uint) (int, error)
	initImpl() *Reader
}

func NewReader(pid, impl uint) *Reader {
	switch impl {
	case READER_IMPL_RPM:
		return (&ReaderImplRPM{Reader: Reader{pid: pid}}).initImpl()
	default:
		panic("Unknown reader implementation")
	}
}

type Reader struct {
	pid uint
	ReaderImpl
}

type ReaderImplRPM struct {
	Reader
	iovPool sync.Pool // pooling of Iovec can reduce the memory allocation and GC cost
}

func (r *ReaderImplRPM) initImpl() *Reader {
	r.iovPool = sync.Pool{
		New: func() interface{} {
			return new(syscall.Iovec)
		},
	}
	r.ReaderImpl = r
	return &r.Reader
}

func (r *ReaderImplRPM) Read(addr, size uint) ([]byte, error) {
	data := make([]byte, size)
	local := r.iovPool.Get().(*syscall.Iovec)
	local.Base = &data[0]
	local.Len = uint64(size)
	remote := r.iovPool.Get().(*syscall.Iovec)
	remote.Base = (*byte)(unsafe.Pointer(uintptr(addr))) //I'm sure that's right,may some other way can clear go vet warning
	remote.Len = uint64(size)
	niladdr := &struct{}{}
	_, _, e1 := syscall.Syscall6(syscall.SYS_PROCESS_VM_READV,
		uintptr(r.pid),
		uintptr(unsafe.Pointer(&local)),
		1,
		uintptr(unsafe.Pointer(&remote)),
		1,
		0)
	if e1 != 0 {
		return nil, errors.New("syscall error ,ret " + e1.Error())
	}
	r.iovPool.Put(local)
	r.iovPool.Put(remote)
	return data, nil
}
