package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"time"
)

type Job struct {
	Data []int64 `json:"data"`
}

type Result struct {
	Count int     `json:"count"`
	Sum   int64   `json:"sum"`
	SumSq float64 `json:"sum_sq"`
}

func compute(data []int64) Result {
	var r Result
	for _, n := range data {
		r.Count++
		r.Sum += n
		r.SumSq += float64(n) * float64(n)
	}
	return r
}

func main() {
	worker := flag.String("worker", "192.168.56.102:9000", "Worker address")
	n := flag.Int("n", 100000, "Number of integers")
	flag.Parse()

	if *n < 2 {
		log.Fatal("n must be at least 2")
	}

	data := make([]int64, *n)
	for i := range data {
		data[i] = int64(i + 1)
	}

	startSeq := time.Now()
	sequential := compute(data)
	seqTime := time.Since(startSeq)

	mid := len(data) / 2
	localData, remoteData := data[:mid], data[mid:]

	startDist := time.Now()

	localResult := compute(localData)

	conn, err := net.DialTimeout("tcp", *worker, 5*time.Second)
	if err != nil {
		log.Fatal("Connect to worker: ", err)
	}

	if err := json.NewEncoder(conn).Encode(Job{Data: remoteData}); err != nil {
		conn.Close()
		log.Fatal("Send job: ", err)
	}

	var remoteResult Result
	if err := json.NewDecoder(conn).Decode(&remoteResult); err != nil {
		conn.Close()
		log.Fatal("Receive result: ", err)
	}
	conn.Close()

	distributedTime := time.Since(startDist)
	final := Result{
		Count: localResult.Count + remoteResult.Count,
		Sum:   localResult.Sum + remoteResult.Sum,
		SumSq: localResult.SumSq + remoteResult.SumSq,
	}

	speedup := float64(seqTime) / float64(distributedTime)
	efficiency := speedup / 2 * 100

	fmt.Println("\n=== Distributed Memory Go ===")
	fmt.Printf("Total data: %d\n", len(data))
	fmt.Printf("Master processed: %d\n", localResult.Count)
	fmt.Printf("Worker processed: %d\n", remoteResult.Count)
	fmt.Printf("Sequential sum: %d\n", sequential.Sum)
	fmt.Printf("Distributed sum: %d\n", final.Sum)
	fmt.Printf("Sequential sum of squares: %.0f\n", sequential.SumSq)
	fmt.Printf("Distributed sum of squares: %.0f\n", final.SumSq)
	fmt.Printf("Results match: %v\n",
		sequential.Count == final.Count &&
			sequential.Sum == final.Sum &&
			math.Abs(sequential.SumSq-final.SumSq) < 0.5)
	fmt.Printf("Sequential time: %v\n", seqTime)
	fmt.Printf("Distributed time: %v\n", distributedTime)
	fmt.Printf("Speedup: %.4fx\n", speedup)
	fmt.Printf("Efficiency: %.2f%%\n", efficiency)
}