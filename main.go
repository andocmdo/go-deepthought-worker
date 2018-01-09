package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

var numWorkers, startPort int
var apiIP string

const workerAPI = "/api/v1/workers"
const jobAPI = "/api/v1/jobs"
const jsonData = "application/json"

// Worker contains state data for workers
type Worker struct {
	ID         int       `json:"id"`
	Valid      bool      `json:"valid"`
	Created    time.Time `json:"created"`
	IPAddr     string    `json:"ipaddr"`
	Port       string    `json:"port"`
	Ready      bool      `json:"ready"`      // updateable
	Working    bool      `json:"working"`    // updateable
	LastUpdate time.Time `json:"lastUpdate"` // updateable
}

func main() {
	// get command args
	flag.IntVar(&numWorkers, "workers", 1, "max number of workers to spawn")
	flag.IntVar(&startPort, "port", 80085, "starting port to accept jobs")
	flag.StringVar(&apiIP, "master", "127.0.0.1:8080", "IP address of the API server (master node)")
	flag.Parse()

	// number of processor cores on system
	coresAvailable := runtime.NumCPU()
	log.Println("Number of processor cores available: " + strconv.FormatInt(int64(coresAvailable), 10))
	log.Println("Number of workers: " + strconv.FormatInt(int64(numWorkers), 10))
	log.Println("Starting port (for zeromq): " + strconv.FormatInt(int64(startPort), 10))

	for i := 0; i < numWorkers; i++ {
		go worker(i, startPort)
		time.Sleep(time.Second * 2)
	}

	for {
	} // run forever
}

func worker(wn int, port int) {
	log.Printf("worker %d: started", wn)
	sPort := strconv.Itoa(port)
	var thisWorker Worker
	//workerID := -1
	//var buf bytes.Buffer

	// register worker with gostockd api server
	thisWorker.Port = sPort
	jsonWorker, _ := json.Marshal(thisWorker)
	requestURL := "http://" + apiIP + workerAPI
	resp, err := http.Post(requestURL, jsonData, bytes.NewBuffer(jsonWorker))
	//resp, err := http.PostForm(requestURL, url.Values{"port": {sPort}})
	if err != nil {
		log.Printf("worker %d: error registering with master server", wn)
		log.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	resp.Body.Close()
	if err = json.Unmarshal(body, &thisWorker); err != nil {
		log.Printf(err.Error())
		return
	}
	if thisWorker.Valid != true {
		log.Printf("worker %d: master server returned worker object with valid flag set false!", wn)
		return
	}
	log.Printf("worker %d: Successfully registered with master server", wn)
	time.Sleep(time.Second * 5)

	// open listening port for jobs
	// TODO 0mq stuff here

	// loop here
	for {
		// update server showing this worker as ready to accept jobs
		requestURL := "http://" + apiIP + workerAPI + "/" + strconv.Itoa(thisWorker.ID)
		// set this worker as ready to accept jobs
		thisWorker.Ready = true
		jsonWorker, _ = json.Marshal(thisWorker)
		resp, err := http.Post(requestURL, jsonData, bytes.NewBuffer(jsonWorker))
		//resp, err := http.PostForm(requestURL, url.Values{"port": {sPort}})
		if err != nil {
			log.Printf("worker %d: error setting READY with master server", wn)
			log.Println(err)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf(err.Error())
			return
		}
		resp.Body.Close()
		if err = json.Unmarshal(body, &thisWorker); err != nil {
			log.Printf(err.Error())
			return
		}
		if thisWorker.Valid != true {
			log.Printf("worker %d: master server returned worker object with false VALID flag when setting READY!", wn)
			return
		}
		log.Printf("worker %d: Successfully notified master server, READY to accept jobs", wn)

		// wait/listen to port for incoming jobs
		time.Sleep(time.Second * 15)

		// decode incoming job
		log.Printf("worker %d: recieved job", wn)
		time.Sleep(time.Second * 5)

		// start job
		log.Printf("worker %d: started job", wn)
		time.Sleep(time.Second * 5)

		// update server that we have started job
		log.Printf("worker %d: updated master for running job", wn)
		time.Sleep(time.Second * 5)

		// wait for job to finish
		log.Printf("worker %d: completed job", wn)
		time.Sleep(time.Second * 5)

		// update server of job completion status
		log.Printf("worker %d: updated master of successful job completion", wn)
	}
}
