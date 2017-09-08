package pillar

import (
	"time"

	"github.com/armadillica/pillar-statscollector/elastic"
	mgo "gopkg.in/mgo.v2"
)

type collector struct {
	now       time.Time
	stats     *elastic.Stats
	filesColl *mgo.Collection
	projColl  *mgo.Collection
}

// collector methods are defined in the collector_xxx.go files.
