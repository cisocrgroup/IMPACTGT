# IMPACTGT
Scripts for IMPACT ground-truth generation.

## Setup
Install `go`, make sure that you have both python2 and python3
installed.
 * Make sure that `$HOME/go/bin` is in your `PATH`
 * install the `impgtt` helper tools: `cd impgtt && go install`
 * Install python3-venv if you are on debian or ubuntu
 * Install libtk

Use `make setup py2=my-python2 py3=my-python3` to setup the
tools. This will
 * install ocropus into the `env/2` virtual environment,
 * install calamari into the `env/3` virtual environment and
 * download the calamari OCR-models into the `model` folder.

If impgtt is for any reason not installed at its default location
`$HOME/go/bin` you can set it: `make setup py2=my-python2 py3=my-python3`.

## Files
There are various scripts in the `scripts` directory:
* `scripts/run.bash` runs the whole segmentation and alignment process
  (ie. runs the following three scripts in the right order)
* `scripts/segment.bash` segments the GT into regions and the regions
  into lines using `ocropus-nlbin`
* `scripts/predict.bash` runs the ocr-recognition (using
  `calamari-predict` and the lines
* `scripts/align.bash` algins the ocred lines with the ground-truth
  lines
* `impgtt/...` IMPACTGT-tools: helper tools for the scripts.

## Usage
General usage: `script/run.bash [-nobin] [-imgext EXT] IN [OUT]`

From this repositorie's root directory run the segmentation over the
data in the `IN` directory using `bash scripts/run.bash IN`.  The
result will be written to the `segmented/IN` directory.  You can use
the `-imgext EXT` option to set the extension of the input images,
i.e. `bash script/run.bash -imgext .sau.png IN` runs the segmentation
over all the `.sau.png` image files.

Currently it is not possible to run the scripts outside of this
repositorie's root directory. This is due to the fact that the
`run.bash` script assumes to find `ocorpus` and `calamari` installed
in the `env/2` and `env/3` directories.
