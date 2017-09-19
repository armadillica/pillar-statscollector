#!/bin/bash -e

REMOTE_HOME="/home/statscoll"
SSH="ssh -o ClearAllForwardings=yes ${DEPLOYHOST}"

case $1 in
    cloud*)
        DEPLOYHOST="$1"
        ;;
    *)
        echo "Use $0 cloud{nr}|cloud.blender.org" >&2
        exit 1
esac

# Check that we're on production branch.
if [ $(git rev-parse --abbrev-ref HEAD) != "production" ]; then
    echo "You are NOT on the production branch, refusing to deploy." >&2
    exit 1
fi

# Check that production branch has been pushed.
if [ -n "$(git log origin/production..production --oneline)" ]; then
    echo "WARNING: not all changes to the production branch have been pushed."
    echo "Press [ENTER] to continue deploying current origin/production, CTRL+C to abort."
    read dummy
fi

echo -n "Building... "
go install
VERSION="$($GOPATH/bin/pillar-statscollector -version)"
echo "version $VERSION"

if ! ping ${DEPLOYHOST} -q -c 1 -W 2 >/dev/null; then
    echo "host ${DEPLOYHOST} cannot be pinged, refusing to deploy." >&2
    exit 2
fi

echo "press [ENTER] to continue, Ctrl+C to abort."
read dummy

rsync -e "$SSH" --progress -vutp $GOPATH/bin/pillar-statscollector ${DEPLOYHOST}:${REMOTE_HOME}/pillar-statscollector

echo
echo "Pillar-Statscollector version $VERSION deployed"
