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
	//resp, err := http.PostForm(requestURL, url.Values{"port": {sPort}})
	if err != nil {
		//log.Printf("worker %d: error setting READY with master server", wn)
		//log.Println(err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Printf(err.Error())
		return err
	}
	resp.Body.Close()
	if err = json.Unmarshal(body, job); err != nil {
		//log.Printf(err.Error())
		return err
	}
	if job.Valid != true {
		//log.Printf("worker %d: master server returned worker object with false VALID flag when setting READY!", wn)
		return errors.New("master server response was returned as invalid")
	}
	master.Valid = true
	master.LastContact = time.Now()
	master.LastUpdate = time.Now()

	return nil
}

func (job *Job) setComplete(master *Server, wrkr *Worker) error {
	job.Running = false
	job.Completed = true // TODO if job did complete correctly, then set error
	job.Ended = time.Now()
	job.Result = "IT WORKS??????" // TODO change this to the programs output
	//job.WorkerID = wrkr.ID

	jsonWorker, _ := json.Marshal(*job)
	resp, err := http.Post(master.URLjobs+"/"+strconv.Itoa(job.ID), jsonData, bytes.NewBuffer(jsonWorker))
	//resp, err := http.PostForm(requestURL, url.Values{"port": {sPort}})
	if err != nil {
		//log.Printf("worker %d: error setting READY with master server", wn)
		//log.Println(err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Printf(err.Error())
		return err
	}
	resp.Body.Close()
	if err = json.Unmarshal(body, job); err != nil {
		//log.Printf(err.Error())
		return err
	}
	if job.Valid != true {
		//log.Printf("worker %d: master server returned worker object with false VALID flag when setting READY!", wn)
		return errors.New("master server response was returned as invalid")
	}
	master.Valid = true
	master.LastContact = time.Now()
	master.LastUpdate = time.Now()

	return nil
}
