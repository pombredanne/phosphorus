package schema

import ()

type Source interface {
	GetChannel() (c chan *Record)
}

type FileSource struct {
}

type S3Source struct {
}
