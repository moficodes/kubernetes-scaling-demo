package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
var cloudRunInstances = make(map[string]Instance)
var gkeInstances = make(map[string]Instance)

var skip = false

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
	ledCollection := mustGetEnv("FIRESTORE_COLLECTION", "led")
	cloudrunCollection := mustGetEnv("CLOUDRUN_COLLECTION", "cloudrun")
	gkeCollection := mustGetEnv("GKE_COLLECTION", "gke")
	// instanceCollection := mustGetEnv("INSTANCE_COLLECTION", "instances")
	done := make(chan bool)

	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	panelPositions := [][]int{
		{12, 13, 14, 15},
		{8, 9, 10, 11},
		{4, 5, 6, 7},
		{0, 1, 2, 3},
	}

	mapping := boardMapping(PANEL_HEIGHT, PANEL_WIDTH, panelPositions)

	mappingData := MappingData{
		Data: mapping,
	}

	_, err = client.Collection("mapping").Doc("data").Set(ctx, mappingData)
	if err != nil {
		log.Println(err)
	}

	colorGrid := make([][]int, PANEL_HEIGHT*len(panelPositions))
	for i := 0; i < len(colorGrid); i++ {
		colorGrid[i] = make([]int, PANEL_WIDTH*len(panelPositions[0]))
	}

	e := echo.New()

	e.Use(
		middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "method=${method}, uri=${uri}, status=${status}\n"}),
		middleware.CORSWithConfig(
			middleware.CORSConfig{
				AllowOrigins: []string{"*"},
			},
		),
		middleware.GzipWithConfig(middleware.GzipConfig{
			Level: 5,
		}),
		middleware.Secure(),
		middleware.StaticWithConfig(middleware.StaticConfig{
			Root:  "public",
			HTML5: true,
		}),
	)

	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	e.GET("/instances", getInstances)
	e.POST("/skip", skipRender)
	e.POST("/unskip", unskipRender)
	// e.POST("/direct", directRender(client, ledCollection, instanceCollection, mapping))
	e.POST("/gameoflife", gameOfLife)
	e.POST("/upload", fileUpload)

	t := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-t.C:
				if skip {
					continue
				}
				// instances = make(map[string]Instance)
				cloudRunInstances = make(map[string]Instance)
				gkeInstances = make(map[string]Instance)

				cloudRunIter := client.Collection(cloudrunCollection).Documents(ctx)
				for {
					doc, err := cloudRunIter.Next()
					if err != nil {
						break
					}
					if err != nil {
						log.Fatal(err)
					}
					instance := Instance{doc.Data()["Id"].(string), int(doc.Data()["Status"].(int64)), doc.Data()["LastReported"].(time.Time)}
					cloudRunInstances[instance.Id] = instance

					if instance.Status == TERMINATED && time.Since(instance.LastReported) > 5*time.Second {
						_, err := client.Collection(cloudrunCollection).Doc(instance.Id).Delete(ctx)
						if err != nil {
							log.Println(err)
						}
					}

					if time.Since(instance.LastReported) > 30*time.Second {
						client.Collection(cloudrunCollection).Doc(instance.Id).Set(ctx, map[string]interface{}{
							"Id":           instance.Id,
							"Status":       TERMINATED,
							"LastReported": time.Now(),
						}, firestore.MergeAll)
					}
				}

				gkeIter := client.Collection(gkeCollection).Documents(ctx)
				for {
					doc, err := gkeIter.Next()
					if err != nil {
						break
					}
					if err != nil {
						log.Fatal(err)
					}
					instance := Instance{doc.Data()["Id"].(string), int(doc.Data()["Status"].(int64)), doc.Data()["LastReported"].(time.Time)}
					gkeInstances[instance.Id] = instance

					if instance.Status == TERMINATED && time.Since(instance.LastReported) > 5*time.Second {
						_, err := client.Collection(gkeCollection).Doc(instance.Id).Delete(ctx)
						if err != nil {
							log.Println(err)
						}
					}

					if time.Since(instance.LastReported) > 30*time.Second {
						client.Collection(gkeCollection).Doc(instance.Id).Set(ctx, map[string]interface{}{
							"Id":           instance.Id,
							"Status":       TERMINATED,
							"LastReported": time.Now(),
						}, firestore.MergeAll)
					}
				}

				ledData := LedData{
					Data: processInstancesForLed(mapping, colorGrid, cloudRunInstances, gkeInstances),
				}

				_, err := client.Collection(ledCollection).Doc("data").Set(ctx, ledData)
				if err != nil {
					log.Println(err)
				}
				log.Printf("wrote %d bytes to firestore", len(ledData.Data))
			}
		}
	}()

	// catch SIGINT and do a graceful shutdown

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for sig := range c {
			if sig == os.Interrupt || sig == syscall.SIGTERM {
				log.Println("Shutting down server")
				close(done)
				e.Shutdown(ctx)
			}
		}
	}()

	log.Println("Starting server on port 8000")
	e.Logger.Fatal(e.Start(":8000"))
}
