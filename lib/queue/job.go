package queue

import (
	"bytes"
	"encoding/gob"
)

// JobOptions defines the options of a job
type JobOptions struct {
	JobID string
}

// Job defines a job
type Job struct {
	Name    string
	Data    []byte
	Options *JobOptions
}

// Decode decodes the job data
func (job *Job) Decode(e interface{}) error {
	d := gob.NewDecoder(bytes.NewReader(job.Data))
	return d.Decode(e)
}
