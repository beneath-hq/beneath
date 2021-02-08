import os
import asyncpraw


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

reddit = asyncpraw.Reddit(
    user_agent=USER_AGENT,
    client_id=CLIENT_ID,
    client_secret=CLIENT_SECRET,
    username=USERNAME,
    password=PASSWORD,
)

MAX_CHARACTERS = 5000