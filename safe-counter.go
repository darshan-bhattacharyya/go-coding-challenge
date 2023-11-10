package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type JsonCounter struct {
	Count        int64 `json:"count"`
	WindowStart  int64 `json:"windowStart"`
	WindowEnd    int64 `json:"windowEnd"`
	WindowLength int64 `json:"windowLength"`
}

type Counter struct {
	mu           sync.RWMutex
	jsonFile     string
	count        int64
	windowStart  time.Time
	windowEnd    time.Time
	windowLength time.Duration
}

func NewCounter(windowLength int, counterFile string) *Counter {
	counter := &Counter{jsonFile: counterFile}
	jsonBytes, err := os.ReadFile(counterFile)
	jsonCounter := JsonCounter{}
	if err != nil {
		log.Printf("error reading counter file: %v", err)
		counter.Start(windowLength)
		return counter
	}
	err = json.Unmarshal(jsonBytes, &jsonCounter)
	if err != nil {
		log.Printf("error unmarshaling counter file: %v", err)
	}
	counter = &Counter{
		jsonFile:     counterFile,
		count:        jsonCounter.Count,
		windowStart:  time.UnixMilli(jsonCounter.WindowStart),
		windowEnd:    time.UnixMilli(jsonCounter.WindowEnd),
		windowLength: time.Duration(jsonCounter.WindowLength),
	}

	return counter
}

func (c *Counter) Start(windowLength int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.windowLength = time.Duration(windowLength)
	c.windowStart = time.Now()
	c.windowEnd = c.windowStart.Add(time.Second * c.windowLength)
	c.count = 0
}

func (c *Counter) Increment() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	currentTime := time.Now()
	if currentTime.Before(c.windowEnd) {
		c.count += 1
		return c.count
	}
	c.count = 1
	c.windowStart = currentTime
	c.windowEnd = currentTime.Add(time.Second * c.windowLength)
	return c.count
}

func (c *Counter) Value() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}

func (c *Counter) SaveToJSON() error {
	file, err := os.Create(c.jsonFile)
	if err != nil {
		return err
	}

	jsonCounter := JsonCounter{
		Count:        c.count,
		WindowStart:  c.windowStart.UnixMilli(),
		WindowEnd:    c.windowEnd.UnixMilli(),
		WindowLength: int64(c.windowLength),
	}

	log.Print("Created json file")

	jsonBytes, err := json.Marshal(jsonCounter)
	if err != nil {
		return err
	}
	log.Printf("JsonBytes: %v", jsonBytes)
	file.Write(jsonBytes)
	file.Close()
	return nil
}
