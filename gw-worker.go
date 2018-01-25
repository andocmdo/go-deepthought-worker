package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
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
	log.Printf("thread %d worker %d : started", wn, wrkr.ID)

	// register worker with gostockd api server
	if err := wrkr.register(&master); err != nil {
		// handle error
		log.Printf("thread %d worker %d : Error registering with master server", wn, wrkr.ID)
		log.Printf(err.Error())
		return
	}
	log.Printf("thread %d worker %d : Successfully registered with master server", wn, wrkr.ID)
	time.Sleep(time.Second * 5)

	// open listening port for jobs
	ln, err := net.Listen("tcp", ":"+wrkr.Port)
	defer ln.Close()
	if err != nil {
		// handle error
		log.Printf("thread %d worker: %d : encountered an error opening listening TCP port "+wrkr.Port, wn, wrkr.ID)
		log.Printf(err.Error())
	}

	// initialize an empty job to place incoming JSON job
	job := NewJob()

	// loop here
	for {
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
		log.Printf("%+v", job)

		// return what we recieved as confirmation, except update dispatched variable
		// TODO check if valid!
		job.Dispatched = true
		err = enc.Encode(&job)
		conn.Close()

		// update server that we are working, and that job is running on this worker
		if err := wrkr.setWorking(&master, job); err != nil {
			//handle error
			log.Printf("thread %d worker %d : Error setting WORKING with master server for job ID %d", wn, wrkr.ID, job.ID)
			log.Printf(err.Error())
			return
		}
		log.Printf("thread %d worker %d : updated master for running job %d", wn, wrkr.ID, job.ID)

		//update server that job is running on this worker
		if err := job.setRunning(&master, wrkr); err != nil {
			//handle error
			log.Printf("thread %d worker %d : Error setting job %d running with master server", wn, wrkr.ID, job.ID)
			log.Printf(err.Error())
			return
		}
		log.Printf("thread %d worker %d : updated master (setrunning) for running job %d", wn, wrkr.ID, job.ID)

		//  Do some 'work'
		time.Sleep(time.Second * 25)

		// after job finishes
		log.Printf("thread %d worker %d : completed job %d", wn, wrkr.ID, job.ID)

		// update server of job completion status

		log.Printf("thread %d worker %d : SHOULD updated master of successful job ?? completion", wn, wrkr.ID)
		conn.Close()
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

	log.Printf("called setWorking")
	log.Printf("%+v", wrkr)
	log.Printf("%+v", job)

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

	log.Printf("exiting setWorking")
	log.Printf("%+v", wrkr)
	log.Printf("%+v", job)

	return nil
}

// TODO write completion server update method
