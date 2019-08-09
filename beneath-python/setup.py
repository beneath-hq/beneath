import setuptools

with open("README.md", "r") as fh:
    long_description = fh.read()

# TODO: input correct specifications
setuptools.setup(
    name="beneath-python-package",
    version=open(
        "beneath/_version.py").readlines()[-1].split()[-1].strip("\"'"),
    author="Benjamin Egelund-Muller and Eric Green",
    author_email="eric@beneath.network",
    description="Client package for Beneath Systems",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://beneath.network/",
    packages=setuptools.find_packages(),
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    scripts=['bin/beneathcmd'],
    install_requires=[
        'apache_beam',
        'argparse',
        'fastavro',
        'grpcio-tools',
        'pandas'
    ]
)
