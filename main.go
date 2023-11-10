package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const COUNTER_HEADER = "request-count"

const ERROR_TOO_MANY_REQUEST = "too many requests in the server"

var (
	counter *Counter
)

func createServer(counterFileName string) *http.Server {
	counter = NewCounter(60, counterFileName)
	muxer := http.NewServeMux()
	muxer.HandleFunc("/", Home)
	server := &http.Server{
		Addr:    ":8000",
		Handler: CounterMiddleware(muxer),
	}
	return server
}

func onGracefulShutdown(counter *Counter, cancel context.CancelFunc) {
	defer cancel()
	err := counter.SaveToJSON()
	if err != nil {
		log.Fatalf("Error while saving counts: %v", err)
	}
}

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	server := createServer("count.json")
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Server is stopping: %v", err)
		}
	}()
	log.Printf("Server started...")

	<-done
	log.Printf("Received signal to Shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer onGracefulShutdown(counter, cancel)

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error in Shutdown: %v", err)
	}
	log.Print("Gracefully shutting down")
}

func CounterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := counter.Increment()

		r.Header.Add(COUNTER_HEADER, fmt.Sprintf("%d", count))
		next.ServeHTTP(w, r)

	})
}

func Home(w http.ResponseWriter, r *http.Request) {
	count, err := strconv.Atoi(r.Header.Get(COUNTER_HEADER))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error: convert count"))
	}
	response := map[string]interface{}{
		"Message": "status is ok",
		"Count":   count,
	}
	respBytes, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error: Cannot marshal json"))
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}
