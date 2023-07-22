package main

import (
	"math"
	"math/rand"
)

func calculatePi(n int) float64 {
	var inside int = 0

	for i := 0; i < n; i++ {
		x, y := rand.Float64(), rand.Float64()
		if math.Sqrt(x*x+y*y) <= 1 {
			inside++
		}
	}

	return 4 * float64(inside) / float64(n)
}

func generatePrimes(n int) []int {
	var primes []int
	if n >= 2 {
		primes = append(primes, 2)
	}

	for i := 3; i <= n; i += 2 {
		isPrime := true
		for _, prime := range primes {
			if i%prime == 0 {
				isPrime = false
				break
			}
		}

		if isPrime {
			primes = append(primes, i)
		}
	}

	return primes
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}
