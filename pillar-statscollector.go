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
	verbose    bool
	debug      bool
	mongoURL   string
	elasticURL string
	before     string
	nopush     bool
}

func parseCliArgs() {
	flag.BoolVar(&cliArgs.verbose, "verbose", false, "Enable info-level logging")
	flag.BoolVar(&cliArgs.debug, "debug", false, "Enable debug-level logging")
	flag.BoolVar(&cliArgs.nopush, "nopush", false, "Log statistics, but don't push to ElasticSearch")
	flag.StringVar(&cliArgs.mongoURL, "mongo", "mongodb://localhost/cloud", "URL of the MongoDB database to connect to")
	flag.StringVar(&cliArgs.elasticURL, "elastic", "http://localhost:9200/", "URL of the ElasticSearch instance to push to")
	flag.StringVar(&cliArgs.before, "before", "", "Only consider objects created before this timestamp, expected in RFC 3339 format")
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
	session, err := mgo.Dial(cliArgs.mongoURL)
	if err != nil {
		log.Panic(err)
	}
	session.SetMode(mgo.Monotonic, true) // Optional. Switch the session to a monotonic behavior.

	var timestamp *time.Time
	if cliArgs.before != "" {
		parsed, parseErr := time.Parse(time.RFC3339, cliArgs.before)
		if parseErr != nil {
			log.Fatalf("Invalid argument -before %q: %s", cliArgs.before, parseErr)
		}
		timestamp = &parsed
	}
	stats, err := pillar.CollectStats(session, timestamp)
	if err != nil {
		log.Fatalf("Error collecting statistics: %s", err)
	}

	if cliArgs.nopush {
		log.Warning("Not pushing to ElasticSearch")
		return
	}
	if err := elastic.Push(cliArgs.elasticURL, stats); err != nil {
		log.Fatalf("Error pushing to ElasticSearch: %s", err)
	}
}
