# Scripts for IMPACT ground-truth generation

## Setup
Install `go` and make shure that you have both python2 and python3
installed. Install `segregs` and `seglines`:
 * `go get https://github.com/finkf/segregs`
 * `go get https://github.com/finkf/seglines`

Use `make setup py2=my-python2 py3=my-python3` to setup the required
tools. This will
 * install ocropus into the `env/2` virtual environment,
 * install calamari into the `env/3` virtual environment and
 * download the calamari OCR-models into the `model` folder.

## Usage
Run the segmentation over the data in the `IN` directory using `bash
script/run.bash IN`.  The result will be written to the `segmented/IN`
directory.  You can use the `-imgext ext` option to set the extension
of the input images, i.e. `bash script/run.bash -imgext .sau.png IN`
runs the segmentation over all the `.sau.png` image files.