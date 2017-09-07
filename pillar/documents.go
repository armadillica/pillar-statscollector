package pillar

import "time"

type fileDoc struct {
	Backend                 string     `json:"backend"`
	LengthAggregatedInBytes int64      `json:"length_aggregate_in_bytes"`
	LinkExpires             *time.Time `json:"link_expires"`
	Status                  string     `json:"status"`
}
