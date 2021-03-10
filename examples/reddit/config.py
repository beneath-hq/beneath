import logging
import os
import sys
from asyncpraw import Reddit
from asyncprawcore import Requestor
from tenacity import before_sleep_log, retry, stop_after_attempt, wait_chain, wait_fixed


def get_env(var) -> str:
    val = os.environ.get(var)
    if val is None:
        raise RuntimeError(f"expected value for environment variable {var}")
    return val


USER_AGENT = get_env("REDDIT_USER_AGENT")
CLIENT_ID = get_env("REDDIT_CLIENT_ID")
CLIENT_SECRET = get_env("REDDIT_CLIENT_SECRET")
USERNAME = get_env("REDDIT_USERNAME")
PASSWORD = get_env("REDDIT_PASSWORD")
SUBREDDIT = get_env("REDDIT_SUBREDDIT")

logger = logging.getLogger("redditscraper")
logger.setLevel(logging.WARNING)


class RetryRequestor(Requestor):
    @retry(
        before_sleep=before_sleep_log(logger, logging.WARNING),
        reraise=True,
        stop=stop_after_attempt(5),
        wait=wait_chain(
            *[wait_fixed(0.5)] + [wait_fixed(2)] + [wait_fixed(10)] + [wait_fixed(60)]
        ),
    )
    async def request(self, *args, **kwargs):
        return await super().request(*args, **kwargs)


reddit = Reddit(
    user_agent=USER_AGENT,
    client_id=CLIENT_ID,
    client_secret=CLIENT_SECRET,
    username=USERNAME,
    password=PASSWORD,
    requestor_class=RetryRequestor,
)


def truncate_text(text):
    max_chars = 5000
    if len(text) > max_chars:
        return text[:max_chars] + " [TRUNCATED]"
    return text
