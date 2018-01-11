package mongo

import (
	"errors"

	"github.com/armadillica/pillar-statscollector/elastic"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var errMongoStoreError = errors.New("error storing document in MongoDB")

// PushHit stores a hit from ElasticSearch in MongoDB.
func PushHit(mgoStats *mgo.Session, hit elastic.Hit) error {
	logger := log.WithField("ID", hit.ID)
	logger.WithField("stats", hit.Source).Debug("storing in MongoDB")

	c := mgoStats.DB("").C(StatsCollection)
	info, err := c.UpsertId(hit.ID, bson.M{"$set": hit.Source})
	if err != nil {
		logger.WithError(err).Error("unable to store hit from Elastic in Mongo")
		return errMongoStoreError
	}
	logger.WithFields(log.Fields{
		"matched": info.Matched,
		"updated": info.Updated,
	}).Debug("stored document")

	return nil
}

// Push stores a stats document in MongoDB.
func Push(mgoStats *mgo.Session, stats interface{}) error {
	log.WithField("stats", stats).Debug("storing in MongoDB")

	return errors.New("mongo.Push() not implemented")
}
