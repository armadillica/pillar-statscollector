package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/armadillica/pillar-statscollector/elastic"
	"github.com/armadillica/pillar-statscollector/grafista"
	"github.com/armadillica/pillar-statscollector/mongo"
	"github.com/armadillica/pillar-statscollector/pillar"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
)

const statscollectorVersion = "1.0.1"

var cliArgs struct {
	version         bool
	verbose         bool
	debug           bool
	mongoURL        string
	mongoStorageURL string
	elasticURL      string
	before          string
	nopush          bool
	allSince        string
	importGrafista  string
	reverseToMongo  bool
	reindex         bool
}

func parseCliArgs() {
	flag.BoolVar(&cliArgs.version, "version", false, "Shows the application version, then exits.")
	flag.BoolVar(&cliArgs.verbose, "verbose", false, "Enable info-level logging.")
	flag.BoolVar(&cliArgs.debug, "debug", false, "Enable debug-level logging.")
	flag.BoolVar(&cliArgs.nopush, "nopush", false, "Log statistics, but don't push to ElasticSearch.")
	flag.StringVar(&cliArgs.mongoURL, "mongo", "mongodb://localhost/cloud", "URL of the MongoDB database to read from.")
	flag.StringVar(&cliArgs.mongoStorageURL, "storage", "", "URL of the MongoDB database to store the Cloud statistics to. Defaults to the URL to read from")
	flag.StringVar(&cliArgs.elasticURL, "elastic", "http://localhost:9200/cloudstats/stats/", "URL of the ElasticSearch instance to push to.")
	flag.StringVar(&cliArgs.before, "before", "", "Only consider objects created before this timestamp; expected in RFC 3339 format.")
	flag.StringVar(&cliArgs.allSince, "allsince", "", "Collect daily statistics since this timestamp until now; expected in RFC 3339 format.")
	flag.StringVar(&cliArgs.importGrafista, "import", "", "Imports data from a Grafista SQLite database and pushes stats to ElasticSearch.")
	flag.BoolVar(&cliArgs.reverseToMongo, "reverse", false, "Query ElasticSearch and store data in MongoDB, which is the reverse of normal operations.")
	flag.BoolVar(&cliArgs.reindex, "reindex", false, "Reindex ElasticSearch from data stored in MongoDB.")
	flag.Parse()

	if cliArgs.mongoStorageURL == "" {
		cliArgs.mongoStorageURL = cliArgs.mongoURL
	}
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

	return pushStats(session, stats)
}

func importFromGrafista(session *mgo.Session, dbFilename string) error {
	push := func(stats interface{}) error {
		return pushStats(session, stats)
	}
	err := grafista.ImportDB(cliArgs.importGrafista, push)
	if err != nil {
		return fmt.Errorf("error importing from Grafista DB: %s", err)
	}
	return nil
}

func importFromElastic(mgoWrite *mgo.Session) error {
	log.Warning("reverse-importing from ElasticSearch to MongoDB")
	return nil
}

func pushStats(session *mgo.Session, stats interface{}) error {
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

	if err := mongo.Push(session, stats); err != nil {
		return fmt.Errorf("error pushing to MongoDB: %s", err)
	}

	if err := elastic.Push(cliArgs.elasticURL, stats); err != nil {
		return fmt.Errorf("error pushing to ElasticSearch: %s", err)
	}
	return nil
}

func connectMongoDB() (mgoCloud, mgoStats *mgo.Session) {
	var err error

	if cliArgs.mongoStorageURL == "" || cliArgs.mongoStorageURL == cliArgs.mongoURL {
		log.WithField("url", cliArgs.mongoURL).Info("connecting to MongoDB for cloud+stats")
		mgoCloud, err = mgo.Dial(cliArgs.mongoURL)
		if err != nil {
			log.Panic(err)
		}
		mgoCloud.SetMode(mgo.Monotonic, true)
		mgoStats = mgoCloud
		return
	}

	log.WithField("url", cliArgs.mongoURL).Info("connecting to MongoDB for cloud")
	mgoCloud, err = mgo.Dial(cliArgs.mongoURL)
	if err != nil {
		log.Panic(err)
	}

	log.WithField("url", cliArgs.mongoStorageURL).Info("connecting to MongoDB for stats")
	mgoStats, err = mgo.Dial(cliArgs.mongoStorageURL)
	if err != nil {
		log.Panic(err)
	}
	mgoCloud.SetMode(mgo.Monotonic, true)
	mgoStats.SetMode(mgo.Monotonic, true)

	return
}

func reverseToMongo(mgoStats *mgo.Session) {
	ch := elastic.ReverseImport(cliArgs.elasticURL)
	log.Debug("waiting for documents to arrive on the channel")
	for hit := range ch {
		mongo.PushHit(mgoStats, hit)
	}
	log.Info("done reverse-importing")
}

func reindex(mgoStats *mgo.Session) {
	ch := mongo.All(mgoStats)
	log.Debug("waiting for documents to arrive on the channel")
	for stats := range ch {
		elastic.Push(cliArgs.elasticURL, stats)
	}
	log.Info("done reindexing")
}

func main() {
	parseCliArgs()
	if cliArgs.version {
		fmt.Println(statscollectorVersion)
		return
	}

	configLogging()
	mgoCloud, mgoStats := connectMongoDB()

	var err error
	if cliArgs.importGrafista != "" {
		err = importFromGrafista(mgoCloud, cliArgs.importGrafista)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if cliArgs.reverseToMongo && cliArgs.reindex {
		log.Fatal("-reverse and -reindex are mutually exclusive")
	}

	if cliArgs.reverseToMongo {
		reverseToMongo(mgoStats)
		return
	}

	if cliArgs.reindex {
		reindex(mgoStats)
		return
	}

	if cliArgs.allSince != "" {
		if cliArgs.before != "" {
			log.Fatalf("Use either -before or -allsince, not both.")
		}

		beginTimestamp, parseErr := time.Parse(time.RFC3339, cliArgs.allSince)
		if parseErr != nil {
			log.Fatalf("Invalid argument -allsince %q: %s", cliArgs.allSince, parseErr)
		}

		err = collectAllSince(mgoCloud, beginTimestamp)
	} else {
		if cliArgs.before == "" {
			err = singleRun(mgoCloud, nil)
		} else {
			parsed, parseErr := time.Parse(time.RFC3339, cliArgs.before)
			if parseErr != nil {
				log.Fatalf("Invalid argument -before %q: %s", cliArgs.before, parseErr)
			}
			err = singleRun(mgoCloud, &parsed)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}
