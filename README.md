# logscissors

This is a golang log rotator implementing the io.Writer interface and it's goroutine safe.

Import it in your program as:
```go
      import "github.com/CormickKneey/logscissors"
```
## API
### pakcage func
	func NewLogScissors(pattern string, period time.Duration) (*LogScissors, error)
    func NewLogScissorsWithPreFilename(pattern string, period time.Duration, preFilename string) (*LogScissors, error)
	func NewLogCleaner(pattern string, maxAge time.Duration) (*LogCleaner, error)
### type LogScissors
	func (tw *LogScissors) Write(p []byte) (n int, err error)
	func (tw *LogScissors) Close() error
### type LogCleaner
	func (cleaner *LogCleaner) Clean() ([]string, error) 



## Write Example
Rotate a log at a specified time.Duration or move the content to other day .
```golang
package main

import (
	"fmt"
	clog "github.com/CormickKneey/logscissors"
	"io"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	// cut log by duration, current log file will be timestamp.log 
	logFileByDate, err := clog.NewLogScissors("/tmp/test-%Y%m%d-%H%M.log", 1*time.Hour)
	if err != nil {
		fmt.Printf("config local file system logger error. %v\n", err)
	}

	// save log by duration, current log file will be the name set by "preFilename"
	logFile, err := clog.NewLogScissorsWithPreFilename("/tmp/test-%Y%m%d-%H%M.log", 1*time.Hour(),"/tmp/test.log")
	if err != nil {
		fmt.Printf("config local file system logger error. %v\n", err)
	}
	defer logFileByDate.Close()
	defer logFile.Close()
}


## Clean Example: 
Clean all old log files that have not being modified for at the least 7 days. And the job is scheduled at 1:05 in every morning.
```go
package main

import (
	"fmt"
	"github.com/pochard/logrotator"
	"github.com/robfig/cron/v3"
	"net/http"
	"time"
)

func main() {
	cleaner, err := logrotator.NewLogCleaner("/data/web_log/*.log", 7 * 24 * time.Hour)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	c := cron.New()
	c.AddFunc("5 1 * * *", func() {
		deleted, err := cleaner.Clean()
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		for _, d := range deleted {
			fmt.Printf("%s deleted\n", d)
		}
	})
	c.Start()

	http.ListenAndServe(":8080", nil)
}

```
