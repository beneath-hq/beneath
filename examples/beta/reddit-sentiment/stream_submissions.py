import beneath
import os
import pytz
import asyncio
import praw
from datetime import datetime
from structlog import get_logger
from textblob import TextBlob

TABLE = "epg/reddit/coronavirus-posts-sentiment"
REDDIT_USER_AGENT = "BeneathClient/0.1 (by /u/greenep12)"
REDDIT_CLIENT_ID = "hQE2QCQmnPhqTg"
REDDIT_CLIENT_SECRET = os.getenv("REDDIT_CLIENT_SECRET", default=None)
SUBREDDIT = "Coronavirus"

log = get_logger()

async def main():
  client = beneath.Client()
  stream = await client.find_stream(TABLE)

  reddit = praw.Reddit(
    user_agent=REDDIT_USER_AGENT,
    client_id=REDDIT_CLIENT_ID,
    client_secret=REDDIT_CLIENT_SECRET,
  )
  
  subreddit = reddit.subreddit(SUBREDDIT)
  for submission in subreddit.stream.submissions():
    title_blob = TextBlob(submission.title)
    row = {
      "subreddit": submission.subreddit.display_name, 
      "author": submission.author.name,
      "timestamp": datetime.utcfromtimestamp(submission.created_utc).replace(tzinfo=pytz.utc),
      "title": submission.title, 
      "url": submission.url,
      "polarity": title_blob.sentiment.polarity, 
      "subjectivity": title_blob.sentiment.subjectivity
    }
    await stream.write([row], immediate=True)
    log.info("write_post", subreddit=row["subreddit"], author=row["author"])

if __name__ == "__main__":
  loop = asyncio.new_event_loop()
  asyncio.set_event_loop(loop)
  loop.run_until_complete(main())
