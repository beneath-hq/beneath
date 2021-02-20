# see https://github.com/python-poetry/poetry/pull/2366#issuecomment-652418094
try:
    import importlib.metadata as importlib_metadata
except ModuleNotFoundError:
    import importlib_metadata
__version__ = importlib_metadata.version(__name__)

from beneath.client import Client
from beneath.connection import AuthenticationError, GraphQLError
from beneath.checkpointer import Checkpointer
from beneath.consumer import Consumer
from beneath.cursor import Cursor
from beneath.easy import (
    easy_consume_stream,
    easy_derive_stream,
    easy_generate_stream,
    easy_read,
)
from beneath.instance import StreamInstance
from beneath.job import Job, JobStatus
from beneath.pipeline import Pipeline, PIPELINE_IDLE
from beneath.schema import Schema
from beneath.stream import Stream

__all__ = [
    "__version__",
    "AuthenticationError",
    "Checkpointer",
    "Client",
    "Consumer",
    "Cursor",
    "easy_consume_stream",
    "easy_derive_stream",
    "easy_generate_stream",
    "easy_read",
    "Job",
    "JobStatus",
    "Pipeline",
    "PIPELINE_IDLE",
    "GraphQLError",
    "Schema",
    "Stream",
    "StreamInstance",
]
