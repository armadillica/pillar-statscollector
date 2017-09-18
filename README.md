# pillar-statscollector

This little program collects various statistics from a Pillar MongoDB database, and sends it to an
ElasticSearch database for storage & analysis.


## CLI options

Run `pillar-statscollector -help` to see the CLI options. For your initial run to see how things
work, run with `-verbose -nopush`.


## Server-side documentation

The Pillar Statscollector runs as the `statscoll` user on the Blender Cloud host. The binary is
stored in `/home/statscoll/pillar-statscollector`, and is run regularly by cron.
