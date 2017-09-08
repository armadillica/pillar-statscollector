package pillar

import (
	log "github.com/sirupsen/logrus"
)

func (c *collector) publicNodesCount(match m) (int, error) {
	var result struct {
		Total int `bson:"total"`
	}

	pipe := c.nodesColl.Pipe([]m{
		m{"$match": notDeletedQuery},
		m{"$match": match},
		m{"$lookup": m{
			"from":         "projects",
			"localField":   "project",
			"foreignField": "_id",
			"as":           "project",
		}},
		m{"$unwind": m{"path": "$project"}},
		m{"$project": m{"project.is_private": 1}},
		m{"$match": m{"project.is_private": m{"$ne": true}}},
		m{"$count": "total"},
	})

	err := pipe.One(&result)
	if err != nil && err != mgo.ErrNotFound {
		return 0, err
	}

	return result.Total, nil
}

func (c *collector) nodesCount() error {
	log.Info("Aggregating nodes stats")

	var err error
	if c.stats.Nodes.PublicAssetCount, err = c.publicNodesCount(m{"node_type": "asset"}); err != nil {
		return err
	}

	if c.stats.Nodes.PublicCommentCount, err = c.publicNodesCount(m{"node_type": "comment"}); err != nil {
		return err
	}

	return nil
}
