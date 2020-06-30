from collections.abc import Mapping
from typing import Iterable

from beneath import config
from beneath.client import Client

_CLIENT: Client = None


def _get_client() -> Client:
  # pylint: disable=global-statement
  global _CLIENT
  if not _CLIENT:
    _CLIENT = Client()
  return _CLIENT


async def easy_read(
  stream_path: str,
  # pylint: disable=redefined-builtin
  filter: str = None,
  to_dataframe=True,
  max_bytes=config.DEFAULT_READ_ALL_MAX_BYTES,
  max_records=None,
  batch_size=config.DEFAULT_READ_BATCH_SIZE,
  warn_max=True,
) -> Iterable[Mapping]:
  client = _get_client()
  stream = await client.find_stream(stream_path)
  instance = stream.primary_instance
  if not instance:
    raise Exception(f"stream {stream_path} doesn't have a primary instance")
  cursor = await instance.query_index(filter=filter)
  return await cursor.read_all(
    to_dataframe=to_dataframe,
    max_bytes=max_bytes,
    max_records=max_records,
    batch_size=batch_size,
    warn_max=warn_max,
  )

