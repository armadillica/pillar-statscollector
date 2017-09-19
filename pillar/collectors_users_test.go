package pillar

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
	check "gopkg.in/check.v1"
	"gopkg.in/jarcoal/httpmock.v1"
	mgo "gopkg.in/mgo.v2"
)

type CollectorsUsersTestSuite struct {
	session *mgo.Session
}

var _ = check.Suite(&CollectorsUsersTestSuite{})

func (s *CollectorsUsersTestSuite) SetUpTest(c *check.C) {
	httpmock.Activate()

	session, err := mgo.Dial("mongodb://localhost/unittests")
	if err != nil {
		log.Panic(err)
	}

	s.session = session
}

func (s *CollectorsUsersTestSuite) TearDownTest(c *check.C) {
	log.Info("SchedulerTestSuite tearing down test, dropping database.")
	httpmock.DeactivateAndReset()
}

func (s *CollectorsUsersTestSuite) TestStoreRequestHappy(t *check.C) {
	responder, err := httpmock.NewJsonResponder(200, storeResponse{456})
	assert.Nil(t, err)
	httpmock.RegisterResponder(
		"GET", "https://store.blender.org/product-counter/?prod=cloud",
		responder,
	)

	stats, err := CollectStats(s.session, nil)

	assert.Nil(t, err)
	assert.Equal(t, 456, stats.Users.SubscriberCount)
}

func (s *CollectorsUsersTestSuite) TestStoreRequestUnhappy(t *check.C) {
	httpmock.RegisterResponder(
		"GET",
		"https://store.blender.org/product-counter/?prod=cloud",
		httpmock.NewErrorResponder(http.ErrHandlerTimeout),
	)

	stats, err := CollectStats(s.session, nil)
	assert.Nil(t, err)
	assert.Zero(t, stats.Users.SubscriberCount)

	// Marshalling to JSON should exclude the susbcriber count.
	jsonBytes, err := json.Marshal(stats)
	assert.Nil(t, err)
	unmarshalled := make(map[string]interface{})
	json.Unmarshal(jsonBytes, &unmarshalled)
	users, ok := unmarshalled["users"]
	assert.True(t, ok)
	usersKnownType, ok := users.(map[string]interface{})
	assert.True(t, ok)

	_, ok = usersKnownType["subscriber_count"]
	assert.False(t, ok)
}
