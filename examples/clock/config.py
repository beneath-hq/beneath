import os


def get_env(var) -> str:
    val = os.environ.get(var)
    if val is None:
        raise RuntimeError(f"expected value for environment variable {var}")
    return val


FREQUENCY = get_env("FREQUENCY")
START_TIME = get_env("START_TIME")
