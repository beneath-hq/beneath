import argparse

from beneath._version import __version__
from beneath.cli import auth
from beneath.cli import model
from beneath.cli import organization
from beneath.cli import project
from beneath.cli import service
from beneath.cli import stream


def main():
  parser = create_argument_parser()
  args = parser.parse_args()
  try:
    func = args.func
  except AttributeError:
    parser.error("too few arguments")
  func(args)


def create_argument_parser():
  parser = argparse.ArgumentParser()
  parser.add_argument('-v', '--version', action='version', version=__version__)
  root = parser.add_subparsers()
  auth.add_subparser(root)
  model.add_subparser(root)
  organization.add_subparser(root)
  project.add_subparser(root)
  service.add_subparser(root)
  stream.add_subparser(root)
  return parser
