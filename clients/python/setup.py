import setuptools

# Get long_description
with open('README.md', 'r') as f:
  LONG_DESCRIPTION = f.read()

# Get version
with open('beneath/_version.py', 'r') as f:
  for line in f:
    if line.startswith('__version__'):
      _, _, VERSION = line.replace('"', '').split()
      break

setuptools.setup(
  name="beneath",
  version=VERSION,
  author="Beneath Systems",
  author_email="hello@beneath.dev",
  description="Python client for Beneath (https://beneath.dev/)",
  url="https://gitlab.com/_beneath/beneath-core/-/tree/master/clients/python",
  long_description=LONG_DESCRIPTION,
  long_description_content_type='text/markdown',
  packages=setuptools.find_packages(),
  classifiers=[
    "Programming Language :: Python :: 3",
    "License :: OSI Approved :: MIT License",
    "Operating System :: OS Independent",
  ],
  scripts=['beneath/bin/beneath'],
  install_requires=[
    'aiogrpc>=1.7',
    'aiohttp>=3.6.2',
    'argparse==1.4',
    'Cython==0.29.15',
    'fastavro<0.22,>=0.21.4',
    'grpcio==1.27.2',
    'pandas==1.0.1',
    'six==1.14.0',
  ],
  extras_require={
    'beam': ['apache_beam[gcp]==2.19.0'],
  },
)
