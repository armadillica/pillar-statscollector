package pillar

import (
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
)

func (c *collector) projectsCount() error {
	log.Info("Aggregating project stats")

	var result struct {
		Private int `bson:"private"`
		Home    int `bson:"home"`
		Public  int `bson:"public"`
	}

	pipe := c.projColl.Pipe(c.aggrPipe([]m{
		m{"$match": notDeletedQuery},
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
	}))

	var err error
	err = pipe.One(&result)
	if err != nil && err != mgo.ErrNotFound {
		return err
	}
	if err == nil {
		c.stats.Projects.PublicCount = result.Public
		c.stats.Projects.PrivateCount = result.Private
		c.stats.Projects.HomeProjectCount = result.Home
	}

	// Do a separate count to ensure we get a correct total, even in the face of small mistakes in
	// the aggregation query.
	c.stats.Projects.TotalCount, err = c.projColl.Find(c.notDeletedQuery()).Count()
	if err != nil && err != mgo.ErrNotFound {
		return err
	}

	c.stats.Projects.TotalDeletedCount, err = c.projColl.Find(c.query(m{"_deleted": true})).Count()
	if err == mgo.ErrNotFound {
		return nil
	}
	return err
}
