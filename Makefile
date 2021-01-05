py2 ?= python
py3 ?= python3

setup: env/2/bin/ocropus-dewarp env/3/bin/calamari-predict model/calamari_models-1.0/fraktur_historical/0.ckpt.json lib/align_gt_ocr.jar

env/2/bin/activate:
	virtualenv -p ${py2} env/2

# Install ocropus into the virtual environment.
env/2/bin/ocropus-dewarp: env/2/bin/activate env/ocropy/setup.py
	source env/2/bin/activate && pip install -r env/ocropy/requirements.txt && deactivate
	source env/2/bin/activate && cd env/ocropy && python setup.py install && deactivate

# Checkout ocropus git repository.
env/ocropy/setup.py:
	git clone --depth 1 https://github.com/ocropus/ocropy env/ocropy

env/3/bin/activate:
	${py3} -m venv env/3

# Install calamari into the virtual environment.
env/3/bin/calamari-predict: env/3/bin/activate
	source env/3/bin/activate && pip install --upgrade pip && deactivate
	source env/3/bin/activate && pip install 'h5py<3' 'tensorflow>=2' calamari_ocr && deactivate

# Download calamari models.
model/calamari.zip:
	mkdir -p model && wget "https://github.com/Calamari-OCR/calamari_models/archive/1.0.zip" -O $@

# Unpack models from calamari.zip.
model/calamari_models-1.0/fraktur_historical/0.ckpt.json: model/calamari.zip
	cd model && unzip calamari.zip calamari_models-1.0/fraktur_historical/*
	cd model && unzip calamari.zip calamari_models-1.0/antiqua_historical/*
	touch $@

# Unpack jar.
%.jar: %.zip
	unzip -p $^ > $@

# Download jar.
lib/align_gt_ocr.zip:
	mkdir -p lib && cd lib && wget http://cis.lmu.de/~finkf/align_gt_ocr.zip
