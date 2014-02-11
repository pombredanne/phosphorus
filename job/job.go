package job

import ()

type State int

const (
	JOB_WAIT State = iota
	JOB_RUN
	JOB_DEAD
	JOB_OK
)

// type Job interface {
// 	State() State
// 	Kill() error
// 	Run() error
// }

// type Queue interface {
// 	Push(Job) error
// 	Pop() (Job, error)
// }

// type IdService interface {
// 	NextId() int64
// }

// type JobDescription struct {
// 	Name     string
// 	Instance func(string) (JobF, error)
// }

// type JobF func() error

// type Job struct {
// 	IndexId  int64  `dynamodb:"_hash"`
// 	Id       int64  `dynamodb:"_range"`
// 	Type     string `dynamodb:"type",json:"-"`
// 	Argument string `dynamodb:"argument",json:"-"`
// 	State    State  `dynamodb:"state",json:"-"`
// 	Fn       *JobF  `json:"-"`
// }

// type JobT struct {

// }
