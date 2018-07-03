package elastic

import "time"

// Stats represents the JSON document pushed to ElasticSearch
type Stats struct {
	ID            string    `json:"-" bson:"_id,omitempty"` // used by MongoDB but not by ElasticSearch.
	SchemaVersion int       `json:"stats_schema_version" bson:"stats_schema_version"`
	Timestamp     time.Time `json:"timestamp" bson:"timestamp"`

	Files struct {
		ExpiredLinkCount                int              `json:"expired_link_count" bson:"expired_link_count"`
		NoLinkCount                     int              `json:"no_link_count" bson:"no_link_count"`
		TotalBytesStorageUsed           int64            `json:"total_bytes_storage_used" bson:"total_bytes_storage_used"`
		TotalBytesStorageUsedPerBackend map[string]int64 `json:"total_bytes_storage_used_per_backend" bson:"total_bytes_storage_used_per_backend"`
		FileCountTotal                  int              `json:"file_count_total" bson:"file_count_total"`
		FileCountPerStatus              map[string]int   `json:"file_count_per_status" bson:"file_count_per_status"`
		FileCountPerBackend             map[string]int   `json:"file_count_per_backend" bson:"file_count_per_backend"`
		// These I really, really want to get in there, but require much more extensive querying.
		// OrphanFileCount                 int32            `json:"orphan_file_count" bson:"orphan_file_count"`
		// TotalOrphanFileSizeInBytes      int64            `json:"total_orphan_file_size_in_bytes" bson:"total_orphan_file_size_in_bytes"`
	} `json:"files" bson:"files"`

	Projects struct {
		PublicCount       int `json:"public_count" bson:"public_count"`
		PrivateCount      int `json:"private_count" bson:"private_count"`
		HomeProjectCount  int `json:"home_project_count" bson:"home_project_count"`
		TotalCount        int `json:"total_count" bson:"total_count"`
		TotalDeletedCount int `json:"total_deleted_count" bson:"total_deleted_count"`
	} `json:"projects" bson:"projects"`

	Nodes struct {
		PublicCountPerNodeType map[string]int `json:"public_node_count_per_type" bson:"public_node_count_per_type"`
		TotalPublicNodeCount   int            `json:"total_public_node_count" bson:"total_public_node_count"`
	} `json:"nodes" bson:"nodes"`

	Users struct {
		TotalCount         int            `json:"total_user_count" bson:"total_user_count"`
		TotalRealUserCount int            `json:"total_real_user_count" bson:"total_real_user_count"`
		CountPerType       map[string]int `json:"count_per_type" bson:"count_per_type"`
		BlenderSyncCount   int            `json:"blender_sync_count" bson:"blender_sync_count"`

		// SubscriberCount comes from the Store, which can be unreachable at times. Rather than
		// passing an explicit count of 0 to ElasticSearch, it's better to omit the key completely
		// and deal with it as missing data.
		SubscriberCount int `json:"subscriber_count,omitempty" bson:"subscriber_count,omitempty"`
	} `json:"users" bson:"users"`

	BlenderID *BlenderID `json:"blender_id,omitempty" bson:"blender_id,omitempty"`
}

// BlenderID models the stats from Blender ID
type BlenderID struct {
	ConfirmedEmailCount   int                    `json:"confirmed_email_count" bson:"confirmed_email_count"`
	UnconfirmedEmailCount int                    `json:"unconfirmed_email_count" bson:"unconfirmed_email_count"`
	PrivacyPolicyAgreed   BlenderIDPrivacyPolicy `json:"privacy_policy_agreed" bson:"privacy_policy_agreed"`
	TotalCount            int                    `json:"total_user_count" bson:"total_user_count"`
}

// BlenderIDPrivacyPolicy is a subdocument of BlenderID and counts user agreements to the privacy policy.
type BlenderIDPrivacyPolicy struct {
	Latest   int `json:"latest"`
	Obsolete int `json:"obsolete"`
	Never    int `json:"never"`
}

type postResponse struct {
	Index   string `json:"_index" bson:"_index"`
	Type    string `json:"_type" bson:"_type"`
	ID      string `json:"_id" bson:"_id"`
	Version int    `json:"_version" bson:"_version"`
	Created bool   `json:"created" bson:"created"`
}
