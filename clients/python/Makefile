SHELL = /bin/bash

VENV_PATH = .venv

.venv:
	python3 -m venv $(VENV_PATH)
	$(VENV_PATH)/bin/pip install -r requirements.txt

apidocs: .venv
	@$(VENV_PATH)/bin/pip install --editable .
	@$(VENV_PATH)/bin/pip install -U -r ./docs-requirements.txt
	@$(VENV_PATH)/bin/sphinx-build -W docs/ docs/_build
.PHONY: apidocs
