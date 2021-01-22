import argparse
import sys

from beneath import __version__
from beneath.cli import auth
from beneath.cli import organization
from beneath.cli import project
from beneath.cli import service
from beneath.cli import stream


def main():
    parser = create_argument_parser()

    # print help if no args were provided
    if len(sys.argv) == 1:
        parser.print_help(sys.stderr)
        sys.exit(1)

    args = parser.parse_args()
    try:
        func = args.func
    except AttributeError:
        parser.error("too few arguments")
    func(args)


def create_argument_parser():
    parser = argparse.ArgumentParser(description="Beneath command line interface")
    parser.add_argument("-v", "--version", action="version", version=__version__)
    root = parser.add_subparsers()
    auth.add_subparser(root)
    organization.add_subparser(root)
    project.add_subparser(root)
    service.add_subparser(root)
    stream.add_subparser(root)
    return parser
