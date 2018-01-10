package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

const jsonData = "application/json"

func main() {
	// get command args
	numWorkers := flag.Int("workers", 1, "max number of workers to spawn")
	startPort := flag.Int("port", 80085, "starting port to accept jobs")
	ipPort := flag.String("master", "127.0.0.1:8080",
		"IP address and port of the API server (master node)")
	api := flag.String("api", "/api/v1/", "api root")
	flag.Parse()

	// number of processor cores on system
	coresAvailable := runtime.NumCPU()
	log.Println("Number of processor cores available: " +
		strconv.FormatInt(int64(coresAvailable), 10))
	log.Println("Number of workers: " +
		strconv.FormatInt(int64(*numWorkers), 10))
	log.Println("Starting port (for zeromq): " +
		strconv.FormatInt(int64(*startPort), 10))

	// init the server struct to hold master server info
	master := Server{URLroot: "http://" + *ipPort, URLjobs: "http://" + *ipPort +
		*api + "jobs", URLworkers: "http://" + *ipPort + *api + "workers"}

	for i := 0; i < *numWorkers; i++ {
		worker := &Worker{Port: strconv.Itoa(*numWorkers + i)}
		go worker.run(i, master)
		time.Sleep(time.Second * 1) // TODO remove this after testing
	}

	for {
	} // run forever
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
	// TODO 0mq stuff here

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
		time.Sleep(time.Second * 15)

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

func (job *Job) setRunning(master *Server, wrkr *Worker) error {
	job.Running = true
	job.WorkerID = wrkr.ID
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
