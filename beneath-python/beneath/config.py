import os

BENEATH_FRONTEND_HOST = "http://localhost:3000"
BENEATH_CONTROL_HOST = "http://localhost:4000"
BENEATH_GATEWAY_HOST = "http://localhost:5000"
BIGQUERY_PROJECT = "beneathcrypto"

def read_secret():
  with open(_secret_file_path(), "r") as f:
    return f.read()


def write_secret(SECRET):
  with open(_secret_file_path(), "w+") as f:
    return f.write(SECRET)


def _config_dir():
  p = os.path.expanduser("~/.beneath")
  if not os.path.exists(p):
    os.makedirs(p)
  return p


def _secret_file_path():
  return os.path.join(_config_dir(), "secret.txt")
