package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	gostock "github.com/andocmdo/gostockd/common"
)

// Job is an alias for adding our own methods to the common gostock.Job struct
type Job gostock.Job

// NewJob is a constructor for Job structs (init Args map)
func NewJob() *Job {
	var j Job
	j.Args = make(map[string]string)
	return &j
}

func (job *Job) setRunning(master *Server, wrkr *Worker) error {
	job.Running = true
	job.WorkerID = wrkr.ID
	job.Started = time.Now()

	jsonWorker, _ := json.Marshal(*job)
	resp, err := http.Post(master.URLjobs+"/"+strconv.Itoa(job.ID), jsonData, bytes.NewBuffer(jsonWorker))
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if err = json.Unmarshal(body, job); err != nil {
		return err
	}
	if job.Valid != true {
		return errors.New("master server response was returned as invalid")
	}
	master.Valid = true
	master.LastContact = time.Now()
	master.LastUpdate = time.Now()

	return nil
}

func (job *Job) setComplete(master *Server, wrkr *Worker) error {
	job.Running = false
	job.Completed = true // TODO if job did not complete correctly, then set error
	job.Ended = time.Now()

	jsonWorker, _ := json.Marshal(*job)
	resp, err := http.Post(master.URLjobs+"/"+strconv.Itoa(job.ID), jsonData, bytes.NewBuffer(jsonWorker))

	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if err = json.Unmarshal(body, job); err != nil {
		return err
	}
	if job.Valid != true {
		return errors.New("master server response was returned as invalid")
	}
	master.Valid = true
	master.LastContact = time.Now()
	master.LastUpdate = time.Now()

	return nil
}
