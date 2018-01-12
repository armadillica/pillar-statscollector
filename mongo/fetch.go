package mongo

import (
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// All returns all documents in the stats collection.
func All(mgoStats *mgo.Session) chan bson.M {
	log.Warn("retrieving all documents in MongoDB")
	c := mgoStats.DB("").C(StatsCollection)
	ch := make(chan bson.M)

	go func() {
		defer close(ch)

		result := &bson.M{}
		seen := 0

		q := c.Find(bson.M{})
		count, err := q.Count()
		if err != nil {
			log.WithError(err).Fatal("error counting documents in MongoDB")
		}

		iter := q.Iter()
		for iter.Next(result) {
			seen++
			log.WithField("id", (*result)["_id"]).Debug("found document in MongoDB")
			ch <- *result

			if seen%100 == 0 {
				log.WithFields(log.Fields{
					"seen":  seen,
					"total": count,
				}).Warn("retrieving documents from MongoDB")
			}

			// Create a new object for iter.Next() to use, to ensure that the receiving
			// end of the channel can do with the received object as they please.
			result = &bson.M{}
		}
		if err := iter.Close(); err != nil {
			log.WithError(err).Fatal("error querying MongoDB")
		}
		log.WithField("seen", seen).Info("all documents in MongoDB retrieved")
	}()

	return ch
}
