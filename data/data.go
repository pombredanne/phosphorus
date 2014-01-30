package data

import (
	// "fmt"
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Record struct {
	RecordId uint32
	Fields   []string
}

type Data struct {
	working    string
	concurrent int
	sem        chan int
	lock       sync.Mutex
}

func NewData(working string, concurrent int) *Data {
	var d Data
	d.working = working
	d.concurrent = concurrent
	d.sem = make(chan int, concurrent)
	for i := 0; i < concurrent; i++ {
		d.sem <- 1
	}

	return &d
}

func fileReader(filename string, records chan *Record) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	r := csv.NewReader(file)
	for line, err := r.Read(); err != io.EOF; line, err = r.Read() {
		if err != nil {
			log.Println(err)
			continue
		}

		rId, err := strconv.ParseUint(line[0], 10, 32)
		if err != nil {
			panic(err)
			// log.Println(err)
			// continue
		}

		records <- &Record{uint32(rId), line[1:]}
	}
	return nil
}

type Slurper func(chan *Record)

func (d *Data) Slurp(fn Slurper) error {
	os.Chdir(d.working)
	d.lock.Lock()
	defer d.lock.Unlock()

	records := make(chan *Record)
	go fn(records)

	var wait sync.WaitGroup
	files, err := filepath.Glob("*")

	if err != nil {
		return err
	}

	for _, filename := range files {
		<-d.sem
		wait.Add(1)
		filename := filename
		go func() {
			// todo: some sort of error channel
			err := fileReader(filename, records)
			if err != nil {
				log.Println(err)
			}
			d.sem <- 1
			wait.Done()
		}()
	}
	wait.Wait()
	close(records)

	return nil
}
