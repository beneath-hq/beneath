from collections.abc import Mapping
from typing import Iterable

from beneath import config
from beneath.consumer import ConsumerCallback
from beneath.client import Client
from beneath.pipeline import AsyncApplyFn, AsyncGenerateFn, Pipeline


_CLIENT: Client = None


def _get_client() -> Client:
    global _CLIENT
    if not _CLIENT:
        _CLIENT = Client()
    return _CLIENT


async def consume(
    stream_path: str,
    cb: ConsumerCallback,
    version: int = None,
    replay_only: bool = False,
    changes_only: bool = False,
    subscription_path: str = None,
    reset_subscription: bool = False,
    stop_when_idle: bool = False,
    max_concurrency: int = 1,
):
    client = _get_client()
    consumer = await client.consumer(
        stream_path=stream_path,
        version=version,
        subscription_path=subscription_path,
    )
    if subscription_path and reset_subscription:
        await consumer.reset()
    await consumer.subscribe(
        cb=cb,
        max_concurrency=max_concurrency,
        replay_only=replay_only,
        changes_only=changes_only,
        stop_when_idle=stop_when_idle,
    )


async def query_index(
    stream_path: str,
    version: int = None,
    filter: str = None,
    to_dataframe=True,
    max_bytes=config.DEFAULT_READ_ALL_MAX_BYTES,
    max_records=None,
    batch_size=config.DEFAULT_READ_BATCH_SIZE,
    warn_max=True,
) -> Iterable[Mapping]:
    client = _get_client()
    stream = await client.find_stream(stream_path)
    if version is None:
        instance = stream.primary_instance
        if not instance:
            raise Exception(f"stream {stream_path} doesn't have a primary instance")
    else:
        instance = await stream.find_instance(version)
    cursor = await instance.query_index(filter=filter)
    return await cursor.read_all(
        to_dataframe=to_dataframe,
        max_bytes=max_bytes,
        max_records=max_records,
        batch_size=batch_size,
        warn_max=warn_max,
    )


async def query_warehouse(
    sql: str,
    analyze: bool = False,
    max_bytes_scanned: int = config.DEFAULT_QUERY_WAREHOUSE_MAX_BYTES_SCANNED,
    to_dataframe=True,
    max_bytes=config.DEFAULT_READ_ALL_MAX_BYTES,
    max_records=None,
    batch_size=config.DEFAULT_READ_BATCH_SIZE,
    warn_max=True,
) -> Iterable[Mapping]:
    client = _get_client()
    job = await client.query_warehouse(
        query=sql, analyze=analyze, max_bytes_scanned=max_bytes_scanned
    )
    cursor = await job.get_cursor()
    return await cursor.read_all(
        to_dataframe=to_dataframe,
        max_bytes=max_bytes,
        max_records=max_records,
        batch_size=batch_size,
        warn_max=warn_max,
    )


def generate_stream_pipeline(
    generate_fn: AsyncGenerateFn,
    output_stream_path: str,
    output_stream_schema: str,
):
    p = Pipeline(parse_args=True)
    t1 = p.generate(generate_fn)
    p.write_stream(t1, stream_path=output_stream_path, schema=output_stream_schema)
    p.main()


def derive_stream_pipeline(
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


def consume_stream_pipeline(
    input_stream_path: str,
    consume_fn: AsyncApplyFn,
    max_concurrency: int = None,
):
    p = Pipeline(parse_args=True)
    t1 = p.read_stream(input_stream_path)
    p.apply(t1, fn=consume_fn, max_concurrency=max_concurrency)
    p.main()
