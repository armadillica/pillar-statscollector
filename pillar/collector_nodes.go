package pillar

import log "github.com/sirupsen/logrus"

func (c *collector) nodesCount() error {
	log.Info("Aggregating nodes stats")

	var result struct {
		NodeType string `bson:"_id"`
		Count    int    `bson:"count"`
	}

	query := c.aggrPipe([]m{
		// Only inspect public projects.
		m{"$match": m{"is_private": m{"$ne": true}}},
		// Find all nodes for these projects.
		m{"$lookup": m{
			"from":         "nodes",
			"localField":   "_id",
			"foreignField": "project",
			"as":           "nodes",
		}},
		// Count per node type.
		m{"$unwind": m{"path": "$nodes"}},
		m{"$group": m{
			"_id":   "$nodes.node_type",
			"count": m{"$sum": 1},
		}},
	})
	iter := c.projColl.Pipe(query).Iter()

	c.stats.Nodes.PublicCountPerNodeType = map[string]int{}
	c.stats.Nodes.TotalPublicNodeCount = 0

	for iter.Next(&result) {
		nodeType := result.NodeType
		if nodeType == "" {
			nodeType = noValueString
		}
		c.stats.Nodes.PublicCountPerNodeType[nodeType] = result.Count
		c.stats.Nodes.TotalPublicNodeCount += result.Count
	}

	return iter.Close()
}
