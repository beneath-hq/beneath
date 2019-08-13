#!/usr/bin/env bash

# 1. generate distribution archives
python3 setup.py sdist bdist_wheel

# 2. upload the distribution archives
python3 -m twine upload --repository-url https://test.pypi.org/legacy/ dist/*
