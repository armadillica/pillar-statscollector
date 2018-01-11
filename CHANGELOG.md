# Pillar-Statscollector Changelog


## Version 2.0 (in development)

- Dropped support for importing data from Grafista.
- Store statistics in MongoDB before pushing to ElasticSearch.
- Added CLI options for resetting (`-reset`) and reindexing (`-reindex`) ElasticSearch from the data
  in MongoDB. Both options can be given at the same time, which will perform the reset before
  reindexing.
- Added CLI option (`-reverse`) for exporting data from ElasticSearch to MongoDB. This is the
  reverse of normal operations (where the data is in MongoDB and sent to Elastic), hence the name.
  It is intended be used once to migrate from a version < 2.0, but it can be re-run as the same
  database IDs are used.
- Targeting ElasticSearch version 6.1.


## Version 1.0.1 (2017-09-19)

- Omit zero subscriber count from pushed document. This happens when we cannot reach the store; in
  that case it's better to omit the data than to send an explicit zero value.


## Version 1.0 (2017-09-18)

- First version we use in production.
