#!/usr/bin/env bash

# 1. Upgrade setuptools and wheel
python3 -m pip install --upgrade setuptools wheel twine

# 2. generate distribution archives
python3 setup.py sdist bdist_wheel

# 3. upload the distribution archives
python3 -m twine upload dist/*
