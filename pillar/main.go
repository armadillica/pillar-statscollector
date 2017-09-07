package pillar

import (
	"time"

	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/armadillica/pillar-statscollector/elastic"
)

type m bson.M

// CollectStats collects all the statistics and returns it as elastic.Stats object.
func CollectStats(session *mgo.Session) (elastic.Stats, error) {
	now := time.Now().UTC()
	stats := elastic.Stats{
		SchemaVersion: 1,
		Timestamp:     now,
	}

	c := collector{
		now,
		&stats,
		session.DB("").C("files"),
		session.DB("").C("projects"),
	}

	if err := c.filesExpiredLinks(); err != nil {
		return stats, err
	}
	if err := c.filesEmptyLinks(); err != nil {
		return stats, err
	}
	if err := c.filesCountStatsPerStorageBackend(); err != nil {
		return stats, err
	}
	if err := c.filesCountStatsPerStatus(); err != nil {
		return stats, err
	}
	if err := c.projectsCount(); err != nil {
		return stats, err
	}

	// Done!
	log.Info("Done collecting statistics")
	return stats, nil
}
