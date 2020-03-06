from beneath.utils.aiobuffer import AIOWindowedBuffer
from beneath.utils.aiopoller import AIOPoller
from beneath.utils.aioworkerpool import AIOWorkerPool

from datetime import datetime, timezone
import pandas as pd


def format_entity_name(name):
  return name.replace("-", "_").lower()


def ms_to_datetime(ms):
  return datetime.fromtimestamp(float(ms) / 1000., timezone.utc)


def ms_to_pd_timestamp(ms):
  return pd.Timestamp(ms, unit='ms')


def timestamp_to_ms(timestamp):
  if isinstance(timestamp, datetime):
    return datetime_to_ms(timestamp)
  if not isinstance(timestamp, int):
    raise TypeError("couldn't parse {} as a timestamp".format(timestamp))
  return timestamp


def datetime_to_ms(dt):
  return int(dt.replace(tzinfo=timezone.utc).timestamp() * 1000)


def format_graphql_time(dt):
  return dt.isoformat() + 'Z'
