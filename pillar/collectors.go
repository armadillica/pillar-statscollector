package pillar

import (
	"fmt"
	"time"

	"github.com/armadillica/pillar-statscollector/elastic"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type m bson.M

type collector struct {
	now        time.Time
	stats      *elastic.Stats
	extraQuery *m
	filesColl  *mgo.Collection
	projColl   *mgo.Collection
	nodesColl  *mgo.Collection
}

var notDeletedQuery = m{"_deleted": m{"$ne": true}}

// collector methods are defined in the collector_xxx.go files.

// CollectStats collects all the statistics and returns it as elastic.Stats object.
func CollectStats(session *mgo.Session, before *time.Time) (elastic.Stats, error) {
	var extraQuery *m
	var now time.Time

	if before == nil {
		now = time.Now().UTC()
		log.Info("Collecting current statistics")
	} else {
		now = *before
		extraQuery = &m{"_created": m{"$lt": before}}
		log.Infof("Collecting statistics before %s", now)
	}

	stats := elastic.Stats{
		SchemaVersion: 1,
		Timestamp:     now,
	}

	c := collector{
		now,
		&stats,
		extraQuery,
		session.DB("").C("files"),
		session.DB("").C("projects"),
		session.DB("").C("nodes"),
	}

	if err := c.filesTotalCount(); err != nil {
		return stats, err
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

	if err := c.nodesCount(); err != nil {
		return stats, err
	}

	// Done!
	log.Info("Done collecting statistics")
	return stats, nil
}

// aggrPipe(p) returns the given pipeline, possibly prepended with a $match: c.extraQuery.
func (c *collector) aggrPipe(pipeline []m) []m {
	if c.extraQuery == nil {
		return pipeline
	}

	pipe := []m{
		m{"$match": *c.extraQuery},
	}
	return append(pipe, pipeline...)
}

// query returns c.extraQuery or an empty map.
// The returned value is a copy, so can be modified without side-effects.
func (c *collector) emptyQuery() m {
	query := m{}

	// Copy the extra query if it's there.
	if c.extraQuery != nil {
		for k, v := range *c.extraQuery {
			query[k] = v
		}
	}
	return query
}

// query returns the given query, possibly combined with c.extraQuery.
// The returned value is a copy, so can be modified without side-effects.
func (c *collector) query(q m) m {
	query := c.emptyQuery()

	// Copy the given query.
	for k, v := range q {
		query[k] = v
	}

	return query
}

// notDeletedQuery returns a "not deleted" query, possibly combined with c.extraQuery.
// The returned value is a copy, so can be modified without side-effects.
func (c *collector) notDeletedQuery() m {
	return c.query(notDeletedQuery)
}
