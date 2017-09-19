# Pillar-Statscollector Changelog


## Version 1.0.1 (2017-09-19)

- Omit zero subscriber count from pushed document. This happens when we cannot reach the store; in
  that case it's better to omit the data than to send an explicit zero value.


## Version 1.0 (2017-09-18)

- First version we use in production.
