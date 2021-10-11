#!/bin/bash
set -e

if [[ $# -lt 2 ]]; then
    echo "usage $0 DIR MODELS..."
    exit 1
fi

dir=$1; shift
models=$@

# ocr
echo find $dir -type f -name '*.bin.png' '|' xargs calamari-predict -j 4 --checkpoint $models --files
find $dir -type f -name '*.bin.png' | xargs calamari-predict -j 4 --checkpoint $models --files
