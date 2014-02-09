package random

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

func Compact(x float64) uint16 {
	return uint16(math.Floor((x + 8.0) * 4096.0))
}

func Uncompact(x uint16) float64 {
	return (float64(x) / 4096.0) - 8.0
}

func Map(filename string) *[1 << 26]uint16 {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	mmap, err := syscall.Mmap(
		int(file.Fd()),
		0,
		1<<27,
		syscall.PROT_READ,
		syscall.MAP_SHARED)
	if err != nil {
		panic(err)
	}

	return (*[1 << 26]uint16)(unsafe.Pointer(&mmap[0]))
}

func Gen(w io.WriteCloser, src rand.Source) error {
	rng := rand.New(src)
	for i := uint64(0); i < 1<<26; i++ {
		err := binary.Write(w, binary.LittleEndian, Compact(rng.NormFloat64()))
		if err != nil {
			return err
		}
	}

	return w.Close()
}

type RandomStore struct {
	files [1 << 7]*[1 << 26]uint16
}

func NewRandomStore(dir string) *RandomStore {
	r := &RandomStore{}
	for i := 0; i < (1 << 7); i++ {
		filename := filepath.Join(dir, fmt.Sprintf("%02x", i))
		r.files[i] = Map(filename)
	}
	return r
}

func (r *RandomStore) Get(i int64) float64 {
	return Uncompact(r.files[i&0x7f][i>>7])
}
