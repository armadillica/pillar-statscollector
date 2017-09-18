package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/armadillica/pillar-statscollector/elastic"
	"github.com/armadillica/pillar-statscollector/grafista"
	"github.com/armadillica/pillar-statscollector/pillar"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
)

const statscollectorVersion = "1.0"

var cliArgs struct {
	version        bool
	verbose        bool
	debug          bool
	mongoURL       string
	elasticURL     string
	before         string
	nopush         bool
	allSince       string
	importGrafista string
}

func parseCliArgs() {
	flag.BoolVar(&cliArgs.version, "version", false, "Shows the application version, then exits.")
	flag.BoolVar(&cliArgs.verbose, "verbose", false, "Enable info-level logging.")
	flag.BoolVar(&cliArgs.debug, "debug", false, "Enable debug-level logging.")
	flag.BoolVar(&cliArgs.nopush, "nopush", false, "Log statistics, but don't push to ElasticSearch.")
	flag.StringVar(&cliArgs.mongoURL, "mongo", "mongodb://localhost/cloud", "URL of the MongoDB database to connect to.")
	flag.StringVar(&cliArgs.elasticURL, "elastic", "http://localhost:9200/cloudstats/stats/", "URL of the ElasticSearch instance to push to.")
	flag.StringVar(&cliArgs.before, "before", "", "Only consider objects created before this timestamp; expected in RFC 3339 format.")
	flag.StringVar(&cliArgs.allSince, "allsince", "", "Collect daily statistics since this timestamp until now; expected in RFC 3339 format.")
	flag.StringVar(&cliArgs.importGrafista, "import", "", "Imports data from a Grafista SQLite database and pushes stats to ElasticSearch.")
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

func collectAllSince(session *mgo.Session, beginTimestamp time.Time) error {
	log.Warningf("Collecting daily statistics since %s, this may take a while", beginTimestamp)
	now := time.Now().UTC()
	stepSize := 24 * time.Hour
	before := beginTimestamp
	pushCount := 0

	for {
		if before.After(now) {
			break
		}

		before = before.Add(stepSize).Round(24 * time.Hour)
		err := singleRun(session, &before)
		if err != nil {
			return fmt.Errorf("running with before=%s: %s", before, err)
		}
		pushCount++
	}

	log.Warnf("Done, pushed %d statistics documents", pushCount)
	return nil
}

func singleRun(session *mgo.Session, timestamp *time.Time) error {
	stats, err := pillar.CollectStats(session, timestamp)
	if err != nil {
		return fmt.Errorf("error collecting statistics: %s", err)
	}

	return pushStats(stats)
}

func importFromGrafista(dbFilename string) error {
	err := grafista.ImportDB(cliArgs.importGrafista, pushStats)
	if err != nil {
		return fmt.Errorf("error importing from Grafista DB: %s", err)
	}
	return nil
}

func pushStats(stats interface{}) error {
	if cliArgs.nopush {
		// Marshal the stats to JSON and log.
		asJSON, err := json.MarshalIndent(&stats, "", "    ")
		if err != nil {
			return fmt.Errorf("unable to marshal to JSON: %s", err)
		}
		log.Infof("Statistics:\n%s\n", string(asJSON))
		log.Warning("Not pushing to ElasticSearch")
		return nil
	}

	if err := elastic.Push(cliArgs.elasticURL, stats); err != nil {
		return fmt.Errorf("error pushing to ElasticSearch: %s", err)
	}
	return nil
}

func main() {
	parseCliArgs()
	if cliArgs.version {
		fmt.Println(statscollectorVersion)
		return
	}

	configLogging()

	if cliArgs.importGrafista != "" {
		err := importFromGrafista(cliArgs.importGrafista)
		if err != nil {
			log.Error(err)
		}
		return
	}

	// Connect to MongoDB
	session, err := mgo.Dial(cliArgs.mongoURL)
	if err != nil {
		log.Panic(err)
	}
	session.SetMode(mgo.Monotonic, true) // Optional. Switch the session to a monotonic behavior.

	if cliArgs.allSince != "" {
		if cliArgs.before != "" {
			log.Fatalf("Use either -before or -allsince, not both.")
		}

		beginTimestamp, parseErr := time.Parse(time.RFC3339, cliArgs.allSince)
		if parseErr != nil {
			log.Fatalf("Invalid argument -allsince %q: %s", cliArgs.allSince, parseErr)
		}

		err = collectAllSince(session, beginTimestamp)
	} else {
		if cliArgs.before == "" {
			err = singleRun(session, nil)
		} else {
			parsed, parseErr := time.Parse(time.RFC3339, cliArgs.before)
			if parseErr != nil {
				log.Fatalf("Invalid argument -before %q: %s", cliArgs.before, parseErr)
			}
			err = singleRun(session, &parsed)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}
