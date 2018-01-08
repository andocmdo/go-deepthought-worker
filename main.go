package main

import (
	"log"
	"runtime"
	"strconv"
)

// TODO parameterize these
const keepFreeCores = 1
const startPort = 80085
const masterIP = "192.168.1.32"
const masterPort = "8080"

func main() {
	// number of processor cores to keep free, the rest will be used to run jobs
	cores := 1
	coresAvailable := runtime.NumCPU()
	log.Println("Number of processor cores available: " + strconv.FormatInt(int64(coresAvailable), 10))
	if coresAvailable > keepFreeCores {
		cores = coresAvailable - keepFreeCores
	}
	log.Println("Number of processor cores to use: ", cores)

	for i := 0; i < cores; i++ {
		go worker(i)
	}
}

func worker(w int) {
	log.Printf("started worker %d", w)

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
