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
    consume,
    load_full,
    query_warehouse,
    write_full,
    generate_stream_pipeline,
    derive_stream_pipeline,
    consume_stream_pipeline,
)
from beneath.instance import StreamInstance
from beneath.job import Job, JobStatus
from beneath.pipeline import Action, Pipeline, PIPELINE_IDLE, Strategy
from beneath.schema import Schema
from beneath.stream import Stream

__all__ = [
    "__version__",
    "Action",
    "AuthenticationError",
    "Checkpointer",
    "Client",
    "Consumer",
    "Cursor",
    "Job",
    "JobStatus",
    "Pipeline",
    "PIPELINE_IDLE",
    "GraphQLError",
    "Schema",
    "Strategy",
    "Stream",
    "StreamInstance",
    "consume",
    "load_full",
    "query_warehouse",
    "write_full",
    "generate_stream_pipeline",
    "derive_stream_pipeline",
    "consume_stream_pipeline",
]
