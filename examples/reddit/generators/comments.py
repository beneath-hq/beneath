from datetime import datetime

from config import reddit, SUBREDDIT


async def generate_comments(p):
    sub = await reddit.subreddit(SUBREDDIT)
    async for comment in sub.stream.comments():
        await comment.submission.load()
        yield {
            "created_on": datetime.utcfromtimestamp(comment.created_utc),
            "id": comment.id,
            "author": comment.author.name,
            "subreddit": comment.subreddit.display_name,
            "post": comment.submission.title,
            "parent_id": comment.parent_id,
            "body": comment.body,
            "permalink": comment.permalink,
            "is_submitter": comment.is_submitter,
            "is_distinguished": not not comment.distinguished,
            "is_stickied": not not comment.stickied,
        }
