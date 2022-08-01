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
	safeBitSet [2097152]uint64 // 2^27 is the max number of user's pages in a process , all pages status cost 16MB memory.
	pageMap    *os.File
}

func (r *ReaderImplRPMWithAnti) initImpl() *Reader {
	/*
		As a temporary solution, inotify will disable by setting 'max_user_watches' to 0.
		Notice: You should start app after setting inotify,because no progress can register new inotify watcher.
		'max_user_watches' will not restore automatically.
		You can manually set 'max_user_watches' to a normal value if you want to use inotify after using MemInspector.
		Restarting the device is also effective.
	*/
	stat, _ := CheckInotify()
	if stat {
		fmt.Printf("*** Inotify is enabled,you should disable it by \"MemInspector inotify off\" first! ***\n")
		os.Exit(1)
	}
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

// TODO: caching page status and make anti-function standalone
/* Anti-Read-Detection built in cache by default, it costs 16MB memory. if you want to disable it for some reason,
should make a new Reader like ReaderImplRPMWithAntiNoCache and remove cache code for performance. */
func (r *ReaderImplRPMWithAnti) Read(addr, size uint) ([]byte, error) {
	pages := (addr&4095 + size) >> 12
	for i := uint(0); i < pages+1; i += 1 { //check every page status
		pageIndex := addr>>12 + i                                      // page is 12 bits
		cacheIndex := pageIndex >> 6                                   // pageIndex / 64
		if ((r.safeBitSet[cacheIndex] >> (pageIndex & 63)) & 1) == 1 { // pageIndex % 64
			continue //cache hit
		}
		status := [8]byte{}
		offset := int64(pageIndex << 3)
		if _, err := r.pageMap.Seek(offset, os.SEEK_SET); err != nil {
			return nil, fmt.Errorf("failed to read page atatus at 0x%x (0x%x): %v", offset, addr, err)
		}
		//use preadv may bypass inotify,but it's not a good idea.
		n, err := r.pageMap.Read(status[:]) //disable inotify first,some app will listen to inotify event to monitor reading.
		if n != 8 {
			return nil, fmt.Errorf("wrong page data at 0x%x (0x%x): %v", offset, addr, err)
		}
		memBits := binary.LittleEndian.Uint64(status[:])
		if memBits>>62 == 0 {
			return nil, fmt.Errorf("page at 0x%x (0x%x) is not present", offset, addr)
		}
		r.safeBitSet[cacheIndex] |= 1 << (pageIndex & 63)
	}
	data, err := r.ReaderImplRPM.Read(addr, size)
	return data, err
}
