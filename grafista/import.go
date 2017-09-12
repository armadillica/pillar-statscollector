package grafista

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/armadillica/pillar-statscollector/elastic"
	_ "github.com/mattn/go-sqlite3" // importing is enough to register the driver.
	log "github.com/sirupsen/logrus"
)

// PushStatsFunc is the function type that is called for each imported day of Grafista stats.
type PushStatsFunc func(stats interface{}) error

func createEmptyStatsDoc() elastic.GrafistaStats {
	statsDoc := elastic.GrafistaStats{
		SchemaVersion: 1,
	}
	statsDoc.Nodes.PublicCountPerNodeType = map[string]int{}
	statsDoc.Users.CountPerType = map[string]int{}
	return statsDoc
}

// ImportDB imports data from a Grafista SQLite database.
func ImportDB(filename string, pushToElastic PushStatsFunc) error {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return err
	}

	// Grafista has per-day statistics, so we can create a stats document for each day.
	rows, err := db.Query(
		"select date(samples.timestamp), series.name, samples.value from samples " +
			"left join series on (samples.serie_id = series.id) " +
			"order by timestamp")
	if err != nil {
		return err
	}
	var (
		timestampStr string
		timestamp    time.Time
		seriesName   string
		sampleValue  int
		statsDoc     elastic.GrafistaStats = createEmptyStatsDoc()
	)

	for rows.Next() {
		err = rows.Scan(&timestampStr, &seriesName, &sampleValue)
		if err != nil {
			return err
		}
		timestamp, err = time.Parse("2006-01-02", timestampStr)
		if err != nil {
			return fmt.Errorf("unable to parse timestamp %q: %s", timestampStr, err)
		}

		if timestamp != statsDoc.Timestamp {
			log.Infof("Stats complete, pushing to ElasticSearch: %v", statsDoc)
			if pushErr := pushToElastic(statsDoc); pushErr != nil {
				return fmt.Errorf("unable to push to Elastic: %s", pushErr)
			}

			// Reset the doc for a new iteration.
			statsDoc = createEmptyStatsDoc()
		}
		statsDoc.Timestamp = timestamp

		// Handle each field.
		switch seriesName {
		case "assets":
			statsDoc.Nodes.PublicCountPerNodeType["asset"] = sampleValue
			break
		case "comments":
			statsDoc.Nodes.PublicCountPerNodeType["comment"] = sampleValue
			break
		case "total_sold":
			statsDoc.Users.CountPerType["subscriber"] = sampleValue
			break
		case "users_total":
			statsDoc.Users.TotalCount = sampleValue
			break
		case "users_blender_sync":
			statsDoc.Users.BlenderSyncCount = sampleValue
			break
		default:
			log.Errorf("Unknown series %q, data will be lost!", seriesName)
		}
	}

	// Push the final document
	if pushErr := pushToElastic(statsDoc); pushErr != nil {
		return fmt.Errorf("unable to push to Elastic: %s", pushErr)
	}

	err = rows.Close()
	if err != nil {
		return err
	}

	return nil
}
