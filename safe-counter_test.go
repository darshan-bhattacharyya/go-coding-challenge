package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	TEST_COUNTER_JSON_CREATE_AND_NOT_EXISTS = "anyjson.json"
	TEST_COUNTER_JSON_EXISTS                = "count_test.json"
)

func TestNewCounter(t *testing.T) {
	counter := NewCounter(10, TEST_COUNTER_JSON_CREATE_AND_NOT_EXISTS)

	if counter == nil {
		t.Fatal("New Counter should return not nil")
	}

	if counter.Value() > 0 {
		t.Error("Initial count should be 0")
	}

	counter = NewCounter(10, TEST_COUNTER_JSON_EXISTS)

	if counter == nil {
		t.Fatal("New Counter should return not nil")
	}
	if counter.Value() == 0 {
		t.Error("Value does not match counter file")
	}
	t.Log("Counter created successfully from file")
}

func TestConcurrentCallsToServer(t *testing.T) {
	server := createServer(TEST_COUNTER_JSON_CREATE_AND_NOT_EXISTS)
	countsChannel := make(chan int, 5)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Server is stopping: %v", err)
		}
	}()

	for i := 0; i < 5; i++ {
		go func() {
			response, err := http.Get("http://localhost:8000")
			if err != nil {
				log.Fatalf("error in calling the server is not expected: %v", err)
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				log.Fatalf("error in reading the body is not expected: %v", err)
			}
			counts := make(map[string]interface{})
			json.Unmarshal(body, &counts)
			countsChannel <- int(counts["Count"].(float64))
		}()
	}

	time.Sleep(time.Second * 1)
	setOfCounts := make(map[int]bool)
	for i := 0; i < 5; i++ {
		count := <-countsChannel
		if _, ok := setOfCounts[count]; ok {
			t.Errorf("got repeating value from count channel: %v", count)
		} else {
			setOfCounts[count] = true
		}
	}
	close(countsChannel)
	t.Log("Successfully made concurrent calls to server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error in Shutdown: %v", err)
	}
}

func TestWindowMoving(t *testing.T) {
	counter := NewCounter(10, TEST_COUNTER_JSON_CREATE_AND_NOT_EXISTS)

	if counter == nil {
		t.Fatal("New Counter should return not nil")
	}

	if counter.Value() > 0 {
		t.Error("Initial count should be 0")
	}

	count := 0

	for i := 0; i < 10; i++ {
		count = int(counter.Increment())
		time.Sleep(time.Second)
	}
	if count != 10 {
		t.Errorf("Count expected: 100, actual: %d", count)
	}
	time.Sleep(time.Second)
	count = int(counter.Increment())
	if count > 10 {
		t.Errorf("Count expected less than 10: actual: %d", count)
	}
	t.Log("Count reset to new window")
}

func TestPersistCounter(t *testing.T) {
	counter := NewCounter(10, TEST_COUNTER_JSON_CREATE_AND_NOT_EXISTS)

	if counter == nil {
		t.Fatal("New Counter should return not nil")
	}

	if counter.Value() > 0 {
		t.Error("Initial count should be 0")
	}

	count := 0

	for i := 0; i < 100; i++ {
		count = int(counter.Increment())
	}

	_, cancel := context.WithCancel(context.Background())
	onGracefulShutdown(counter, cancel)

	jsonBytes, err := os.ReadFile(TEST_COUNTER_JSON_CREATE_AND_NOT_EXISTS)
	jsonCounter := JsonCounter{}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = json.Unmarshal(jsonBytes, &jsonCounter)
	if err != nil {
		t.Fatalf("error unmarshaling counter file: %v", err)
	}

	if len(jsonCounter.RequestTimestamps) != count {
		t.Errorf("Expected count: %d, actual: %d", count, len(jsonCounter.RequestTimestamps))
	}
	err = os.Remove(TEST_COUNTER_JSON_CREATE_AND_NOT_EXISTS)
	if err != nil {
		t.Errorf("Error deleting the file: %v", err)
	}
}
