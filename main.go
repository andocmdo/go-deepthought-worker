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

	/*
		id := 69
		// This is where we would process our job
		cmd := exec.Command("bash", "-c", "sleep 10; date")
		out, err := cmd.Output()
		if err != nil {
			log.Printf("error on job %d", id)
			log.Printf(err.Error())
		}
		log.Printf("Job %d output: %s", id, out)

		// And when finished, note the time, check for errors, etc

				job.Ended = time.Now()
				job.Running = false
				job.Completed = true
				job.Result = string(out)


			_, err = RepoUpdateJob(job)
			if err != nil {
				log.Printf("error on job %d", id)
				log.Printf(err.Error())
				return
			}
			log.Printf("worker %d finished job %d", w, id)
	*/
}
