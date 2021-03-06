# CONTRIBUTING

### Developing the library

- Make sure Python 3 is installed
- Install [poetry](https://python-poetry.org/docs/)
- Run `poetry install` (Poetry is configured to create and use a virtual environment)
- Run commands with `poetry run COMMAND` or start a virtual environment shell with `poetry shell`

### Publishing to PyPI

- Increment the version number in `pyproject.toml`
- Run `poetry publish --build`
- Update the config of recommended and deprecated versions in `services/data/clientversion/clientversion.go`

### Autogenerated API reference (docs)

We use [Sphinx](https://www.sphinx-doc.org/en/master/) to generate docs for this library based on the files in `docs` and the comments in the source files.

The autogenerated documentation is generated as HTML files in the `clients/python/docs/_build` folder (which is ignored by git). You can generate it locally by running `make docs` and opening `docs/_build/index.html` in your browser.

Every time changes are merged and pushed into the `stable` branch, an automatic build and deploy is triggered via Netlify (config in `netlify.toml`). The stable docs are published live at [https://python.docs.beneath.dev/](https://python.docs.beneath.dev/). Click the Netlify badge in `README.md` to see the build status.

### Resources

Here are some good resources to understand how Python packages work:

- [Packaging Python Projects](https://packaging.python.org/tutorials/packaging-projects/). It describes how to upload packages to PyPI.
- [Command Line Scripts](https://python-packaging.readthedocs.io/en/latest/command-line-scripts.html). It describes how to include command-line scripts in Python packages.
