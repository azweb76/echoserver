package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/pbnjay/memory"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func ByteCountIEC(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func PrintResourceUsage(w http.ResponseWriter) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Fprint(w, "<h1>Resource Utilization</h1><pre>")
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Fprintf(w, "HeapAlloc  => %s<br/>", ByteCountIEC(m.HeapAlloc))
	fmt.Fprintf(w, "HeapInuse  => %s<br/>", ByteCountIEC(m.HeapInuse))
	fmt.Fprintf(w, "Alloc      => %s<br/>", ByteCountIEC(m.Alloc))
	fmt.Fprintf(w, "TotalAlloc => %s<br/>", ByteCountIEC(m.TotalAlloc))
	fmt.Fprintf(w, "Sys        => %s<br/>", ByteCountIEC(m.Sys))
	fmt.Fprintf(w, "Total      => %s<br/>", ByteCountIEC(memory.TotalMemory()))
	fmt.Fprintf(w, "Free       => %s<br/>", ByteCountIEC(memory.FreeMemory()))
	fmt.Fprintf(w, "NumGC      => %v<br/>", m.NumGC)
	fmt.Fprint(w, "</pre>")
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func calculatePrimeNumbers(num1, num2 int) int {
	primeCount := 0
	for num1 <= num2 {
		isPrime := true
		for i := 2; i <= int(math.Sqrt(float64(num1))); i++ {
			if num1%i == 0 {
				isPrime = false
				break
			}
		}
		if isPrime {
			primeCount += primeCount
		}
		num1++
	}
	return primeCount
}

func generateLoad(timeSeconds int, percentage int, factors int) {
	done := make(chan int)
	var unitHundredsOfMicrosecond int = 1000
	runMicrosecond := unitHundredsOfMicrosecond * (100 - percentage)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for {
				for {
					time.Sleep(time.Nanosecond * time.Duration(runMicrosecond))
					calculatePrimeNumbers(1, factors)
					select {
					case <-done:
						return
					default:
					}
				}
			}
		}()
	}
	time.Sleep(time.Second * time.Duration(timeSeconds))
	close(done)
}

func crashHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Crash initiated\n")
	os.Exit(1)
}

func handler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "<head><title>EchoServer</title></head><body>")
	fmt.Fprintf(w, "<h1>OS</h1><pre>")
	if hostname, err := os.Hostname(); err == nil {
		fmt.Fprintf(w, "Hostname: %s", hostname)
	}
	fmt.Fprintf(w, "</pre>")
	fmt.Fprintf(w, "<h1>Headers</h1><pre>")
	header := r.Header
	for key, element := range header {
		fmt.Fprintf(w, "%s=>%s<br/>", key, element)
	}
	fmt.Fprintf(w, "</pre>")
	fmt.Fprintf(w, "<h1>Query</h1><pre>")
	query := r.URL.Query()
	for key, element := range query {
		fmt.Fprintf(w, "%s=>%s<br/>", key, element)
	}
	fmt.Fprintf(w, "</pre>")
	PrintResourceUsage(w)
	if dataLength, err := strconv.ParseInt(query.Get("data"), 10, 32); err == nil {
		if dataLength > 0 {
			fmt.Fprintf(w, "<h1>Sample Data (length=%d)</h1><pre>", dataLength)
			fmt.Fprint(w, RandStringRunes(int(dataLength)))
			fmt.Fprintf(w, "</pre>")
		}
	}
	if loadSeconds, err := strconv.ParseInt(query.Get("loadSeconds"), 10, 32); err == nil {
		loadPct, err := strconv.ParseInt(query.Get("loadPct"), 10, 32)
		if err != nil {
			loadPct = 100
		}
		loadFactors, err := strconv.ParseInt(query.Get("loadFactors"), 10, 32)
		if err != nil {
			loadFactors = 10000
		}
		generateLoad(int(loadSeconds), int(loadPct), int(loadFactors))
	}
	fmt.Fprintf(w, "</body>")
}

func getEnv(name string, defaultValue string) string {
	value, exists := os.LookupEnv("ECHOSERVER_" + name)
	if exists {
		return value
	}
	return defaultValue
}

func run() {
	crashOnStart, _ := strconv.ParseBool(getEnv("CRASH", "false"))
	if crashOnStart {
		log.Panic("CRASH env detected. Terminating.")
	}
	http.HandleFunc("/", handler)
	http.HandleFunc("/crash", crashHandler)
	log.Print("Starting on 0.0.0.0:8080")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Panic(err)
	}
}

func healthcheck() {
	requestURL := "http://localhost:8080"
	_, err := http.Get(requestURL)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "run" {
		run()
	} else {
		healthcheck()
	}
}
