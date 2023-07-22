package main

import (
	"sync"
	"time"
)

type Instance struct {
	Id           string    `json:"id"`
	Status       int       `json:"status"`
	LastReported time.Time `json:"lastReported"`
}

type RequestCounter struct {
	activeRequests int
	mutex          sync.Mutex
}

func (rc *RequestCounter) Increment() {
	rc.mutex.Lock()
	rc.activeRequests++
	rc.mutex.Unlock()
}

func (rc *RequestCounter) Decrement() {
	rc.mutex.Lock()
	rc.activeRequests--
	rc.mutex.Unlock()
}

func (rc *RequestCounter) GetActiveRequests() int {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	return rc.activeRequests
}
