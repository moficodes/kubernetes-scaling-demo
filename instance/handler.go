package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

func sqrt(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":

		for i := 0; i < 1000000; i++ {
			_ = math.Sqrt(float64(i))
		}

		fmt.Fprintf(w, "calculated %d sqrt", 1_000_000)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func pi(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		count := mustGetEnv("PI_COUNT", "1000000")
		countInt, err := strconv.Atoi(count)
		if err != nil {
			countInt = 1000000
		}
		pi := calculatePi(countInt)
		fmt.Fprintf(w, "calculated pi: %f", pi)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func prime(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		time.Sleep(200 * time.Millisecond)
		primes := generatePrimes(10_000)
		fmt.Fprintf(w, "found %d primes", len(primes))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func fib(w http.ResponseWriter, r *http.Request) {
	// Getting the Fibonacci term from the query, defaulting to 40 if not provided
	nStr := r.URL.Query().Get("n")
	if nStr == "" {
		nStr = "40"
	}
	n, err := strconv.Atoi(nStr)
	if err != nil {
		http.Error(w, "Invalid n provided", http.StatusBadRequest)
		return
	}

	// Calculate the nth Fibonacci number in an endless loop

	res := fibonacci(n)

	fmt.Fprintf(w, "Fibonacci(%d) = %d", n, res)
}
