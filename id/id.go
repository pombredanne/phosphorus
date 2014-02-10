package id

import (
	"fmt"
	"sync"
	"syscall"
	"time"
)

const EPOCH = 1388534400000

type Generator struct {
	machine      int16
	counter      int16
	lastRollover int64
	lock         sync.Mutex
}

func NewGenerator(machine int16) *Generator {
	return &Generator{
		machine: machine & 0x3ff,
		counter: 0}
}

func (g *Generator) Id() (i int64, err error) {
	g.lock.Lock()
	defer g.lock.Unlock()
	now := timeLib()
	i = (now-EPOCH)<<22 | int64(g.machine)<<12 | int64(g.counter)

	if g.counter+1 >= 0x1000 {
		if now <= g.lastRollover {
			err = fmt.Errorf("ID rollover before next ms!")
			return
		}
		g.counter = 0
		g.lastRollover = now
	} else {
		g.counter++
	}

	return
}

func (g *Generator) SafeId() (i int64) {
	for {
		i, err := g.Id()
		if err != nil {
			time.Sleep(time.Millisecond)
			continue
		}
		return i
	}
}

func timeLib() int64 {
	return time.Now().UnixNano() / 1e6
}

// Left here as a warning to others:
// BenchmarkSystime	10000000	       221 ns/op
// BenchmarkLibtime	100000000	        16.4 ns/op
func timeSys(tv *syscall.Timeval) int64 {
	syscall.Gettimeofday(tv)
	return int64(tv.Sec)*1e3 + int64(tv.Usec)/1e3
}
