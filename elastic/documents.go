package elastic

import "time"

// Stats represents the JSON document pushed to ElasticSearch
type Stats struct {
	SchemaVersion                   int              `json:"schema_version"`
	Timestamp                       time.Time        `json:"timestamp"`
	ExpiredLinkCount                int              `json:"expired_link_count"`
	NoLinkCount                     int              `json:"no_link_count"`
	TotalBytesStorageUsedPerBackend map[string]int64 `json:"total_bytes_storage_used_per_backend"`
	FileCountPerStatus              map[string]int   `json:"file_count_per_status"`
	FileCountPerBackend             map[string]int   `json:"file_count_per_backend"`

	// These I really, really want to get in there, but require much more extensive querying.
	// OrphanFileCount                 int32            `json:"orphan_file_count"`
	// TotalOrphanFileSizeInBytes      int64            `json:"total_orphan_file_size_in_bytes"`
}
