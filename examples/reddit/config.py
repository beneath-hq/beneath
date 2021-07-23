import logging
import os
from asyncpraw import Reddit
from asyncprawcore import Requestor
from tenacity import before_sleep_log, retry, stop_after_attempt, wait_chain, wait_fixed


def get_env(var, required) -> str:
    val = os.environ.get(var)
    if val is None:
        if required:
            raise RuntimeError(f"expected value for environment variable {var}")
        return ""
    return val


CLIENT_ID = get_env("REDDIT_CLIENT_ID", required=True)
CLIENT_SECRET = get_env("REDDIT_CLIENT_SECRET", required=True)
USERNAME = get_env("REDDIT_USERNAME", required=True)
PASSWORD = get_env("REDDIT_PASSWORD", required=False)
SUBREDDIT = get_env("REDDIT_SUBREDDIT", required=True)
USER_AGENT = f"Posts and comments 1.0 (by u/{USERNAME})"

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


def truncate_string(s):
    max_bytes = 5000
    if len(s.encode("utf-8")) > max_bytes:
        return s.encode("utf-8")[:max_bytes].decode("utf-8", "ignore") + " [TRUNCATED]"
    return s
