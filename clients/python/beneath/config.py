import os

DEV = os.environ.get("BENEATH_ENV") in ["dev", "development"]

BIGQUERY_PROJECT = "beneath"
PYTHON_CLIENT_ID = "beneath-python"

DEFAULT_READ_BATCH_SIZE = 1000
DEFAULT_READ_ALL_MAX_BYTES = 25 * 2 ** 20
DEFAULT_WRITE_DELAY_MS = 1000
DEFAULT_CHECKPOINT_COMMIT_DELAY_MS = 30000
DEFAULT_SUBSCRIBE_POLL_AT_LEAST_EVERY_MS = 60000
DEFAULT_SUBSCRIBE_POLL_AT_MOST_EVERY_MS = 1000
DEFAULT_QUERY_WAREHOUSE_TIMEOUT_MS = 60 * 1000
DEFAULT_QUERY_WAREHOUSE_MAX_BYTES_SCANNED = 10000000000
MAX_RECORD_SIZE_BYTES = 8192
MAX_BATCH_SIZE_BYTES = 10000000
MAX_BATCH_SIZE_COUNT = 10000

if DEV:
    BENEATH_FRONTEND_HOST = "http://localhost:3000"
    BENEATH_CONTROL_HOST = "http://localhost:4000"
    BENEATH_GATEWAY_HOST = "http://localhost:5000"
    BENEATH_GATEWAY_HOST_GRPC = "localhost:50051"
else:
    BENEATH_FRONTEND_HOST = "https://beneath.dev"
    BENEATH_CONTROL_HOST = "https://control.beneath.dev"
    BENEATH_GATEWAY_HOST = "https://data.beneath.dev"
    BENEATH_GATEWAY_HOST_GRPC = "grpc.data.beneath.dev"


def read_secret():
    with open(_secret_file_path(), "r") as f:
        return f.read()


def write_secret(secret):
    with open(_secret_file_path(), "w") as f:
        return f.write(secret)


def _secret_file_path():
    name = "secret_dev.txt" if DEV else "secret.txt"
    return os.path.join(_config_dir(), name)


def _config_dir():
    p = os.path.expanduser("~/.beneath")
    if not os.path.exists(p):
        os.makedirs(p)
    return p
