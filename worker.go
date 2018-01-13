package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	gostock "github.com/andocmdo/gostockd/common"
	zmq "github.com/pebbe/zmq4"
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
	job := NewJob() // initialize an empty job to place incoming JSON job
	//  Socket to talk to clients
	responder, _ := zmq.NewSocket(zmq.REP)
	responder.Connect("tcp://" + wrkr.IPAddr + ":" + wrkr.Port)
	defer responder.Close()

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
		//  Wait for next request from client
		request, _ := responder.Recv(0)
		fmt.Printf("Received request: [%s]\n", request)

		//  Do some 'work'
		time.Sleep(time.Second)

		//  Send reply back to client
		responder.Send("World", 0)
		time.Sleep(time.Second)

		// decode incoming job
		log.Printf("thread %d worker %d : recieved job", wn, wrkr.ID)
		time.Sleep(time.Second * 5)

		// start job
		log.Printf("thread %d worker %d : started job", wn, wrkr.ID)
		time.Sleep(time.Second * 5)

		// update server that we are working, and that job is running on this worker
		if err := wrkr.setWorking(&master, job); err != nil {
			//handle error
			log.Printf("thread %d worker %d : Error setting WORKING with master server", wn, wrkr.ID)
			log.Printf(err.Error())
			return
		}
		log.Printf("thread %d worker %d : updated master for running job NUMBER WHAT?", wn, wrkr.ID)
		if err := job.setRunning(&master, wrkr); err != nil {
			//handle error
			log.Printf("thread %d worker %d : Error setting job running with master server", wn, wrkr.ID)
			log.Printf(err.Error())
			return
		}
		log.Printf("thread %d worker %d : updated master for running job NUMBER WHAT?", wn, wrkr.ID)
		time.Sleep(time.Second * 25)

		// wait for job to finish
		log.Printf("thread %d worker %d : completed job ??", wn, wrkr.ID)

		// update server of job completion status

		log.Printf("thread %d worker %d : updated master of successful job ?? completion", wn, wrkr.ID)
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
