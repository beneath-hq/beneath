from datetime import datetime, timezone
import pandas as pd

from beneath.utils.aiodelaybuffer import AIODelayBuffer
from beneath.utils.aiopoller import AIOPoller
from beneath.utils.aioticker import AIOTicker
from beneath.utils.aioworkerpool import AIOWorkerPool
from beneath.utils.infer_avro import infer_avro
from beneath.utils.qualifiers import (
    pretty_entity_name,
    ProjectQualifier,
    ServiceQualifier,
    StreamQualifier,
    SubscriptionQualifier,
)

__all__ = [
    "AIODelayBuffer",
    "AIOPoller",
    "AIOTicker",
    "AIOWorkerPool",
    "ServiceQualifier",
    "StreamQualifier",
    "SubscriptionQualifier",
    "ProjectQualifier",
    "pretty_entity_name",
    "infer_avro",
]


def format_entity_name(name):
    return name.replace("-", "_").lower()


def ms_to_datetime(ms):
    return datetime.fromtimestamp(float(ms) / 1000.0, timezone.utc)


def ms_to_pd_timestamp(ms):
    return pd.Timestamp(ms, unit="ms")


def timestamp_to_ms(timestamp):
    if isinstance(timestamp, datetime):
        return datetime_to_ms(timestamp)
    if not isinstance(timestamp, int):
        raise TypeError("couldn't parse {} as a timestamp".format(timestamp))
    return timestamp


def datetime_to_ms(dt):
    return int(dt.replace(tzinfo=timezone.utc).timestamp() * 1000)


def format_graphql_time(dt):
    return dt.isoformat() + "Z"
