#!/bin/bash
set -e

if [[ $# -lt 2 ]]; then
    echo "usage $0 DIR MODELS..."
    exit 1
fi

bdir=$(dirname $0)
dir=$1; shift
models=$@

# ocr
echo find $dir -type f -name '*.bin.png' "|" xargs calamari-predict -j 4 --checkpoint $models --files
find $dir -type f -name '*.bin.png' | xargs calamari-predict -j 4 --checkpoint $models --files

# align
echo java -Dfile.encoding=UTF8 -jar "$bdir/../lib/align_gt_ocr.jar" -f $dir
java -Dfile.encoding=UTF8 -jar "$bdir/../lib/align_gt_ocr.jar" -f $dir
