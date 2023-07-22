package main

import "net/http"

func requestCounterMiddleware(rc *RequestCounter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rc.Increment()
		defer rc.Decrement()
		next(w, r)
	}
}
