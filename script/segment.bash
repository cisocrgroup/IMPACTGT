#!/bin/bash
set -e

dobin=true
imgext=".png"
while true; do
    case $1 in
	-nobin)
	    dobin=false
	    shift;;
	-imgext)
	    shift
	    imgext=$1
	    shift;;
	*)
	    break;;
    esac
done

if [[ $# != 2 ]]; then
    echo "usage $0 [-nobin|-imgext ext] IN OUT"
    exit 1
fi

if [[ "$1" == "$2" ]]; then
    echo "error: IN and OUT are the same"
    exit 1
fi

bdir=$1
odir=$2
rm -rf $odir
mkdir -p $odir

# # Binarize the images.
# echo $dobin
# if [[ $dobin == true ]]; then
#     for tif in $bdir/*.tif; do
# 	ocropus-nlbin $tif
#     done
# fi

# Copy the image files and their segments.
for img in $bdir/*$imgext; do
    xml=${img%$imgext}.xml
    base=$(basename $img)
    echo ln $img $odir/$base || cp $img $odir/$base
    ln $img $odir/$base || cp $img $odir/$base
    echo ln $xml $odir/${base%$imgext}.xml || cp $xml $odir/${base%$imgext}.xml
    ln $xml $odir/${base%$imgext}.xml || cp $xml $odir/${base%$imgext}.xml
    echo segregs -padding 10 $xml $img $odir/${base%$imgext}
    segregs -padding 10 $xml $img $odir/${base%$imgext}
done

# Segment the regions into lines.
for json in $odir/*.json; do
    echo ocropus-gpageseg -n ${json/%.json/.png}
    ocropus-gpageseg -n ${json/%.json/.png}
    echo seglines $json
    seglines $json
done
