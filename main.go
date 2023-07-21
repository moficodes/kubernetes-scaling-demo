package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const PANEL_WIDTH = 16
const PANEL_HEIGHT = 16

const RED_MASK = 0x00FF0000
const GREEN_MASK = 0x0000FF00
const BLUE_MASK = 0x000000FF

type LedData struct {
	Data []byte `json:"data" firestore:"data"`
}

func singleBoardMapping(rows, cols, position int) [][]int {
	var board [][]int
	for i := 0; i < rows; i++ {
		board = append(board, make([]int, cols))
	}

	value := rows*cols - 1
	max := rows * cols

	for i := 0; i < rows; i++ {
		if i%2 == 0 {
			for j := 0; j < cols; j++ {
				board[i][j] = value + (position * max)
				value--
			}
		} else {
			for j := cols - 1; j >= 0; j-- {
				board[i][j] = value + (position * max)
				value--
			}
		}
	}
	return board
}

func boardMapping(rows, cols int, positions [][]int) []int {
	var board [][]int
	var result []int

	width := len(positions)
	height := len(positions[0])
	for i := 0; i < rows*height; i++ {
		board = append(board, make([]int, cols*width))
	}

	for i := 0; i < len(positions); i++ {
		for j := 0; j < len(positions[i]); j++ {
			sbm := singleBoardMapping(rows, cols, positions[i][j])
			for x := 0; x < len(sbm); x++ {
				for y := 0; y < len(sbm[x]); y++ {
					board[i*rows+x][j*cols+y] = sbm[x][y]
				}
			}
		}
	}

	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[i]); j++ {
			result = append(result, board[i][j])
		}
	}
	return result
}

func print2DArr(data []int) {
	for i := 0; i < len(data); i++ {
		if i%64 == 0 {
			fmt.Println()
		}
		fmt.Printf("%-4d ", data[i])
	}
}

const (
	UNDEFINED int = iota
	ACTIVE
	IDLE
	TERMINATED
)

type Instance struct {
	Id           string    `json:"id" firestore:"id,omitempty"`
	Status       int       `json:"status" firestore:"status"`
	LastReported time.Time `json:"lastReported" firestore:"lastReported,omitempty"`
}

var instances = make(map[string]Instance)

// func subscribe(projectId, subscription string, ctx context.Context, done chan bool) {
// 	client, err := pubsub.NewClient(ctx, projectId)
// 	if err != nil {
// 		log.Fatalf("Could not create pubsub Client: %v", err)
// 	}
// 	defer client.Close()

// 	sub := client.Subscription(subscription)
// 	sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
// 		instance := Instance{}
// 		err := json.Unmarshal(msg.Data, &instance)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		instances[instance.Id] = instance
// 		log.Println(instance.Id, instance.Status)
// 		msg.Ack()
// 	})

// 	<-done
// }

func worker(ctx context.Context, messageChannel chan *pubsub.Message, wg *sync.WaitGroup) {
	defer wg.Done()
	for msg := range messageChannel {
		instance := Instance{}
		err := json.Unmarshal(msg.Data, &instance)
		if err != nil {
			log.Fatal(err)
		}
		instances[instance.Id] = instance
		log.Println(instance.Id, instance.Status)
		msg.Ack()
	}
}

func getInstances(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
		e := json.NewEncoder(w)
		e.Encode(instances)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func intToThreeBytes(i int) []byte {
	return []byte{byte(i >> 16), byte(i >> 8), byte(i)}
}

func print2D(board [][]int) {
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[i]); j++ {
			fmt.Printf("%-4d ", board[i][j])
		}
		fmt.Println()
	}
}

func processInstancesForLed(mapping []int, board [][]int) []byte {

	keys := make([]string, len(instances))
	i := 0
	for k := range instances {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	counter := 0

	for i := len(board) - 1; i >= 0; i-- {
		for j := 0; j < len(board[i]); j++ {
			if counter < len(keys) {
				instance := instances[keys[counter]]
				if instance.Status == ACTIVE {
					board[i][j] = 0x00FF00
				} else if instance.Status == IDLE {
					board[i][j] = 0xFFEA00
				} else if instance.Status == TERMINATED {
					board[i][j] = 0xFF0000
				}
				counter++
			} else {
				board[i][j] = 0x000000
			}
		}
	}

	resInt := make([]int, len(mapping))

	count := 0
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[i]); j++ {
			resInt[mapping[count]] = board[i][j]
			count++
		}
	}

	var res []byte
	for i := 0; i < len(resInt); i++ {
		res = append(res, intToThreeBytes(resInt[i])...)
	}
	return res
}

func main() {
	done := make(chan bool)

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

	colorGrid := make([][]int, 64)
	for i := 0; i < 64; i++ {
		colorGrid[i] = make([]int, 64)
	}

	ctx := context.Background()
	projectId := "mofilabs-next-demo-02"
	subscription := "instance-status"

	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	pubsubClient, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		log.Fatal(err)
	}
	defer pubsubClient.Close()

	// go subscribe(projectId, subscription, ctx, done)
	t := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-t.C:
				for id, instance := range instances {
					// keep the instance in the map for 10 seconds after it's terminated
					if instance.Status == TERMINATED && time.Since(instance.LastReported) > 10*time.Second {
						fmt.Println("deleting", id)
						delete(instances, id)
					}

					if time.Since(instance.LastReported) > 30*time.Second {
						delete(instances, id)
					}
				}
				client.Collection("scaling").Doc("instances").Set(ctx, instances)
				ledData := LedData{
					Data: processInstancesForLed(mapping, colorGrid),
				}

				client.Collection("led").Doc("data").Set(ctx, ledData)

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
	go srv.ListenAndServe()

	workerCount := 25
	messageChannel := make(chan *pubsub.Message, 1000)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ctx, messageChannel, &wg)
	}

	sub := pubsubClient.Subscription(subscription)

	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		messageChannel <- msg
	})

	if err != nil {
		log.Fatal(err)
	}
}
