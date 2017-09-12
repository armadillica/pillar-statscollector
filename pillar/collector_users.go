package pillar

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
)

func (c *collector) usersCount() error {
	log.Info("Aggregating users stats")

	var perTypeResult struct {
		Type  string `bson:"_id"`
		Count int    `bson:"count"`
	}

	var err error
	totalUserCount, err := c.usersColl.Find(c.emptyQuery()).Count()
	if err != nil && err != mgo.ErrNotFound {
		return err
	}

	pipe := c.usersColl.Pipe(c.aggrPipe([]m{
		m{"$match": m{"roles": m{"$in": []string{"service", "demo", "subscriber"}}}},
		m{"$project": m{
			"roles": 1,
			"type": m{"$cond": m{
				"if":   m{"$in": []string{"service", "$roles"}},
				"then": "service",
				"else": m{"$cond": m{
					"if":   m{"$in": []string{"demo", "$roles"}},
					"then": "demo",
					"else": m{"$cond": m{
						"if":   m{"$in": []string{"subscriber", "$roles"}},
						"then": "subscriber",
						"else": "unknown"},
					}}}}}},
		},
		m{"$group": m{
			"_id":   "$type",
			"count": m{"$sum": 1},
		}},
	}))
	iter := pipe.Iter()

	c.stats.Users.CountPerType = map[string]int{}

	for iter.Next(&perTypeResult) {
		c.stats.Users.CountPerType[perTypeResult.Type] = perTypeResult.Count
	}

	serviceUsercount := c.stats.Users.CountPerType["service"]
	c.stats.Users.TotalRealUserCount = totalUserCount - serviceUsercount
	c.stats.Users.TotalCount = totalUserCount

	return iter.Close()
}

func (c *collector) countBlenderSyncUsers() error {
	log.Info("Counting Blender Sync users")

	var result struct {
		Total int `bson:"total"`
	}

	pipe := c.nodesColl.Pipe(c.aggrPipe([]m{
		// 0 Find all startups.blend that are not deleted
		m{"$match": m{
			"_deleted": m{"$ne": true},
			"name":     "startup.blend",
		}},
		// 1 Group them per project (drops any duplicates)
		m{"$group": m{"_id": "$project"}},
		// 2 Join the project info
		m{"$lookup": m{
			"from":         "projects",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "project",
		}},
		// 3 Unwind the project list (there is always only one project)
		m{"$unwind": m{"path": "$project"}},
		// 4 Find all home projects
		m{"$match": m{"project.category": "home"}},
		m{"$count": "total"},
	}))

	var err error
	err = pipe.One(&result)
	if err != nil && err != mgo.ErrNotFound {
		return err
	}

	c.stats.Users.BlenderSyncCount = result.Total
	return nil
}

// Connects to Blender Store to fetch the current number of subscriptions.
func (c *collector) countSubscriptions(storeURL string) error {
	log.Infof("Connecting to %s", storeURL)

	resp, err := http.Get(storeURL)
	if err != nil {
		return fmt.Errorf("error getting Blender Store stats: %s", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("error %d getting Blender Store stats", resp.StatusCode)
	}

	var storeData struct {
		Total int `json:"total_sold"`
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&storeData); err != nil {
		return fmt.Errorf("error decoding response from store: %s", err)
	}

	c.stats.Users.SubscriberCount = storeData.Total
	return nil
}
