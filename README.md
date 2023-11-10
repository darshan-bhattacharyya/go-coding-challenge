# Problem Statement
Create a Go HTTP server which returns total number of request made to the server in previous 60 seconds window. Also the server should persist the counts in order to start where it left off while being restarted.


## Run Locally

Clone the project

```bash
  git clone https://github.com/darshan-bhattacharyya/go-coding-challenge.git
```

Go to the project directory

```bash
  cd go-coding-challenge
```

Build the executable

```bash
  go build
```

Start the server

On Mac or Linux
```bash
  ./go-coding-challenge
```

Stop the server using Control + C on Mac or Ctrl + C on Windows

## Solution Overview

I used a mutex based counter approach to keep track of the counts that are reaching the server. For this approach I created the following struct -
```go
type Counter struct {
	mu           sync.RWMutex
	jsonFile     string
	count        int64
	windowStart  time.Time
	windowEnd    time.Time
	windowLength time.Duration
}
``` 
Here I am keeping track of the window start, end, length and as well as count. While each Increament operation we check whether the current time falls within this window. If not we reset the window and count.

```go
func NewCounter(windowLength int, counterFile string) *Counter )
```
The above constructor is provided to create a new instance of the counter (Pass window length in seconds).

Counting is done on the middleware level and obtained value of the count is forwarded to the handler function by using request header value.

To keep the count values and window position persistent I am saving this values in a json file while graceful shutdown of the system. Following format is used.

```json
{
    "count": 1,
    "windowStart": 1699537335123,
    "windowEnd": 1699537395123,
    "windowLength": 10
}
```

Graceful shutdown is handled by listening for SIGINT and SIGTERM on a signal channel.

## Unit Test Cases

Run unit tests with following command.
```bash
  go test -cover
```

Following unit test cases are used to validate the system.

```go
func TestConcurrentCallsToServer(t *testing.T)
```
To test the concurrency capabilities of the system by throwing concurrent request to the system and checking if the count values are unique each time.

```go
func TestWindowReset(t *testing.T)
```
To check if the window is getting reset after said amount of time.

```go
func TestPersistCounter(t *testing.T)
```
To check data persistent capability.

```go
func TestNewCounter(t *testing.T) 
```
To test the constructor.