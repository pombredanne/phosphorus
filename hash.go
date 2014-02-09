package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"willstclair.com/phosphorus/random"
)

var cmdHash = &Command{
	Run:       runHash,
	UsageLine: "hash",
	Short:     "hash",
}

var (
	hashDir  string
	hashSeed string
)

func init() {
	cmdHash.Flag.StringVar(&hashDir, "dir", "", "")
	cmdHash.Flag.StringVar(&hashSeed, "seed", "phosphorus", "")
}

func runHash(cmd *Command, args []string) {
	log.Println("Hello")
	wait := sync.WaitGroup{}

	w := make(chan *_job)
	for i := 0; i < (runtime.NumCPU() + 1); i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for j := range w {
				err := random.Gen(j.wc, rand.NewSource(j.seed))
				if err != nil {
					panic(err)
				}
			}
		}()
	}

	rng := rand.New(seedSource(hashSeed))
	for i := 0; i < (1 << 7); i++ {
		filename := filepath.Join(hashDir, fmt.Sprintf("%02x", i))
		file, err := os.Create(filename)
		if err != nil {
			panic(err)
		}

		wait.Add(1)
		w <- &_job{io.WriteCloser(file), rng.Int63()}
		fmt.Print(".")
	}
	fmt.Println()
	close(w)
	fmt.Println("Waiting to finish.")
	wait.Wait()
}

func seedSource(s string) rand.Source {
	h := fnv.New64a()
	h.Write([]byte(s))
	return rand.NewSource(int64(h.Sum64()))
}

type _job struct {
	wc   io.WriteCloser
	seed int64
}
