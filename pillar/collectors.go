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
	projColl  *mgo.Collection
}

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

func (c *collector) projectsCount() error {
	log.Info("Aggregating project stats")

	var result struct {
		Private int `bson:"private"`
		Home    int `bson:"home"`
		Public  int `bson:"public"`
	}
	// project:
	// is_private: {$and: [{$eq: ["$is_private", true]}, {$ne: ["$category", "home"]}]},
	// is_home: {$eq: ["$category", "home"]},
	// is_public: {$and: [{$eq: ["$is_private", false]}, {$ne: ["$category", "home"]}]},
	// group:
	// _id: null,
	// home: {$sum: {$cond: {if: "$is_home", then: 1, else: 0}}},
	// public: {$sum: {$cond: {if: "$is_public", then: 1, else: 0}}},
	// private: {$sum: {$cond: {if: "$is_private", then: 1, else: 0}}},

	pipe := c.projColl.Pipe([]m{
		m{"$project": m{
			"is_private": m{"$and": []m{
				m{"$eq": []interface{}{"$is_private", true}},
				m{"$ne": []interface{}{"$category", "home"}},
			}},
			"is_home": m{"$eq": []interface{}{"$category", "home"}},
			"is_public": m{"$and": []m{
				m{"$eq": []interface{}{"$is_private", false}},
				m{"$ne": []interface{}{"$category", "home"}},
			}},
		}},
		m{"$group": m{
			"_id":     nil,
			"home":    m{"$sum": m{"$cond": m{"if": "$is_home", "then": 1, "else": 0}}},
			"public":  m{"$sum": m{"$cond": m{"if": "$is_public", "then": 1, "else": 0}}},
			"private": m{"$sum": m{"$cond": m{"if": "$is_private", "then": 1, "else": 0}}},
		}},
	})

	if err := pipe.One(&result); err != nil {
		return err
	}

	c.stats.Projects.PublicCount = result.Public
	c.stats.Projects.PrivateCount = result.Private
	c.stats.Projects.HomeProjectCount = result.Home

	// Do a separate count to ensure we get a correct total, even in the face of small mistakes in
	// the aggregation query.
	var err error
	c.stats.Projects.TotalCount, err = c.projColl.Count()
	return err
}
