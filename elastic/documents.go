package elastic

import "time"

// Stats represents the JSON document pushed to ElasticSearch
type Stats struct {
	SchemaVersion int       `json:"stats_schema_version"`
	Timestamp     time.Time `json:"timestamp"`

	Files struct {
		ExpiredLinkCount                int              `json:"expired_link_count"`
		NoLinkCount                     int              `json:"no_link_count"`
		TotalBytesStorageUsed           int64            `json:"total_bytes_storage_used"`
		TotalBytesStorageUsedPerBackend map[string]int64 `json:"total_bytes_storage_used_per_backend"`
		FileCountTotal                  int              `json:"file_count_total"`
		FileCountPerStatus              map[string]int   `json:"file_count_per_status"`
		FileCountPerBackend             map[string]int   `json:"file_count_per_backend"`
		// These I really, really want to get in there, but require much more extensive querying.
		// OrphanFileCount                 int32            `json:"orphan_file_count"`
		// TotalOrphanFileSizeInBytes      int64            `json:"total_orphan_file_size_in_bytes"`
	} `json:"files"`

	Projects struct {
		PublicCount       int `json:"public_count"`
		PrivateCount      int `json:"private_count"`
		HomeProjectCount  int `json:"home_project_count"`
		TotalCount        int `json:"total_count"`
		TotalDeletedCount int `json:"total_deleted_count"`
	} `json:"projects"`

	Nodes struct {
		PublicCountPerNodeType map[string]int `json:"public_node_count_per_type"`
		TotalPublicNodeCount   int            `json:"total_public_node_count"`
	} `json:"nodes"`

	Users struct {
		TotalCount         int            `json:"total_user_count"`
		TotalRealUserCount int            `json:"total_real_user_count"`
		CountPerType       map[string]int `json:"count_per_type"`
		BlenderSyncCount   int            `json:"blender_sync_count"`

		// SubscriberCount comes from the Store, which can be unreachable at times. Rather than
		// passing an explicit count of 0 to ElasticSearch, it's better to omit the key completely
		// and deal with it as missing data.
		SubscriberCount int `json:"subscriber_count,omitempty"`
	} `json:"users"`
}

// GrafistaStats represents the JSON document pushed to ElasticSearch as fetched from Grafista.
// It contains historical data from a subset of the full set of stats described above.
type GrafistaStats struct {
	SchemaVersion int       `json:"stats_schema_version"`
	Timestamp     time.Time `json:"timestamp"`

	Nodes struct {
		PublicCountPerNodeType map[string]int `json:"public_node_count_per_type"`
	} `json:"nodes"`

	Users struct {
		TotalCount       int `json:"total_user_count,omitempty"`
		BlenderSyncCount int `json:"blender_sync_count,omitempty"`
		SubscriberCount  int `json:"subscriber_count,omitempty"`
	} `json:"users"`
}

type postResponse struct {
	Index   string `json:"_index"`
	Type    string `json:"_type"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Created bool   `json:"created"`
}
