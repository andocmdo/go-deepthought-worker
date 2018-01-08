package main

import (
	"flag"
	"log"
	"runtime"
	"strconv"
)

func main() {
	// get command args
	numWorkersPtr := flag.Int("workers", 1, "max number of workers to spawn")
	startPortPtr := flag.Int("port", 80085, "starting port to accept jobs")
	flag.Parse()

	// number of processor cores on system
	coresAvailable := runtime.NumCPU()
	log.Println("Number of processor cores available: " + strconv.FormatInt(int64(coresAvailable), 10))

	for i := 0; i < *numWorkersPtr; i++ {
		go worker(i, *startPortPtr)
	}
}

func worker(wnum int, startPort int) {
	log.Printf("started worker %d accepting jobs on port %d", wnum, startPort)

	// register worker with gostockd api server

	// open listening port for jobs

	// while loop here
	// update server showing this worker as ready to accept jobs

	// wait/listen to port for incoming jobs

	// decode incoming job

	// start job

	// update server that we have started job

	// wait for job to finish

	// update server of job completion status
	// back to while loop

}
