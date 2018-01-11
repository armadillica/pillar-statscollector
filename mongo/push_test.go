package mongo

import (
	"github.com/armadillica/pillar-statscollector/elastic"
	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
	check "gopkg.in/check.v1"
	mgo "gopkg.in/mgo.v2"
)

type PushTestSuite struct {
	session *mgo.Session
}

var _ = check.Suite(&PushTestSuite{})

func (s *PushTestSuite) SetUpTest(c *check.C) {
	session, err := mgo.Dial("mongodb://localhost/unittests")
	if err != nil {
		log.Panic(err)
	}

	s.session = session
}

func (s *PushTestSuite) TearDownTest(c *check.C) {
	log.Info("SchedulerTestSuite tearing down test, dropping database.")
	s.session.DB("").DropDatabase()
}

func (s *PushTestSuite) TestStoreWithoutID(t *check.C) {
	stats := elastic.Stats{}
	stats.Users.SubscriberCount = 3214
	err := Push(s.session, &stats)
	assert.Nil(t, err)

	// An ID should have been generated.
	assert.NotNil(t, stats.ID)

	// We should be able to find the document in the database by this ID.
	found := elastic.Stats{}
	err = coll(s.session).FindId(stats.ID).One(&found)
	assert.Nil(t, err)
	assert.Equal(t, stats.ID, found.ID)
	assert.Equal(t, 3214, found.Users.SubscriberCount)
}

func (s *PushTestSuite) TestStoreWithID(t *check.C) {
	// Use an ID that could have been returned by ElasticSearch
	stats := elastic.Stats{ID: "AV6XOE1FyS6gf1Jm5ekm"}
	stats.Users.SubscriberCount = 3214
	err := Push(s.session, &stats)
	assert.Nil(t, err)

	// The ID should have been kept.
	assert.Equal(t, "AV6XOE1FyS6gf1Jm5ekm", stats.ID)

	// We should be able to find the document in the database by this ID.
	found := elastic.Stats{}
	err = coll(s.session).FindId("AV6XOE1FyS6gf1Jm5ekm").One(&found)
	assert.Nil(t, err)
	assert.Equal(t, stats.ID, found.ID)
	assert.Equal(t, 3214, found.Users.SubscriberCount)
}
