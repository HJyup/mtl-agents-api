.PHONY: generate-protos run install test clean

PYTHON = python3
PIP = pip3

SERVICE_NAME = agent

install:
	$(PIP) install -r requirements.txt

run:
	$(PYTHON) -m server

test:
	$(PYTHON) -m unittest discover -s tests

clean:
	rm -rf __pycache__
	rm -rf */__pycache__
	rm -rf */*/__pycache__
	find . -name "*.pyc" -delete 