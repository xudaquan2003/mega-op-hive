#!/bin/bash

failed=""
sims=$(find simulators -name go.mod)
for d in $sims; do
    d="$(dirname $d)"
    echo "building $d"
    ( cd $d; go build . )
    status=$?
    if [ $status -ne 0 ]; then
        failed="y"
        echo "build failed with exit status $status"
    fi
done

# Exit with non-zero status code if any build failed.
if [ -n "$failed" ]; then
    exit 1
fi
