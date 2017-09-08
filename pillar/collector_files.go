package pillar

import (
	log "github.com/sirupsen/logrus"
)

func (c *collector) filesTotalCount() error {
	var err error

	log.Info("Counting files")
	c.stats.Files.FileCountTotal, err = c.filesColl.Count()

	return err
}

func (c *collector) filesExpiredLinks() error {
	var err error

	log.Info("Counting files with expired links")
	c.stats.Files.ExpiredLinkCount, err = c.filesColl.Find(m{"link_expires": m{"$lt": c.now}}).Count()

	return err
}

func (c *collector) filesEmptyLinks() error {
	var err error

	log.Info("Counting files with nil/empty links")
	c.stats.Files.NoLinkCount, err = c.filesColl.Find(
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

	c.stats.Files.TotalBytesStorageUsedPerBackend = map[string]int64{}
	c.stats.Files.FileCountPerBackend = map[string]int{}

	for iter.Next(&perBackendResult) {
		backend := perBackendResult.Backend
		c.stats.Files.TotalBytesStorageUsedPerBackend[backend] = perBackendResult.TotalBytes
		c.stats.Files.FileCountPerBackend[backend] = perBackendResult.Count
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

	c.stats.Files.FileCountPerStatus = map[string]int{}

	for iter.Next(&perStatusResult) {
		status := perStatusResult.Status
		c.stats.Files.FileCountPerStatus[status] = perStatusResult.Count
	}
	return iter.Close()
}
