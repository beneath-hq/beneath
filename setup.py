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
  author_email="hello@beneath.network",
  description="Python client for Beneath (https://beneath.network/)",
  url="https://gitlab.com/_beneath/beneath-python",
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
    'apache_beam[gcp]==2.16.0',
    'argparse==1.4',
    'Cython==0.29.13',
    'fastavro==0.21.24',
    'grpcio==1.24.1',
    'pandas==0.25.1',
    'six==1.12.0',
  ]
)
