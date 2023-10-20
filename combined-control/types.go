package main

import "time"

type LedData struct {
	Data []byte `json:"data" firestore:"data"`
}

type PixelGrid struct {
	Pixels [][]int `json:"pixels" firestore:"pixels"`
}

type MappingData struct {
	Data []int `json:"data" firestore:"data"`
}

type Instance struct {
	Id           string    `json:"id" firestore:"id,omitempty"`
	Status       int       `json:"status" firestore:"status"`
	LastReported time.Time `json:"lastReported" firestore:"lastReported,omitempty"`
}
