package main

import (
	"encoding/csv"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rakyll/hey/requester"
)

type HeyConfig struct {
	URL         string `json:"url"`
	Request     int    `json:"request"`
	Concurrency int    `json:"concurrency"`
	Duration    int    `json:"duration"`
	Timeout     int    `json:"timeout"`
}

type Result struct {
	Requests int     `json:"requests"`
	P50      float64 `json:"p50"`
	P90      float64 `json:"p90"`
	P95      float64 `json:"p95"`
	P99      float64 `json:"p99"`

	StatusCodes map[string]int `json:"statusCodes"`
}

func getEnvOrDefault(key, def string) string {
	env := os.Getenv(key)
	if env == "" {
		env = def
	}
	return env
}

func getEnvOrDefaultInt(key string, def int) int {
	env := os.Getenv(key)
	if env == "" {
		return def
	}
	i, err := strconv.Atoi(env)
	if err != nil {
		return def
	}
	return i
}

func generate(cfg HeyConfig) func(echo.Context) error {
	return func(c echo.Context) error {
		u, _ := url.Parse(cfg.URL)
		w := &requester.Work{
			Request: &http.Request{
				Method: "GET",
				URL:    u,
			},
			N:       cfg.Request,
			C:       cfg.Concurrency,
			Output:  "csv",
			Timeout: cfg.Timeout,
		}
		w.Init()

		dur := time.Duration(cfg.Duration) * time.Second

		if dur > 0 {
			go func() {
				time.Sleep(dur)
				w.Stop()
			}()
		}

		backupStdOut := os.Stdout

		reader, writer, _ := os.Pipe()
		os.Stdout = writer

		w.Run()

		writer.Close()
		os.Stdout = backupStdOut
		csvReader := csv.NewReader(reader)
		data, err := csvReader.ReadAll()
		// for _, line := range data {
		// 	fmt.Println(line)
		// }
		if err != nil {
			panic(err)
		}
		res := csvReport(data)
		return c.JSON(http.StatusOK, res)
	}
}

func csvReport(data [][]string) Result {
	res := Result{}
	res.Requests = len(data) - 1
	res.StatusCodes = make(map[string]int)
	var total float64
	latencies := make([]float64, 0, res.Requests)

	for i, line := range data {
		if i == 0 {
			continue
		}
		responseTime, _ := strconv.ParseFloat(line[0], 64)
		latencies = append(latencies, responseTime)
		total += responseTime
		res.StatusCodes[line[6]]++
	}

	res.P50, res.P90, res.P95, res.P99 = calculatePercentiles(latencies)
	return res
}

func calculatePercentiles(latencies []float64) (p50, p90, p95, p99 float64) {
	sort.Float64s(latencies)
	n := len(latencies)

	if n == 0 {
		return 0, 0, 0, 0
	}

	p50Index := (0.50 * float64(n))
	p90Index := (0.90 * float64(n))
	p95Index := (0.95 * float64(n))
	p99Index := (0.99 * float64(n))

	p50 = getValueByIndex(latencies, p50Index)
	p90 = getValueByIndex(latencies, p90Index)
	p95 = getValueByIndex(latencies, p95Index)
	p99 = getValueByIndex(latencies, p99Index)

	return
}

func getValueByIndex(latencies []float64, index float64) float64 {
	if int(index) >= len(latencies) {
		return latencies[len(latencies)-1]
	}
	return latencies[int(index)]
}

func main() {
	url := getEnvOrDefault("URL", "https://instance-raap3scyuq-uc.a.run.app/prime")
	request := getEnvOrDefaultInt("REQUEST", 100)
	concurrency := getEnvOrDefaultInt("CONCURRENCY", 10)
	duration := getEnvOrDefaultInt("DURATION", 15)
	timeout := getEnvOrDefaultInt("TIMEOUT", 10)

	cfg := HeyConfig{
		URL:         url,
		Request:     request,
		Concurrency: concurrency,
		Duration:    duration,
		Timeout:     timeout,
	}
	e := echo.New()
	e.Use(
		middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "method=${method}, uri=${uri}, latency=${latency_human}, status=${status}\n",
		}),
		middleware.Recover(),
		middleware.StaticWithConfig(middleware.StaticConfig{
			Root:  "public",
			HTML5: true,
		}),
	)
	e.POST("/generate", generate(cfg))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
