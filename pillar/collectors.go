package pillar

import (
	"time"

	"github.com/armadillica/pillar-statscollector/elastic"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type collector struct {
	now       time.Time
	stats     *elastic.Stats
	filesColl *mgo.Collection
	projColl  *mgo.Collection
	nodesColl *mgo.Collection
}

type m bson.M

var notDeletedQuery = m{"_deleted": m{"$ne": true}}

// collector methods are defined in the collector_xxx.go files.
