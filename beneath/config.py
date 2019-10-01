import os


BIGQUERY_PROJECT = "beneath"
PYTHON_CLIENT_ID = "beneath-python"

MAX_READ_MB = 10
READ_BATCH_SIZE = 1000

if os.environ.get('ENV') == 'development':
  BENEATH_FRONTEND_HOST = "http://localhost:3000"
  BENEATH_CONTROL_HOST = "http://localhost:4000"
  BENEATH_GATEWAY_HOST = "http://localhost:5000"
  BENEATH_GATEWAY_HOST_GRPC = "localhost:50051"
else:
  BENEATH_FRONTEND_HOST = "https://beneath.network"
  BENEATH_CONTROL_HOST = "https://control.beneath.network"
  BENEATH_GATEWAY_HOST = "https://data.beneath.network"
  BENEATH_GATEWAY_HOST_GRPC = "data.beneath.network"


def read_secret():
  with open(_secret_file_path(), "r") as f:
    return f.read()


def write_secret(secret):
  with open(_secret_file_path(), "w") as f:
    return f.write(secret)


def _secret_file_path():
  return os.path.join(_config_dir(), "secret.txt")


def _config_dir():
  p = os.path.expanduser("~/.beneath")
  if not os.path.exists(p):
    os.makedirs(p)
  return p
