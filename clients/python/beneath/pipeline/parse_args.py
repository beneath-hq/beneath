import argparse
import os

from beneath.cli.utils import mb_to_bytes


def parse_pipeline_args():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "action",
        type=str,
        help="the action to execute: one of 'run', 'test', 'stage' or 'teardown'",
    )
    parser.add_argument(
        "service_path",
        type=str,
        help="path for the pipeline ('username/project/name')",
    )
    parser.add_argument(
        "--strategy",
        type=str,
        default="continuous",
        help="the run strategy: one of 'continuous', 'delta' or 'batch'",
    )
    parser.add_argument(
        "--version",
        type=str,
        default="0",
        help="the version number to use for the output streams (defaults to 0)",
    )
    parser.add_argument(
        "--read-quota-mb",
        type=mb_to_bytes,
        default=None,
        help="sets a limit on the pipeline's monthly read quota (set to 0 for unlimited quota)",
    )
    parser.add_argument(
        "--write-quota-mb",
        type=mb_to_bytes,
        default=None,
        help="sets a limit on the pipeline's monthly write quota (set to 0 for unlimited quota)",
    )
    parser.add_argument(
        "--scan-quota-mb",
        type=mb_to_bytes,
        default=None,
        help="sets a limit on the pipeline's monthly scan quota (set to 0 for unlimited quota)",
    )

    args = parser.parse_args()

    os.environ["BENEATH_PIPELINE_ACTION"] = args.action
    os.environ["BENEATH_PIPELINE_STRATEGY"] = args.strategy
    os.environ["BENEATH_PIPELINE_VERSION"] = args.version
    os.environ["BENEATH_PIPELINE_SERVICE_PATH"] = args.service_path

    if args.read_quota_mb is not None:
        os.environ["BENEATH_PIPELINE_SERVICE_READ_QUOTA"] = str(args.read_quota_mb)
    if args.write_quota_mb is not None:
        os.environ["BENEATH_PIPELINE_SERVICE_WRITE_QUOTA"] = str(args.write_quota_mb)
    if args.scan_quota_mb is not None:
        os.environ["BENEATH_PIPELINE_SERVICE_SCAN_QUOTA"] = str(args.scan_quota_mb)
