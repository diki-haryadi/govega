# httpprofiling
HTTP profiling package 

## Quick Start

Installation
    $ go get github.com/diki-haryadi/govega/profiling/httpprofiling


## Usage

Options:
| Option           | Description                                                                                                                                         |
|------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| WithReadTimeout  | Timeout for the server to read the request data. <br><br>Default: 5m                                                                                |
| WithWriteTimeout | Timeout for the server to write response.<br><br>Default: 5m                                                                                        |
| WithPort         | The port to start the http profiler.<br><br>Default: 8432                                                                                           |
| WithRouter       | Use custom router for http profiler                                                                                                                 |
| WithManualStart  | If set to true, http profiler will not start right away on init.<br>The caller will need to manually start the profiler by calling `Start` function |

Example:

Simple init http profiler (unless there is specific usecase, this should be enough most of the time)

```go
package main

import (
	"context"

	"github.com/diki-haryadi/govega/profiling/httpprofiling"
)

func main() {
	if _, err := httpprofiling.InitProfiler(); err != nil {
		panic(err)
	}

	// do stuff
}
```

If you need to gracefully shutdown the http profiler

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/diki-haryadi/govega/log"
	"github.com/diki-haryadi/govega/profiling/httpprofiling"
)

func main() {
	profiler, err := httpprofiling.InitProfiler()
	if err != nil {
		panic(err)
	}

	// do stuff

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	<- ch

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := profiler.Stop(ctx); err != nil {
		log.WithError(err).Errorln("failed on stoping profiler")
	}
}
```

Manually start the http profiler

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/diki-haryadi/govega/log"
	"github.com/diki-haryadi/govega/profiling/httpprofiling"
)

func main() {
	profiler, err := httpprofiling.InitProfiler(httpprofiling.WithManualStart(true))
	if err != nil {
		panic(err)
	}

	// do stuff

	if err := profiler.Start(); err != nil {
		panic(err)
	}

	// do another stuff

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	<- ch

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// if needed to gracefully stop the http profiler
	if err := profiler.Stop(ctx); err != nil {
		log.WithError(err).Errorln("failed on stoping profiler")
	}
}
```
