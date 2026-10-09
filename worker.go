package main

import (
	"encoding/json"
	"log"
	"net"
)

type Job struct {
	Data []int64 `json:"data"`
}

type Result struct {
	Count int     `json:"count"`
	Sum   int64   `json:"sum"`
	SumSq float64 `json:"sum_sq"`
}

func main() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("Worker listening on port 9000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()

	var job Job
	if err := json.NewDecoder(conn).Decode(&job); err != nil {
		log.Println("Decode:", err)
		return
	}

	var result Result
	for _, n := range job.Data {
		result.Count++
		result.Sum += n
		result.SumSq += float64(n) * float64(n)
	}

	if err := json.NewEncoder(conn).Encode(result); err != nil {
		log.Println("Encode:", err)
		return
	}

	log.Printf("Processed %d numbers\n", result.Count)
}