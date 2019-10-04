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
  name="beneath-python",
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
  scripts=['bin/beneath'],
  install_requires=[
    'apache_beam[gpc]>=2.14.0',
    'argparse>=1.1',
    'fastavro>=0.21.24',
    'grpcio>=1.23.0',
    'pandas>=0.24.2',
    'six>=1.12.0',
  ]
)
