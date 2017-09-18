#!/bin/bash -e

HASH=$(git show-ref --head --hash HEAD | head -n1)
BUILDDATE=$(date --rfc-3339=date)

cat > VERSION.txt <<EOT
Git revision: ${HASH}
Built on: ${BUILDDATE}
EOT

# We can't statically build due to the sqlite3-binding.c source file.
# Fortunately the binary won't depend on a sqlite3.so file.
go install github.com/armadillica/pillar-statscollector
cp -a $GOPATH/bin/pillar-statscollector .
strip pillar-statscollector

# Use the executable to build our Docker image.
docker build -t armadillica/pillar-statscollector:${HASH} .
docker tag armadillica/pillar-statscollector:${HASH} armadillica/pillar-statscollector:latest
