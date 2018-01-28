package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	gostock "github.com/andocmdo/gostockd/common"
)

// Worker is an alias for adding our own methods to the common Worker struct
type Worker gostock.Worker

// NewWorker is a constructor for Worker structs (init Args map)
func NewWorker() *Worker {
	var w Worker
	//j.Args = make(map[string]string)
	return &w
}

func (wrkr *Worker) run(wn int, master Server) {
	log.Printf("thread %d started", wn)

	// register worker with gostockd api server
	if err := wrkr.register(&master); err != nil {
		// handle error
		log.Printf("thread %d worker %d : Error registering with master server", wn, wrkr.ID)
		log.Printf(err.Error())
		return
	}
	log.Printf("thread %d worker %d : Successfully registered with master server", wn, wrkr.ID)
	//time.Sleep(time.Second * 5)

	// initialize an empty job to place incoming JSON job
	job := NewJob()

	// loop here
	for {
		// open listening port for jobs
		ln, err := net.Listen("tcp", ":"+wrkr.Port)
		if err != nil {
			// handle error
			log.Printf("thread %d worker: %d : encountered an error opening listening TCP port "+wrkr.Port, wn, wrkr.ID)
			log.Printf(err.Error())
		}
		// update server showing this worker as ready to accept jobs
		if err := wrkr.setReady(&master); err != nil {
			//handle error
			log.Printf("thread %d worker %d : Error setting READY with master server", wn, wrkr.ID)
			log.Printf(err.Error())
			return
		}
		log.Printf("thread %d worker %d : Successfully notified master server, READY to accept jobs", wn, wrkr.ID)

		// wait/listen to port for incoming jobs
		conn, err := ln.Accept()

		// using json decoder now
		enc := json.NewEncoder(conn) // Will write to network.
		dec := json.NewDecoder(conn) // Will read from network.

		err = dec.Decode(&job)
		if err != nil {
			log.Printf("thread %d worker: %d : encountered an error opening listening TCP port "+wrkr.Port, wn, wrkr.ID)
			log.Printf(err.Error())
		}
		log.Printf("thread %d worker %d : Recieved job # %d ", wn, wrkr.ID, job.ID)
		//log.Printf("%+v", job)

		// return what we recieved as confirmation, except update dispatched variable
		// TODO check if valid!
		job.Dispatched = true
		err = enc.Encode(&job)
		conerr := conn.Close()
		if conerr != nil {
			log.Printf("thread %d worker %d : error closing conn for job %d", wn, wrkr.ID, job.ID)
		}
		lnerr := ln.Close()
		if lnerr != nil {
			log.Printf("thread %d worker %d : error closing ln for job %d", wn, wrkr.ID, job.ID)
		}

		//time.Sleep(time.Second * 5)

		// update server that we are working, and that job is running on this worker
		if wrkerr := wrkr.setWorking(&master, job); err != nil {
			//handle error
			log.Printf("thread %d worker %d : Error setting WORKING with master server for job ID %d", wn, wrkr.ID, job.ID)
			log.Printf(wrkerr.Error())
			return
		}
		log.Printf("thread %d worker %d : updated master for running job %d", wn, wrkr.ID, job.ID)

		//update server that job is running on this worker
		if runerr := job.setRunning(&master, wrkr); err != nil {
			//handle error
			log.Printf("thread %d worker %d : Error setting job %d running with master server", wn, wrkr.ID, job.ID)
			log.Printf(runerr.Error())
			return
		}
		log.Printf("thread %d worker %d : updated master (setrunning) for running job %d", wn, wrkr.ID, job.ID)

		//  Do some 'work'
		// in this test we are going to sleep and also run 'echo' command
		log.Printf("thread %d worker %d : sleeping for 3s before running job %d", wn, wrkr.ID, job.ID)
		time.Sleep(time.Second * 3)
		log.Printf("thread %d worker %d : starting command for job %d", wn, wrkr.ID, job.ID)

		/*
			cmdString := job.Args["command"] + " test"
			cmd := exec.Command("bash", "-c", cmdString)
			log.Printf("thread %d worker %d : built command for job %d", wn, wrkr.ID, job.ID)
			out, cmderr := cmd.Output()
			log.Printf("thread %d worker %d : ran command for job %d", wn, wrkr.ID, job.ID)
		*/
		cmdString := job.Args["command"] + " test"
		cmd := exec.Command("bash", "-c", cmdString)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmderr := cmd.Run()
		outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
		log.Printf("out: %s \t err: %s\n", outStr, errStr)
		if cmderr != nil {
			log.Printf("thread %d worker %d : ran command but had error for job %d", wn, wrkr.ID, job.ID)
			job.Result = string(errStr)
			job.Success = false
			log.Printf("error: %s", err)

		} else {

			log.Printf("thread %d worker %d : command successful for job %d", wn, wrkr.ID, job.ID)

			job.Result = string(outStr)
			job.Success = true
		}

		log.Printf("thread %d worker %d : completed job %d, result was: %s", wn, wrkr.ID, job.ID, job.Result)

		// after job finishes, update job
		log.Printf("thread %d worker %d : completed job %d", wn, wrkr.ID, job.ID)
		if completeErr := job.setComplete(&master, wrkr); err != nil {
			//handle error
			log.Printf("thread %d worker %d : Error setting job %d complete with master server", wn, wrkr.ID, job.ID)
			log.Printf(completeErr.Error())
			return
		}
		log.Printf("thread %d worker %d : updated master (setComplete) for completed job %d", wn, wrkr.ID, job.ID)

	}
}

