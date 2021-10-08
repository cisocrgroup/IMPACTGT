# IMPACTGT
Scripts for IMPACT ground-truth generation.

## Setup
Install `go`, make sure that you have both python2 and python3 and
java installed. Install `segregs`, `seglines` and `alignes`:
 * `go install github.com/cisocrgroup/segregs@latest`
 * `go install github.com/finkf/seglines@latest`
 * `go install github.com/finkf/alignes@latest`
 * Add `$HOME/go/bin` to your `PATH`
 * Install python3-venv if you are on debian or ubuntu
 * Install libtk

Use `make setup py2=my-python2 py3=my-python3` to setup the
tools. This will
 * install ocropus into the `env/2` virtual environment,
 * install calamari into the `env/3` virtual environment and
 * download the calamari OCR-models into the `model` folder.

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

## Usage
General usage: `script/run.bash [-nobin] [-imgext EXT] IN [OUT]`

Run the segmentation over the data in the `IN` directory using `bash
scripts/run.bash IN`.  The result will be written to the `segmented/IN`
directory.  You can use the `-imgext EXT` option to set the extension
of the input images, i.e. `bash script/run.bash -imgext .sau.png IN`
runs the segmentation over all the `.sau.png` image files.
