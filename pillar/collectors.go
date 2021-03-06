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
	usersColl  *mgo.Collection
}

var notDeletedQuery = m{"_deleted": m{"$ne": true}}

const noValueString = "-none-" // Used to prevent empty keys in maps.

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
		session.DB("").C("users"),
	}

	// Collect subscribers/Blender ID in another goroutine, because they need to do an HTTP call.
	storeDone := make(chan error)
	bidDone := make(chan error)
	go func() {
		subsErr := c.countSubscriptions("https://store.blender.org/product-counter/?prod=cloud")
		storeDone <- subsErr
	}()
	go func() {
		bidErr := c.countBlenderID("https://www.blender.org/id/api/stats")
		bidDone <- bidErr
	}()

	if err := c.filesTotalCount(); err != nil {
		return stats, fmt.Errorf("filesTotalCount: %s", err)
	}
	if err := c.filesExpiredLinks(); err != nil {
		return stats, fmt.Errorf("filesExpiredLinks: %s", err)
	}
	if err := c.filesEmptyLinks(); err != nil {
		return stats, fmt.Errorf("filesEmptyLinks: %s", err)
	}
	if err := c.filesCountStatsPerStorageBackend(); err != nil {
		return stats, fmt.Errorf("filesCountStatsPerStorageBackend: %s", err)
	}
	if err := c.filesCountStatsPerStatus(); err != nil {
		return stats, fmt.Errorf("filesCountStatsPerStatus: %s", err)
	}

	if err := c.projectsCount(); err != nil {
		return stats, fmt.Errorf("projectsCount: %s", err)
	}

	if err := c.nodesCount(); err != nil {
		return stats, fmt.Errorf("nodesCount: %s", err)
	}

	if err := c.usersCount(); err != nil {
		return stats, fmt.Errorf("usersCount: %s", err)
	}
	if err := c.countBlenderSyncUsers(); err != nil {
		return stats, fmt.Errorf("countBlenderSyncUsers: %s", err)
	}

	// Wait for the remote calls to be done.
	if err := <-storeDone; err != nil {
		log.Warningf("Ignoring error from store: %s", err)
	}
	if err := <-bidDone; err != nil {
		log.Warningf("Ignoring error from Blender ID: %s", err)
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
