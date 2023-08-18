package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

const (
	UNDEFINED int = iota
	ACTIVE
	IDLE
	TERMINATED
)

func mustGetEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" && defaultValue != "" {
		value = defaultValue
	}
	if value == "" {
		log.Fatalf("Missing required environment variable %s", key)
	}
	return value
}

func main() {
	projectId := mustGetEnv("PROJECT_ID", "")
	instanceCollection := mustGetEnv("INSTANCE_COLLECTION", "instances")

	randomHostId := uuid.New().String()
	hostName := mustGetEnv("HOSTNAME", randomHostId)
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		log.Fatalln(err)
	}

	defer client.Close()

	rc := &RequestCounter{
		activeRequests: 0,
		mutex:          sync.Mutex{},
	}

	instance := Instance{Id: hostName, LastReported: time.Now()}

	done := make(chan bool)
	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				activeRequests := rc.GetActiveRequests()
				status := IDLE
				if activeRequests > 0 {
					status = ACTIVE
				}

				if status != instance.Status || time.Since(instance.LastReported) > 15*time.Second {
					log.Printf("Status changed from %d to %d\n", instance.Status, status)
					instance.Status = status
					instance.LastReported = time.Now()
					_, err := client.Collection(instanceCollection).Doc(hostName).Set(ctx, instance)
					if err != nil {
						log.Fatalln(err)
					}
				}
			}
		}
	}()

	srv := &http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/sqrt", requestCounterMiddleware(rc, sqrt))
	http.HandleFunc("/pi", requestCounterMiddleware(rc, pi))
	http.HandleFunc("/prime", requestCounterMiddleware(rc, prime))
	http.HandleFunc("/fib", requestCounterMiddleware(rc, fib))

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Active requests: %d", rc.GetActiveRequests())
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for sig := range c {
			if sig == os.Interrupt || sig == syscall.SIGTERM {
				log.Println("Shutting down server")
				done <- true
				instance := Instance{Id: hostName, Status: TERMINATED, LastReported: time.Now()}

				_, err := client.Collection(instanceCollection).Doc(hostName).Set(ctx, instance)
				if err != nil {
					log.Fatalln(err)
				}
				srv.Shutdown(ctx)
			}
		}
	}()

	log.Printf("Starting server %s on port 8080\n", hostName)
	srv.ListenAndServe()
}
