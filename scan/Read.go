package scan

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"syscall"
	"unsafe"
)

// TODO: may make each reader in a single file
const (
	READER_IMPL_RPM = iota
	READER_IMPL_RPM_WITH_ANTI
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
	case READER_IMPL_RPM_WITH_ANTI:
		return (&ReaderImplRPMWithAnti{ReaderImplRPM: ReaderImplRPM{Reader: Reader{pid: pid}}}).initImpl()
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
	_, _, e1 := syscall.Syscall6(syscall.SYS_PROCESS_VM_READV,
		uintptr(r.pid),
		uintptr(unsafe.Pointer(local)),
		1,
		uintptr(unsafe.Pointer(remote)),
		1,
		0)
	if e1 != 0 {
		return nil, errors.New("syscall error:" + e1.Error())
	}
	r.iovPool.Put(local)
	r.iovPool.Put(remote)
	return data, nil
}

type ReaderImplRPMWithAnti struct {
	ReaderImplRPM
	safeBitSet [262144]uint64 // 2^24 is the max number of user's pages in a process , all pages status cost 2MB memory.
	pageMap    *os.File
}

func (r *ReaderImplRPMWithAnti) initImpl() *Reader {
	r.ReaderImplRPM.initImpl()
	page, err := os.OpenFile("/proc/"+strconv.Itoa(int(r.Reader.pid))+"/pagemap", os.O_RDONLY|os.O_SYNC, 0)
	if err != nil {
		println(fmt.Errorf("open pagemap file failed:%s", err.Error()))
		os.Exit(-1)
	}
	r.pageMap = page
	r.ReaderImpl = r
	return &r.Reader
}

// TODO: caching page status and make anti standalone
func (r *ReaderImplRPMWithAnti) Read(addr, size uint) ([]byte, error) {
	status := [8]byte{}
	offset := int64(addr / 4096 * 8)
	if _, err := r.pageMap.Seek(offset, os.SEEK_SET); err != nil {
		return nil, fmt.Errorf("failed to read page atatus at 0x%x (0x%x): %v", offset, addr, err)
	}
	n, err := r.pageMap.Read(status[:])
	if n != 8 {
		return nil, fmt.Errorf("wrong page data at 0x%x (0x%x): %v", offset, addr, err)
	}
	memBits := binary.LittleEndian.Uint64(status[:])
	if memBits>>62 == 0 {
		return nil, fmt.Errorf("page at 0x%x (0x%x) is not present", offset, addr)
	}
	data, err := r.ReaderImplRPM.Read(addr, size)
	return data, err
}
