package main

import (
	"flag"
	"log"
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
		worker := &Worker{Port: strconv.Itoa(*startPort + i)}
		go worker.run(i, master)
		time.Sleep(time.Second * 1) // TODO remove this after testing
	}

	for {
	} // run forever
}
