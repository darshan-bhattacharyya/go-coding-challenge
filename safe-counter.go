package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type JsonCounter struct {
	RequestTimestamps []int64 `json:"requestTimestamps"`
	WindowStart       int64   `json:"windowStart"`
	WindowEnd         int64   `json:"windowEnd"`
	WindowLength      int64   `json:"windowLength"`
}

type Counter struct {
	mu                sync.RWMutex
	jsonFile          string
	requestTimestamps []time.Time
	windowStart       time.Time
	windowEnd         time.Time
	windowLength      time.Duration
}

func NewCounter(windowLength int, counterFile string) *Counter {
	counter := &Counter{jsonFile: counterFile, requestTimestamps: make([]time.Time, 0, 1000)}
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
		windowStart:  time.UnixMilli(jsonCounter.WindowStart),
		windowEnd:    time.UnixMilli(jsonCounter.WindowEnd),
		windowLength: time.Duration(jsonCounter.WindowLength),
	}
	for _, unixTimestamp := range jsonCounter.RequestTimestamps {
		counter.requestTimestamps = append(counter.requestTimestamps, time.UnixMilli(unixTimestamp))
	}

	return counter
}

func (c *Counter) Start(windowLength int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.windowLength = time.Duration(windowLength)
}

func (c *Counter) Increment() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.windowEnd = time.Now()
	c.windowStart = c.windowEnd.Add(-(time.Second * c.windowLength))
	i := 0
	c.requestTimestamps = append(c.requestTimestamps, c.windowEnd)
	for _, timestamp := range c.requestTimestamps {
		if timestamp.After(c.windowStart) {
			c.requestTimestamps[i] = timestamp
			i += 1
		}
	}
	c.requestTimestamps = c.requestTimestamps[0:i]
	return len(c.requestTimestamps)
}

func (c *Counter) Value() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.requestTimestamps)
}

func (c *Counter) SaveToJSON() error {
	file, err := os.Create(c.jsonFile)
	if err != nil {
		return err
	}

	jsonCounter := JsonCounter{
		WindowStart:  c.windowStart.UnixMilli(),
		WindowEnd:    c.windowEnd.UnixMilli(),
		WindowLength: int64(c.windowLength),
	}

	for _, timestamp := range c.requestTimestamps {
		jsonCounter.RequestTimestamps = append(jsonCounter.RequestTimestamps, timestamp.UnixMilli())
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
