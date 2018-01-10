package main

import (
	"time"
)

// Worker contains state data for workers
type Worker struct {
	ID         int       `json:"id"`
	JobID      int       `json:"jobID"` // updateable
	Valid      bool      `json:"valid"`
	Created    time.Time `json:"created"`
	IPAddr     string    `json:"ipaddr"`
	Port       string    `json:"port"`
	Ready      bool      `json:"ready"`      // updateable
	Working    bool      `json:"working"`    // updateable
	LastUpdate time.Time `json:"lastUpdate"` // updateable
}
