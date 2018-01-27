package main

import (
	"flag"
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
	log.Println("Starting port (for listening TCP port to accept jobs): " +
		strconv.FormatInt(int64(*startPort), 10))

	// init the server struct to hold master server info
	master := Server{URLroot: "http://" + *ipPort, URLjobs: "http://" + *ipPort +
		*api + "jobs", URLworkers: "http://" + *ipPort + *api + "workers"}

	// TODO also when the threads have started, we will wait as well if we lose connection?
	for {
		resp, err := http.Get(master.URLroot)
		//defer resp.Body.Close()
		if err == nil {
			resp.Body.Close()
			break
		}
		log.Print("Error connecting to master server. Is it running?)
		log.Print("Error: ", err.Error())
		log.Print("Retry connection to master in 30 secs")
		time.Sleep(time.Second * 30)

	}

	for i := 0; i < *numWorkers; i++ {
		worker := &Worker{Port: strconv.Itoa(*startPort + i)}
		go worker.run(i, master)
		time.Sleep(time.Second * 1) // TODO remove this after testing
	}

	for {
	} // run forever TODO look for an error condition, maybe use a mutex to set error
}
