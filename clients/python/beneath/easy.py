from collections.abc import Mapping
from typing import Iterable, List, Union

import pandas as pd

from beneath import config
from beneath.consumer import ConsumerCallback
from beneath.client import Client
from beneath.job import Job
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
    """
    Shorthand for creating a client, creating a stream consumer, replaying the stream's history and
    subscribing to changes.

    Args:
        stream_path (str):
            Path to the stream to subscribe to. The consumer will subscribe to the stream's
            primary version.
        cb (async def fn(record)):
            Async function for processing a record.
        version (int):
            The instance version to use for stream. If not set, uses the primary instance.
        replay_only (bool):
            If true, will not read changes, but only replay historical records.
            Defaults to False.
        changes_only (bool):
            If true, will not replay historical records, but only subscribe to new changes.
            Defaults to False.
        subscription_path (str):
            Format "ORGANIZATION/PROJECT/NAME". If set, the consumer will use a checkpointer
            to save cursors. That means processing will not restart from scratch if the process
            ends or crashes (as long as you use the same subscription name). To reset a
            subscription, call ``reset`` on the consumer.
        reset_subscription (bool):
            If true and ``subscription_path`` is set, will reset the consumer and start the
            subscription from scratch.
        stop_when_idle (bool):
            If true, will return when "caught up" and no new changes are available.
            Defaults to False.
        max_concurrency (int):
            The maximum number of callbacks to call concurrently. Defaults to 1.
    """
    client = _get_client()
    if subscription_path is not None:
        await client.start()
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
    if subscription_path is not None:
        await client.start()


async def load_full(
    stream_path: str,
    version: int = None,
    filter: str = None,
    to_dataframe=True,
    max_bytes=config.DEFAULT_READ_ALL_MAX_BYTES,
    max_records=None,
    batch_size=config.DEFAULT_READ_BATCH_SIZE,
    warn_max=True,
) -> Iterable[Mapping]:
    """
    Shorthand for creating a client, finding a stream, and reading (filtered) records from it.

    Args:
        stream_path (str):
            The path to the stream in the format of "USERNAME/PROJECT/STREAM"
        version (int):
            The instance version to read from. If not set, defaults to the stream's primary
            instance.
        filter (str):
            A filter to apply to the index. Filters allow you to quickly
            find specific record(s) in the index based on the record key.
            For details on the filter syntax,
            see https://about.beneath.dev/docs/reading-writing-data/index-filters/.
        to_dataframe (bool):
            If true, will return the result as a Pandas dataframe. Defaults to true.
        max_bytes (int)
            Sets the maximum number of bytes to read before returning with a warning.
            Defaults to 25 MB (Avro-encoded).
        max_records (int):
            Sets the maximum number of records to read before returning with a warning.
            Defaults to unlimited (see ``max_bytes``).
        batch_size (int):
            Sets the number of records to fetch in each network request. Defaults to 1000.
            One call to ``query_index`` may make many network requests (until all records
            have been loaded or ``max_bytes`` or ``max_records`` is breached).
        warn_max (bool):
            If true, will emit a warning if ``max_bytes`` or ``max_records`` were breached and
            the function returned without loading the full result. Defaults to true.
    """
    client = _get_client()
    stream = await client.find_stream(stream_path)
    if version is None:
        instance = stream.primary_instance
        if not instance:
            raise Exception(f"stream {stream_path} doesn't have a primary instance")
    else:
        instance = await stream.find_instance(version)
    if stream.use_index:
        cursor = await instance.query_index(filter=filter)
    elif filter is not None:
        raise ValueError(
            f"cannot use filter for {stream_path} because it doesn't have indexing enabled"
        )
    else:
        cursor = await instance.query_log()
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
) -> Union[Iterable[Mapping], Job]:
    """
    Shorthand for creating a client, running a warehouse (OLAP) query, and reading the result.
    If analyze=True, the analyzed job is returned instead of the result.

    Args:
        sql (str):
            The analytical SQL query to run. To learn about the query language,
            see https://about.beneath.dev/docs/reading-writing-data/warehouse-queries/.
        analyze (bool):
            If true, analyzes the query and returns info about referenced streams
            and expected bytes scanned, but doesn't actually run the query.
        max_bytes_scanned (int):
            Sets a limit on the number of bytes the query can scan.
            If exceeded, the job will fail with an error.
            Defaults to 10GB.
        to_dataframe (bool):
            If true, will return the result as a Pandas dataframe. Defaults to true.
        max_bytes (int)
            Sets the maximum number of bytes to read before returning with a warning.
            Defaults to 25 MB (Avro-encoded).
        max_records (int):
            Sets the maximum number of records to read before returning with a warning.
            Defaults to unlimited (see ``max_bytes``).
        batch_size (int):
            Sets the number of records to fetch in each network request. Defaults to 1000.
            One call to ``query_index`` may make many network requests (until all records
            have been loaded or ``max_bytes`` or ``max_records`` is breached).
        warn_max (bool):
            If true, will emit a warning if ``max_bytes`` or ``max_records`` were breached and
            the function returned without loading the full result. Defaults to true.
    """
    client = _get_client()
    job = await client.query_warehouse(
        query=sql, analyze=analyze, max_bytes_scanned=max_bytes_scanned
    )
    if analyze:
        return job
    cursor = await job.get_cursor()
    return await cursor.read_all(
        to_dataframe=to_dataframe,
        max_bytes=max_bytes,
        max_records=max_records,
        batch_size=batch_size,
        warn_max=warn_max,
    )


async def write_full(
    stream_path: str,
    records: Union[Iterable[dict], pd.DataFrame],
    key: Union[str, List[str]] = None,
    description: str = None,
    recreate_on_schema_change=False,
):
    """
    Shorthand for creating a client, inferring a stream schema, creating a stream, and writing
    data to the stream. It wraps ``Client.write_full``. Calling ``write_full`` multiple times
    for the same stream, will create and write to multiple versions, but will not append data
    an existing version.

    Args:
        stream_path (str):
            The (desired) path to the stream in the format of "USERNAME/PROJECT/STREAM".
            The project must already exist.
        records (list(dict) | pandas.DataFrame):
            The full dataset to write, either as a list of records or as a Pandas DataFrame.
            This function uses ``beneath.infer_avro`` to infer a schema for the stream based
            on the records.
        key (str | list(str)):
            The fields to use as the stream's key. If not set, will default to the dataframe
            index if ``records`` is a Pandas DataFrame, or add a column of incrementing numbers
            if ``records`` is a list.
        description (str):
            A description for the stream.
        recreate_on_schema_change (bool):
            If true, and there's an existing stream at ``stream_path`` with a schema that is
            incompatible with the inferred schema for ``records``, it will delete the existing
            stream and create a new one instead of throwing an error. Defaults to false.
    """
    client = _get_client()
    await client.start()
    await client.write_full(
        stream_path=stream_path,
        records=records,
        key=key,
        description=description,
        recreate_on_schema_change=recreate_on_schema_change,
    )
    await client.stop()


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
