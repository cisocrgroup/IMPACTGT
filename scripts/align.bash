#!/bin/bash
set -e

if [[ $# -lt 1 ]]; then
    echo "usage $0 DIR"
    exit 1
fi

dir=$1

# align
echo alignes $dir/*.json
alignes $dir/*.json
