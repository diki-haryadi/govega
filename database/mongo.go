package database

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	otelmongo "go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

var mongoClients = make(map[string]*mongo.Client)

type (
	Client struct {
		URI     string `json:"uri" mapstructure:"uri"`
		DB      string `json:"database" mapstructure:"database"`
		AppName string `json:"name" mapstructure:"name"`
		// on second
		ConnectTimeout time.Duration
		// on second
		PingTimeout time.Duration
	}
	MongoDatabase struct {
		MongoDatabase *mongo.Database
	}
)

var (
	defaultConnectTimeout = 10 * time.Second
	defaultPingTimeout    = 2 * time.Second
)

func MongoConnectClient(c *Client) *MongoDatabase {
	client, err := c.MongoConnect()
	if err != nil {
		panic(err)
	}
	return &MongoDatabase{
		MongoDatabase: client.Database(c.DB),
	}
}

func (c *Client) MongoConnect() (mc *mongo.Client, err error) {

	if cl, ok := mongoClients[c.URI]; ok {
		return cl, nil
	}

	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = defaultConnectTimeout
	}

	if c.PingTimeout == 0 {
		c.PingTimeout = defaultPingTimeout
	}

	connectCtx, cancelConnectCtx := context.WithTimeout(context.Background(), c.ConnectTimeout)
	defer cancelConnectCtx()

	otelMon := otelmongo.NewMonitor()
	opts := []*options.ClientOptions{
		options.Client().SetConnectTimeout(c.ConnectTimeout).ApplyURI(c.URI).SetAppName(c.AppName),
		options.Client().SetMonitor(otelMon),
	}

	mc, err = mongo.Connect(connectCtx, opts...)
	if err != nil {
		err = errors.Wrap(err, "failed to create mongodb client")
		return
	}

	pingCtx, cancelPingCtx := context.WithTimeout(context.Background(), c.PingTimeout)
	defer cancelPingCtx()

	if err = mc.Ping(pingCtx, readpref.Primary()); err != nil {
		err = errors.Wrap(err, "failed to establish connection to mongodb server")
	}

	mongoClients[c.URI] = mc
	return
}

func GetMongoClient(url string) *mongo.Client {
	if url == "" {
		url = os.Getenv("MONGO_SERVER_URL")
	}

	if url == "" {
		return nil
	}

	c, ok := mongoClients[url]
	if ok {
		return c
	}

	name := os.Getenv("NAME")
	if name == "" {
		name = "Default"
	}

	cfg := &Client{URI: url, AppName: name}
	client, err := cfg.MongoConnect()
	if err != nil {
		return nil
	}

	return client
}
