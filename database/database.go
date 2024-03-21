package database

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	log "github.com/sirupsen/logrus"
)

const (
	DriverMySQL     = "mysql"
	DriverPostgress = "postgres"
)

type (
	Store struct {
		Master *sqlx.DB
		Slave  *sqlx.DB
	}

	Database struct {
		Conn          *sqlx.DB
		Dsn           string
		RetryInterval int
		MaxIdleConn   int
		MaxConn       int
		doneChan      chan bool
	}

	DatabaseConfig struct {
		MasterDSN     string `json:"master_dsn" mapstructure:"master_dsn"`
		SlaveDSN      string `json:"slave_dsn" mapstructure:"slave_dsn"`
		RetryInterval int    `json:"retry_interval" mapstructure:"retry_interval"`
		MaxIdleConn   int    `json:"max_idle" mapstructure:"max_idle"`
		MaxConn       int    `json:"max_con" mapstructure:"max_con"`
	}
)

func (s *Store) GetMaster() *sqlx.DB {
	return s.Master
}

func (s *Store) GetSlave() *sqlx.DB {
	return s.Slave
}

func New(dbCfg DatabaseConfig, dbDriver string) *Store {

	mstr := &Database{}
	mstr.Dsn = dbCfg.MasterDSN
	mstr.RetryInterval = dbCfg.RetryInterval
	mstr.MaxIdleConn = dbCfg.MaxIdleConn
	mstr.MaxConn = dbCfg.MaxConn
	mstr.doneChan = make(chan bool)

	if err := mstr.ConnectAndMonitor(dbDriver); err != nil {
		log.Fatal("ERROR]: Could not initiate Master DB connection: " + err.Error())
		return nil
	}

	slv := &Database{}
	slv.Dsn = dbCfg.MasterDSN
	slv.RetryInterval = dbCfg.RetryInterval
	slv.MaxIdleConn = dbCfg.MaxIdleConn
	slv.MaxConn = dbCfg.MaxConn
	slv.doneChan = make(chan bool)

	if err := slv.ConnectAndMonitor(dbDriver); err != nil {
		log.Fatal("ERROR]: Could not initiate Master DB connection: " + err.Error())
		return nil
	}

	return &Store{Master: mstr.Conn, Slave: slv.Conn}
}

func (d *Database) ConnectAndMonitor(driver string) error {

	if err := d.Connect(driver); err != nil {
		log.Printf("[ERROR]: database connection error: %s\n", err.Error())
		return err
	}

	ticker := time.NewTicker(
		time.Duration(d.RetryInterval) * time.Second)

	go func() error {
		for {
			select {
			case <-ticker.C:
				switch d.Conn {
				case nil:
					d.Connect(driver)
				default:
					if err := d.Conn.Ping(); err != nil {
						log.Printf("[ERROR]: database reconnection error: %s\n", err.Error())
						return err
					}
				}
			case <-d.doneChan:
				return nil
			}
		}
	}()

	return nil
}

func (d *Database) Connect(driver string) error {

	db, err := sqlx.Open(driver, d.Dsn)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	d.Conn = db

	db.SetMaxOpenConns(d.MaxConn)
	db.SetMaxIdleConns(d.MaxIdleConn)

	return nil
}
