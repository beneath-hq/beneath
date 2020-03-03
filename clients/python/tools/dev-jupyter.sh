#!/usr/bin/env bash

# check we're in the client repo
if [[ ! -f tools/$(basename $0) ]]; then
  echo "you must execute this script from the python client root folder (clients/python)"
  exit 1
fi

PYTHONPATH=$(pwd $0):$PYTHONPATH BENEATH_ENV=dev jupyter notebook
