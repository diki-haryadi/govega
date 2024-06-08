# event
# Event Emitter

Supported driver
- Logger (sender & writer)
- Kafka (sender)
- MongoDB outbox (sender & writer)
- SQL outbox (sender & writer)

Support for hybrid mode, combination of sender and writer

Hybrid mode can be used to combine outbox pattern with direct publish more efficiently. For exmaple, with the combination of SQL writer and Kafka sender, when `Publish` is called, it will try to write the event in SQL database first, and then asynchronously trying to send the event to kafka topic after that. If the publishing is success, then the SQL writer will delete the record. In an exception case, when the sender is failed to send the event, another service suchs as [JobLst](https://gitlab.sicepat.tech/tools/joblst) could be used to query the database and send the event to kafka.

## Usage

API

```
Publish(ctx context.Context, event, key string, message interface{}, metadata map[string]interface{}) error
```


### Event Config

Event config can be used to map internal event name into actual topic name and to add or replace default metadata

```go
package main

import (
	"github.com/diki-haryadi/govega/event"
)

func main() {
	conf := &EmitterConfig{
		Sender: &DriverConfig{
			Type: "logger",
		},
		EventConfig : &EventConfig{
			EventMap : map[string]string{
				"test" : "kafka-test-topic", // map event test into topic kakfa-test-topic
			},
			Metadata : map[string]map[string]interface{} {
				"test" : { // add metdata on event test
					"schema": "test-schema",	
				},
			}
		}
	}

    ctx := context.Background()

	em, err := event.New(ctx, conf)
    em.Publish(ctx, "test", "t123", "testdata", nil)
}
```


### Example (Logger)

```go
package main

import (
	"github.com/diki-haryadi/govega/event"
)

func main() {
	conf := &EmitterConfig{
		Sender: &DriverConfig{
			Type: "logger",
		},
	}

    ctx := context.Background()

	em, err := event.New(ctx, conf)
    em.Publish(ctx, "test", "t123", "testdata", nil)
}
```

### Example (Kafka Sender)

```go
package main

import (
	"github.com/diki-haryadi/govega/event"
    _ "github.com/diki-haryadi/govega/event/kafka"
)

func main() {
	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type: "kafka",
			Config: map[string]interface{}{
				"brokers": []string{"localhost:9092"},
			},
		},
	}

    ctx := context.Background()

	em, err := event.New(ctx, conf)
    em.Publish(ctx, "test", "t123", "testdata", nil)
}
```

### Example (Hybrid Kafka + SQL)

```go
package main

import (
	"github.com/diki-haryadi/govega/database"
	"github.com/diki-haryadi/govega/event"
	"github.com/diki-haryadi/govega/event/sql"
    _ "github.com/diki-haryadi/govega/event/kafka"
    _ "github.com/diki-haryadi/govega/event/sql"
)

func main() {
	dbc := database.DBConfig{
		MasterDSN:     "root:password@(localhost:3306)/outbox?parseTime=true",
		SlaveDSN:      "root:password@(localhost:3306)/outbox?parseTime=true",
		RetryInterval: 10,
		MaxIdleConn:   5,
		MaxConn:       10,
	}

	db := database.New(dbc, database.DriverMySQL)

	conf := &event.EmitterConfig{
		Sender: &event.DriverConfig{
			Type: "kafka",
			Config: map[string]interface{}{
				"brokers": []string{"localhost:9092"},
			},
		},
        Writer: &event.DriverConfig{
			Type: "sql",
			Config: map[string]interface{}{
				"driver": "mysql",
				"table":  "outbox",
				"connection": db,
			},
		},
	}

    ctx := context.Background()

    //Start SQL transaction
    tx, err := db.Master.Begin()

    //Attach transaction into context
	ttx := context.WithValue(ctx, sql.NewSQLTxContext(sql.DefaultContextKey), tx)

	em, err := event.New(ctx, conf)
    em.Publish(ttx, "test", "t123", "testdata", nil)

    //Commit transaction
    tx.Commit()
}
```

# Event Consumer

Supported driver
- Logger
- Kafka

## Usage

API

```go
//Use add middlewares to actual event handler before accessing the actual handler
Use(middlewares ...EventMiddleware)
//Subscribe to a topic with specific group
Subscribe(ctx context.Context, topic, group string, handler EventHandler) error
//Start activate the consumer and start receiving event
Start() error
//Stop gracefully stop the consumer waiting for all workers to complete before exiting
Stop() error
```


### Event Config

Event config can be used to map internal event name into actual topic name and
group name

```go
package main

import (
	"github.com/diki-haryadi/govega/event"
	"github.com/diki-haryadi/govega/log"
)

func main() {
	conf := &ConsumerConfig{
		Listener: &DriverConfig{
			Type: "logger",
		},
		EventConfig : &EventConfig{
			EventMap : map[string]string{
				"test" : "kafka-test-topic", // map event test into topic kakfa-test-topic
			},
			GroupMap : map[string]string{
				"test" : "kafka-test-group", // map group test into group kakfa-test-group
			},
		}
	}

    ctx := context.Background()

	consumer, err := event.NewConsumer(ctx, conf)
	if err != nil {
		//handle error
	}

	consumer.Subscribe(context.Background(), "test", "test",
		func(ctx context.Context, msg *event.EventMessage) error {
			log.WithFields(log.Fields{
				"message": fmt.Sprintf("%+v", msg),
			}).Println("new message")
			return nil
		},
	)
}
```

### Worker Pool Config

Worker pool config can be used to set the number or workes for each topic group subscriber

The number of worker pool for each topic group subscriber will be decided with order:
- specific topic group from config
- topic default worker pool from config
- default worker pool from config
- default worker pool (1)

```go
package main

import (
	"github.com/diki-haryadi/govega/event"
	"github.com/diki-haryadi/govega/log"
)

func main() {
	conf := &ConsumerConfig{
		Listener: &DriverConfig{
			Type: "logger",
		},
		EventConfig : &EventConfig{
			EventMap : map[string]string{
				"test" : "kafka-test-topic", // map event test into topic kakfa-test-topic
			},
			GroupMap : map[string]string{
				"test" : "kafka-test-group", // map group test into group kakfa-test-group
			},
		},
		WorkerPoolConfig: &WorkerPoolConfig{
			"default": 2, //default worker pool for each subscriber
			"kafka-test-topic": {
				"default": 1, // default worker pool for kafka-test-topic Topic subscriber
				"kafka-test-group": 3 // default worker pool for kafka-test-topic Topic with kafka-test-group Group
			}
		}
	}

    ctx := context.Background()
	consumer, err := event.NewConsumer(ctx, conf)
	if err != nil {
		//handle error
	}

	consumer.Subscribe(context.Background(), "test", "test",
		func(ctx context.Context, msg *event.EventMessage) error {
			log.WithFields(log.Fields{
				"message": fmt.Sprintf("%+v", msg),
			}).Println("new message")
			return nil
		},
	)
}
```

### Example (Kafka)

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/diki-haryadi/govega/event"
	_ "github.com/diki-haryadi/govega/event/kafka"
	"github.com/diki-haryadi/govega/log"
)

