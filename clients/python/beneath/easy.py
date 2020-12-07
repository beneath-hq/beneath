from collections.abc import Mapping
from typing import Iterable

from beneath import config
from beneath.client import Client
from beneath.pipeline import AsyncApplyFn, AsyncGenerateFn, Pipeline

_CLIENT: Client = None


def _get_client() -> Client:
    global _CLIENT
    if not _CLIENT:
        _CLIENT = Client()
    return _CLIENT


async def easy_read(
    stream_path: str,
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


def easy_generate_stream(
    generate_fn: AsyncGenerateFn, output_stream_path: str, output_stream_schema: str
):
    p = Pipeline(parse_args=True)
    t1 = p.generate(generate_fn)
    p.write_stream(t1, stream_path=output_stream_path, schema=output_stream_schema)
    p.main()


def easy_derive_stream(
    input_stream_path: str,
    apply_fn: AsyncApplyFn,
    output_stream_path: str,
    output_stream_schema: str,
    max_concurrency: int = None,
):
    p = Pipeline(parse_args=True)
    t1 = p.read_stream(input_stream_path)
    t2 = p.apply(t1, fn=apply_fn, max_concurrency=max_concurrency)
    p.write_stream(t2, stream_path=output_stream_path, schema=output_stream_schema)
    p.main()


def easy_consume_stream(
    input_stream_path: str, consume_fn: AsyncApplyFn, max_concurrency: int = None
):
    p = Pipeline(parse_args=True)
    t1 = p.read_stream(input_stream_path)
    p.apply(t1, fn=consume_fn, max_concurrency=max_concurrency)
    p.main()
