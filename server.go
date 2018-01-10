package main

import (
	"time"
)

type Server struct {
	URLroot     string    `json:"urlRoot"`
	URLjobs     string    `json:"urlJobs"`
	URLworkers  string    `json:"urlWorkers"`
	Valid       bool      `json:"valid"`       // updateable
	LastContact time.Time `json:"lastContact"` // updateable
	LastUpdate  time.Time `json:"lastUpdate"`  // updateable
}

func (s *Server) validate() error {
	// request index (make http connection) to test server
	// then update server struct
	return nil
}
