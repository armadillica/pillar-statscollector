package main

import (
	"flag"
	"time"

	"github.com/armadillica/pillar-statscollector/elastic"
	"github.com/armadillica/pillar-statscollector/pillar"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
)

var cliArgs struct {
	verbose  bool
	debug    bool
	MongoURL string
}

func parseCliArgs() {
	flag.BoolVar(&cliArgs.verbose, "verbose", false, "Enable info-level logging")
	flag.BoolVar(&cliArgs.debug, "debug", false, "Enable debug-level logging")
	flag.StringVar(&cliArgs.MongoURL, "mongo", "mongodb://localhost/cloud", "URL of the MongoDB database to connect to")
	flag.Parse()
}

func configLogging() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// Only log the warning severity or above by default.
	level := log.WarnLevel
	if cliArgs.debug {
		level = log.DebugLevel
	} else if cliArgs.verbose {
		level = log.InfoLevel
	}
	log.SetLevel(level)
}

func main() {
	parseCliArgs()
	configLogging()

	// Connect to MongoDB
	session, err := mgo.Dial(cliArgs.MongoURL)
	if err != nil {
		log.Panic(err)
	}
	session.SetMode(mgo.Monotonic, true) // Optional. Switch the session to a monotonic behavior.

	timestamp := time.Now().UTC() // TODO: take the time from the CLI?
	// timestamp := time.Time{}.UTC()
	stats, err := pillar.CollectStats(session, &timestamp)
	if err != nil {
		log.Fatalf("Error collecting statistics: %s", err)
	}

	if err := elastic.Push(stats); err != nil {
		log.Fatalf("Error pushing to ElasticSearch: %s", err)
	}
}
