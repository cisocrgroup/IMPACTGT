py2 ?= python
py3 ?= python3
venv ?= virtualenv

setup: env/2/bin/ocropus-gpageseg env/3/bin/calamari-predict models/calamari_models-1.0/fraktur_historical/0.ckpt.json ${impgtt}

env/2/bin/activate:
	${venv} -p ${py2} env/2

# Install ocropus into the virtual environment.
env/2/bin/ocropus-gpageseg: env/2/bin/activate env/ocropy/setup.py
	. env/2/bin/activate && pip install -r env/ocropy/requirements.txt && deactivate
	. env/2/bin/activate && cd env/ocropy && python setup.py install && deactivate
	patch -u $@ patches/coordinates.patch
	patch -u env/2/bin/ocropus-nlbin patches/angle.patch

# Checkout ocropus git repository.
env/ocropy/setup.py:
	git clone --depth 1 -b master https://github.com/ocropus/ocropy env/ocropy

env/3/bin/activate:
	${py3} -m venv env/3

# Install calamari into the virtual environment.
env/3/bin/calamari-predict: env/3/bin/activate
	. env/3/bin/activate && pip install --upgrade pip && deactivate
	. env/3/bin/activate && pip install calamari_ocr && deactivate

# Download calamari models.
models/calamari.zip:
	mkdir -p models && wget "https://github.com/Calamari-OCR/calamari_models/archive/1.0.zip" -O $@

# Unpack models from calamari.zip.
models/calamari_models-1.0/fraktur_historical/0.ckpt.json: models/calamari.zip
	cd models && unzip calamari.zip calamari_models-1.0/fraktur_historical/*
	cd models && unzip calamari.zip calamari_models-1.0/antiqua_historical/*
	touch $@

clean:
	${RM} -r env models
.PHONY: clean
