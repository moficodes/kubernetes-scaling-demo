package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const (
	PANEL_WIDTH  = 16
	PANEL_HEIGHT = 16
)

const (
	UNDEFINED int = iota
	ACTIVE
	IDLE
	TERMINATED
)

var instances = make(map[string]Instance)

func mustGetEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" && defaultValue != "" {
		value = defaultValue
	}
	if value == "" {
		log.Fatalf("Environment variable %s must be set", key)
	}
	return value
}

func main() {
	projectId := mustGetEnv("PROJECT_ID", "")

	done := make(chan bool)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	http.HandleFunc("/instances", getInstances)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Println(err)
		}
		log.Println("connection upgraded to ws")
		go func() {
			defer conn.Close()

			for {
				msg, err := json.Marshal(instances)
				if err != nil {
					log.Println(err)
					return
				}
				err = wsutil.WriteServerMessage(conn, ws.OpText, msg)
				if err != nil {
					log.Println(err)
					return
				}
				<-done
			}
		}()
	})

	panelPositions := [][]int{
		{12, 13, 14, 15},
		{8, 9, 10, 11},
		{4, 5, 6, 7},
		{0, 1, 2, 3},
	}

	mapping := boardMapping(PANEL_HEIGHT, PANEL_WIDTH, panelPositions)

	colorGrid := make([][]int, PANEL_HEIGHT*len(panelPositions))
	for i := 0; i < len(colorGrid); i++ {
		colorGrid[i] = make([]int, PANEL_WIDTH*len(panelPositions[0]))
	}

	fmt.Println(len(colorGrid))
	fmt.Println(len(colorGrid[0]))

	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	// go subscribe(projectId, subscription, ctx, done)
	t := time.NewTicker(3 * time.Second)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-t.C:
				instances = make(map[string]Instance)
				iter := client.Collection("instances").Documents(ctx)
				for {
					doc, err := iter.Next()
					if err != nil {
						break
					}
					if err != nil {
						log.Fatal(err)
					}
					instance := Instance{doc.Data()["Id"].(string), int(doc.Data()["Status"].(int64)), doc.Data()["LastReported"].(time.Time)}
					instances[instance.Id] = instance

					if instance.Status == TERMINATED && time.Since(instance.LastReported) > 5*time.Second {
						client.Collection("instances").Doc(instance.Id).Delete(ctx)
					}

					if time.Since(instance.LastReported) > 30*time.Second {
						client.Collection("instances").Doc(instance.Id).Set(ctx, map[string]interface{}{
							"Id":           instance.Id,
							"Status":       TERMINATED,
							"LastReported": time.Now(),
						}, firestore.MergeAll)
					}
				}

				ledData := LedData{
					Data: processInstancesForLed(mapping, colorGrid, instances),
				}

				_, err := client.Collection("led").Doc("data").Set(ctx, ledData)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	// catch SIGINT and do a graceful shutdown

	srv := &http.Server{Addr: ":8000"}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for sig := range c {
			if sig == os.Interrupt || sig == syscall.SIGTERM {
				log.Println("Shutting down server")
				done <- true

				srv.Shutdown(ctx)
			}
		}
	}()

	log.Println("Starting server on port 8000")
	srv.ListenAndServe()
}