func (wrkr *Worker) register(master *Server) error {
	jsonWorker, _ := json.Marshal(wrkr)
	resp, err := http.Post(master.URLworkers, jsonData, bytes.NewBuffer(jsonWorker))
	//resp, err := http.PostForm(requestURL, url.Values{"port": {wrkr.Port}})
	if err != nil {
		//log.Printf("worker %d: error registering with master server", wn)
		log.Print("error registering worker with master server: ", err.Error())
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print("error registering worker with master server: ", err.Error())
		return err
	}
	resp.Body.Close()
	if err = json.Unmarshal(body, &wrkr); err != nil {
		log.Print("error registering worker with master server: ", err.Error())
		return err
	}
	if wrkr.Valid != true {
		//log.Printf("worker %d: master server returned worker object with valid flag set false!", wn)
		return errors.New("master server response was returned as invalid")
	}
	//log.Printf("worker %d: Successfully registered with master server", wn)
	master.Valid = true
	master.LastContact = time.Now()
	master.LastUpdate = time.Now()

	return nil
}

func (wrkr *Worker) setReady(master *Server) error {
	// set this worker as ready to accept jobs
	wrkr.Ready = true
	wrkr.Working = false
	jsonWorker, _ := json.Marshal(wrkr)
	resp, err := http.Post(master.URLworkers+"/"+strconv.Itoa(wrkr.ID), jsonData, bytes.NewBuffer(jsonWorker))
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
	if err = json.Unmarshal(body, &wrkr); err != nil {
		//log.Printf(err.Error())
		return err
	}
	if wrkr.Valid != true {
		//log.Printf("worker %d: master server returned worker object with false VALID flag when setting READY!", wn)
		return errors.New("master server response was returned as invalid")
	}
	master.Valid = true
	master.LastContact = time.Now()
	master.LastUpdate = time.Now()

	return nil
}

func (wrkr *Worker) setWorking(master *Server, job *Job) error {
	// set this worker as ready to accept jobs
	wrkr.Ready = false
	wrkr.Working = true
	wrkr.JobID = job.ID

	jsonWorker, _ := json.Marshal(wrkr)
	resp, err := http.Post(master.URLworkers+"/"+strconv.Itoa(wrkr.ID), jsonData, bytes.NewBuffer(jsonWorker))
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
	if err = json.Unmarshal(body, &wrkr); err != nil {
		//log.Printf(err.Error())
		return err
	}
	if wrkr.Valid != true {
		//log.Printf("worker %d: master server returned worker object with false VALID flag when setting READY!", wn)
		return errors.New("master server response was returned as invalid")
	}
	master.Valid = true
	master.LastContact = time.Now()
	master.LastUpdate = time.Now()

	return nil
}

// TODO write completion server update method
func (wrkr *Worker) setComplete(master *Server, job *Job) error {
	// set this worker as ready to accept jobs
	wrkr.Ready = true
	wrkr.Working = false
	wrkr.JobID = job.ID

	jsonWorker, _ := json.Marshal(wrkr)
	resp, err := http.Post(master.URLworkers+"/"+strconv.Itoa(wrkr.ID), jsonData, bytes.NewBuffer(jsonWorker))
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
	if err = json.Unmarshal(body, &wrkr); err != nil {
		//log.Printf(err.Error())
		return err
	}
	if wrkr.Valid != true {
		//log.Printf("worker %d: master server returned worker object with false VALID flag when setting READY!", wn)
		return errors.New("master server response was returned as invalid")
	}
	master.Valid = true
	master.LastContact = time.Now()
	master.LastUpdate = time.Now()

	return nil
}
