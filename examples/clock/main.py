import beneath
import math
from datetime import timedelta

from config import FREQUENCY
from generators import clock

with open("schemas/tick.gql", "r") as file:
    SCHEMA = file.read()


def make_stream_suffix(seconds):
    td = timedelta(seconds=float(seconds))
    days = td.days
    hours = td.seconds // 3600
    minutes = (td.seconds // 60) % 60
    seconds = td.seconds - hours * 3600 - minutes * 60
    return f"{str(days) + 'd' if days > 0 else ''}\
{str(hours) + 'hr' if hours > 0 else ''}\
{str(minutes) + 'm' if minutes > 0 else ''}\
{str(seconds) + 's' if seconds > 0 else ''}"


if __name__ == "__main__":
    beneath.easy_generate_stream(
        generate_fn=clock.ticker,
        output_stream_path="examples/clock/tick-" + make_stream_suffix(FREQUENCY),
        output_stream_schema=SCHEMA,
    )
