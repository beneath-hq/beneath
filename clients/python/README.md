# Beneath Python Client Library

This repository contains the source code for the [Beneath](https://beneath.dev) Python library. Learn more about it with these resources:

- [Overview](https://about.beneath.dev/docs/overview/)
- [Reading data from Beneath](https://about.beneath.dev/docs/read-data-into-jupyter-notebook/)
- [Writing data to Beneath](https://about.beneath.dev/docs/write-data-from-your-app/)

### Providing feedback

Beneath is just entering public beta, so there's bound to be some rough edges. Bugs, feature requests, suggestions – we'd love to hear about them. To file an issue, [click here](https://gitlab.com/_beneath/beneath-python/issues).

### Developing the library

- Make sure Python 3 is installed and available as `python3`
- Initialize and source the Python virtual environment with:

    python3 -m venv .venv
    source .venv/bin/activate
    pip install -r requirements.txt

- Run tests with `pytest` (though it's sparse on tests at moment)
- Run `deactivate` to exit the virtual environment and `source .env/bin/activate` to re-activate it
- For use in VS Code, open `python` as a workspace root folder. Press `CMD+Shift+P`, search for `Python: Select Interpreter`, and select the Python 3 executable in the `.venv` virtual environment.

Here are some good resources to understand how Python packages work:

- [Packaging Python Projects](https://packaging.python.org/tutorials/packaging-projects/). It describes how to upload packages to PyPI.
- [Command Line Scripts](https://python-packaging.readthedocs.io/en/latest/command-line-scripts.html). It describes how to include command-line scripts in Python packages.

### Publishing to PyPI

- Increment the version number in `beneath/_version.py`
- Run `./tools/pypi-publish.sh`
- Make sure to appropriately update configuration of recommended and deprecated versions in `beneath-core/gateway/grpc/client_version.go`
