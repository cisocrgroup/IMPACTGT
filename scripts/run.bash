#!/bin/bash
set -e

nobin=""
imgext=".png"
bdir=$(dirname $0)

while true; do
    case $1 in
	-nobin)
	    nobin="-nobin"
	    shift;;
	-imgext)
	    shift
	    imgext=$1
	    shift;;
	*)
	    break;;
    esac
done

if [[ $# == 1 ]]; then
    idir=$1
    odir=segmented/$(basename $idir)
elif [[ $# == 2 ]]; then
     idir=$1
     odir=$2
else
	echo "usage: $0 [-nobin] [-imgext EXT] IN [OUT]"
	exit 1
fi

# Prepare output directory.
mkdir -p $odir

# Segment (needs ocorpus).
source $bdir/../env/2/bin/activate
echo $bdir/segment.bash $nobin -imgext $imgext $idir $odir
$bdir/segment.bash $nobin -imgext $imgext $idir $odir
deactivate

# Predict (needs calamari).
source $bdir/../env/3/bin/activate
$bdir/predict.bash $odir \
	$bdir/../models/calamari_models-1.0/fraktur_historical/3.ckpt \
	$bdir/../models/calamari_models-1.0/fraktur_historical/4.ckpt \
	$bdir/../models/calamari_models-1.0/antiqua_historical/3.ckpt \
	$bdir/../models/calamari_models-1.0/antiqua_historical/4.ckpt
deactivate

# Align lines
$bdir/align.bash $odir
# cleanup
# $bdir/cleanup.bash $odir
