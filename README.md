# Problem Statement
Create a Go HTTP server which returns total number of request made to the server in previous 60 seconds window. Also the server should persist the counts in order to start where it left off while being restarted.


## Run Locally

Clone the project

```bash
  git clone REPO
```

Go to the project directory

```bash
  cd MYPROJECT
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
```
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

To keep the count values and window position persistent I am saving this values in a json file while graceful shutdown of the system. Following format is used.

```
{
    "count": 1,
    "windowStart": 1699537335123,
    "windowEnd": 1699537395123,
    "windowLength": 10
}
```

Graceful shutdown is handled by listening for SIGINT and SIGTERM.

