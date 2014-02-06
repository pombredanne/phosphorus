package schema

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Source interface {
	GetChannel() (c chan *Record)
}

type SourceField struct {
	Name   string
	Column int
}

type SourceFields []SourceField

func (sf SourceFields) parse(fields []string) (attrs map[string]string) {
	attrs = make(map[string]string)
	for _, f := range sf {
		attrs[f.Name] = fields[f.Column-1]
	}
	return
}

type FileSource struct {
	Fields     SourceFields `json:"fields"`
	IdColumn   int          `json:"id_column"`
	Delimiter  string       `json:"delimiter"`
	Glob       string       `json:"glob"`
	Concurrent int          `json:"concurrent"`
	paths      []string
	c          chan *Record
	wait       sync.WaitGroup
	sem        chan int
}

func (f *FileSource) fill() {
	for _, path := range f.paths {
		<-f.sem
		f.wait.Add(1)
		go f.read(path)
	}
	f.wait.Wait()
	close(f.c)
}

func (f *FileSource) read(path string) {
	defer func() {
		f.wait.Done()
		f.sem <- 1
	}()

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.Comma = rune(f.Delimiter[0])
	for line, err := r.Read(); err != io.EOF; line, err = r.Read() {
		if err != nil {
			panic(err)
		}
		id, err := strconv.ParseUint(line[f.IdColumn-1], 10, 32)
		if err != nil {
			panic(err)
		}
		attrs := f.Fields.parse(line)

		f.c <- &Record{uint32(id), attrs}
	}
}

func (f *FileSource) GetChannel() (c chan *Record, err error) {
	f.paths, err = filepath.Glob(f.Glob)
	if err != nil {
		return
	}

	f.sem = make(chan int, f.Concurrent)
	for i := 0; i < f.Concurrent; i++ {
		f.sem <- 1
	}
	f.c = make(chan *Record)

	go f.fill()
	c = f.c
	return
}
