package data

import (
	// "fmt"
	"encoding/csv"
	"io"
	"log"
	"os"
	// "path"
	"github.com/crowdmob/goamz/dynamodb"
	"path/filepath"
	"strconv"
	"sync"
	"willstclair.com/phosphorus/environment"
)

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

type Mapper func(interface{}) (interface{}, error)

type File struct {
	Mappers []Mapper
	Stream  chan interface{}
}

func (f *File) Load(r io.Reader) (err error) {
	defer r.Close()

	rdr := csv.NewReader(f.handle)

outer:
	for line, err := rdr.Read(); err != io.EOF; line, err = rdr.Read() {
		if err != nil {
			// todo: error channel
			log.Println(err)
			continue
		}

		var record interface{}
		record = line
		for _, fn := range f.Mappers {
			record, err = fn(record)
			if err != nil {
				log.Println(err)
				continue outer
			}
		}
		f.Stream <- record
	}
	return
}

type Attribute struct {
	Column    int
	Name      string
	ShortName string
}

type Description struct {
	Attributes []Attribute
}

func (d *Description) Parse(fields []string) (r *Record) {
	r = &Record{strconv.ParseUint(fields[0], 0, 32), make(map[string]string)}
	for _, a := range d.Attributes {
		r.Fields[a.ShortName] = fields[a.Column]
	}

	return
}

type Record struct {
	RecordId uint32
	Fields   map[string]string
}

func (r *Record) ToItem() (item *environment.Item) {
	var attrs []dynamodb.Attribute

	for k, v := range r.Fields {
		attrs = append(attrs, dynamodb.NewStringAttribute(k, v))
	}

	item = &environment.Item{
		dynamodb.Key{environment.Enc64(r.RecordId), ""},
		attrs}
	return
}