func main() {
	conf := &event.ConsumerConfig{
		Listener: &event.DriverConfig{
			Type: "kafka",
			Config: map[string]interface{}{
				"brokers": []string{"localhost:9092"},
			},
		},
		EventConfig: &event.EventConfig{
			EventMap: map[string]string{
				"test": "kafka-test-topic", // map event test into topic kakfa-test-topic
			},
			GroupMap: map[string]string{
				"test": "kafka-test-group", // map group test into group kakfa-test-group
			},
		},
		WorkerPoolConfig: &event.WorkerPoolConfig{
			"default": 2, //default worker pool for each subscriber
			"kafka-test-topic": map[string]interface{}{
				"default":          1, // default worker pool for kafka-test-topic Topic subscriber
				"kafka-test-group": 3, // default worker pool for kafka-test-topic Topic with kafka-test-group Group
			},
		},
		ConsumeStrategy: &event.DriverConfig{
			Type:   "commit_on_success", // available strategy commit_on_success and always_commit, default: commit_on_success
			Config: map[string]interface{}{},
		},
	}

	ctx := context.Background()
	consumer, err := event.NewConsumer(ctx, conf)
	if err != nil {
		log.WithError(err).Fatalln("failed to configure consumer")
	}

	err = consumer.Subscribe(context.Background(), "test", "test",
		func(_ context.Context, msg *event.EventConsumeMessage) error {
			log.WithFields(log.Fields{
				"message": fmt.Sprintf("%+v", msg),
			}).Println("new message")
			return nil
		},
	)
	if err != nil {
		log.WithError(err).Panicln("failed to subscribe topic")
	}

	if err := consumer.Start(); err != nil {
		log.WithError(err).Fatalln("failed to start consumer")
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	<-signalCh

	if err := consumer.Stop(); err != nil {
		log.WithError(err).Errorln("error on stopping consumer")
	}

	log.Println("bye 👋")
}
```
