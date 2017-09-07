package pillar

import (
	"time"

	"github.com/armadillica/pillar-statscollector/elastic"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
)

type collector struct {
	now       time.Time
	stats     *elastic.Stats
	filesColl *mgo.Collection
}

func (c *collector) filesExpiredLinks() error {
	var err error

	log.Info("Counting files with expired links")
	c.stats.ExpiredLinkCount, err = c.filesColl.Find(m{"link_expires": m{"$lt": c.now}}).Count()

	return err
}

func (c *collector) filesEmptyLinks() error {
	var err error

	log.Info("Counting files with nil/empty links")
	c.stats.NoLinkCount, err = c.filesColl.Find(
		m{"$or": []m{
			m{"link": nil},
			m{"link": m{"$exists": false}},
			m{"link": ""},
		}}).Count()

	return err
}

func (c *collector) filesCountStatsPerStorageBackend() error {
	log.Info("Aggregating file statistics per storage backend")

	var perBackendResult struct {
		Backend    string `bson:"_id"`
		Count      int    `bson:"count"`
		TotalBytes int64  `bson:"total_bytes"`
	}

	pipe := c.filesColl.Pipe([]m{
		m{"$group": m{
			"_id":         "$backend",
			"count":       m{"$sum": 1},
			"total_bytes": m{"$sum": "$length_aggregate_in_bytes"},
		}},
	})
	iter := pipe.Iter()
	for iter.Next(&perBackendResult) {
		backend := perBackendResult.Backend
		c.stats.TotalBytesStorageUsedPerBackend[backend] = perBackendResult.TotalBytes
		c.stats.FileCountPerBackend[backend] = perBackendResult.Count
	}

	return iter.Close()
}

func (c *collector) filesCountStatsPerStatus() error {
	log.Info("Aggregating file statistics per status")

	var perStatusResult struct {
		Status string `bson:"_id"`
		Count  int    `bson:"count"`
	}

	pipe := c.filesColl.Pipe([]m{
		m{"$group": m{
			"_id":   "$status",
			"count": m{"$sum": 1},
		}},
	})
	iter := pipe.Iter()
	for iter.Next(&perStatusResult) {
		status := perStatusResult.Status
		c.stats.FileCountPerStatus[status] = perStatusResult.Count
	}
	return iter.Close()
}
