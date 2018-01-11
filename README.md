# pillar-statscollector

This little program collects various statistics from a Pillar MongoDB database, and sends it to an
ElasticSearch database for storage & analysis.


## Requirements

- A MongoDB database to collect Blender Cloud statistics from. Configure with the `-mongo` CLI option.
- A MongoDB database to store collected statistics (can be the same as above). Configure with the
  `-storage` CLI option; it defaults to the same database as above.
- A network connection to connect to the Blender Store and collect more statistics.
- An ElasticSearch server to index collected statistics. Configure with the `-elastic` CLI option.


## CLI options

Run `pillar-statscollector -help` to see the CLI options. For your initial run to see how things
work, run with `-verbose -nopush`.


## Server-side documentation

The Pillar Statscollector runs as the `statscoll` user on the Blender Cloud host. The binary is
stored in `/home/statscoll/pillar-statscollector`, and is run regularly by cron.
