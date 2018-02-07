package pillar

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/armadillica/pillar-statscollector/elastic"
	log "github.com/sirupsen/logrus"
)

// Connects to Blender ID to fetch user stats.
func (c *collector) countBlenderID(blenderIDURL string) error {
	log.WithField("url", blenderIDURL).Info("connecting to Blender ID")

	resp, err := http.Get(blenderIDURL)
	if err != nil {
		return fmt.Errorf("error getting Blender ID stats: %s", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("error %d getting Blender ID stats", resp.StatusCode)
	}

	var blenderIDData blenderIDResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&blenderIDData); err != nil {
		return fmt.Errorf("error decoding response from Blender ID: %s", err)
	}

	c.stats.BlenderID = &elastic.BlenderID{
		ConfirmedEmailCount:   blenderIDData.Users.ConfirmedEmailCount,
		UnconfirmedEmailCount: blenderIDData.Users.UnconfirmedEmailCount,
		TotalCount:            blenderIDData.Users.TotalCount,
	}
	return nil
}
