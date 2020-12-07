# see https://github.com/python-poetry/poetry/pull/2366#issuecomment-652418094
try:
    import importlib.metadata as importlib_metadata
except ModuleNotFoundError:
    import importlib_metadata
__version__ = importlib_metadata.version(__name__)

from beneath.client import Client
from beneath.connection import AuthenticationError, GraphQLError
from beneath.easy import (
    easy_consume_stream,
    easy_derive_stream,
    easy_generate_stream,
    easy_read,
)
from beneath.logging import logger
from beneath.pipeline import Pipeline, PIPELINE_IDLE

__all__ = [
    "__version__",
    "AuthenticationError",
    "Client",
    "easy_consume_stream",
    "easy_derive_stream",
    "easy_generate_stream",
    "easy_read",
    "logger",
    "Pipeline",
    "PIPELINE_IDLE",
    "GraphQLError",
]
