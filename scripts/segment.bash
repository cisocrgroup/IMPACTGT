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
    echo "usage $0 [-nobin|-imgext EXT] IN OUT"
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
# if [[ $dobin == true ]]; then
#     for tif in $bdir/*.tif; do
# 	ocropus-nlbin $tif
#     done
# fi

# Copy the xml and image files.
for xml in $bdir/*.xml; do
    img=${xml%.xml}$imgext
    base=$(basename $img)
    echo ln $img $odir/$base '||' cp $img $odir/$base
    ln $img $odir/$base || cp $img $odir/$base
    echo ln $xml $odir/${base%$imgext}.xml '||' cp $xml $odir/${base%$imgext}.xml
    ln $xml $odir/${base%$imgext}.xml || cp $xml $odir/${base%$imgext}.xml
    echo impgtt segregs --padding 10 $xml $img $odir/${base%$imgext}
    impgtt segregs --padding 10 $xml $img $odir/${base%$imgext}
done

# Segment the regions into lines.
for json in $odir/*.json; do
    echo ocropus-nlbin -n -Q4  ${json/%.json/.png}
    ocropus-nlbin -n -Q4 ${json/%.json/.png}
    echo ocropus-gpageseg -n --maxcolseps 0 --csminheight 100000 --usegauss -Q4 ${json/%.json/.bin.png}
    ocropus-gpageseg -n --maxcolseps 0 --csminheight 100000 --usegauss -Q 4 ${json/%.json/.bin.png}
    echo impgtt seglines $json
    impgtt seglines $json
done
