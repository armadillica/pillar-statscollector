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

type CollectorBIDTestSuite struct {
	session *mgo.Session
}

var _ = check.Suite(&CollectorBIDTestSuite{})

func (s *CollectorBIDTestSuite) SetUpTest(c *check.C) {
	httpmock.Activate()

	session, err := mgo.Dial("mongodb://localhost/unittests")
	if err != nil {
		log.Panic(err)
	}

	s.session = session
}

func (s *CollectorBIDTestSuite) TearDownTest(c *check.C) {
	log.Info("SchedulerTestSuite tearing down test, dropping database.")
	s.session.DB("").DropDatabase()
	httpmock.DeactivateAndReset()
}

func (s *CollectorBIDTestSuite) TestBIDRequestHappy(t *check.C) {
	responder, err := httpmock.NewJsonResponder(200,
		blenderIDResponse{blenderIDUsers{31879, 66171, 98050}})
	assert.Nil(t, err)
	httpmock.RegisterResponder(
		"GET", "https://www.blender.org/id/api/stats",
		responder,
	)

	stats, err := CollectStats(s.session, nil)

	assert.Nil(t, err)
	if stats.BlenderID == nil {
		assert.Fail(t, "stats.BlenderID is unexpectedly nil")
		return
	}
	assert.Equal(t, 31879, stats.BlenderID.ConfirmedEmailCount)
	assert.Equal(t, 66171, stats.BlenderID.UnconfirmedEmailCount)
	assert.Equal(t, 98050, stats.BlenderID.TotalCount)
}

func (s *CollectorBIDTestSuite) TestBIDRequestUnhappy(t *check.C) {
	httpmock.RegisterResponder(
		"GET",
		"https://www.blender.org/id/api/stats",
		httpmock.NewErrorResponder(http.ErrHandlerTimeout),
	)

	stats, err := CollectStats(s.session, nil)
	assert.Nil(t, err)
	assert.Nil(t, stats.BlenderID)

	// Marshalling to JSON should exclude the Blender ID stats.
	jsonBytes, err := json.Marshal(stats)
	assert.Nil(t, err)
	unmarshalled := make(map[string]interface{})
	json.Unmarshal(jsonBytes, &unmarshalled)
	_, ok := unmarshalled["blender_id"]
	assert.False(t, ok)
}
