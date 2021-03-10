from datetime import datetime

from config import reddit, SUBREDDIT, truncate_text


async def generate_comments(p):
    sub = await reddit.subreddit(SUBREDDIT)
    async for comment in sub.stream.comments(skip_existing=True):
        yield {
            "created_on": datetime.utcfromtimestamp(comment.created_utc),
            "id": comment.id,
            "author": comment.author.name if comment.author else "[deleted]",
            "subreddit": comment.subreddit.display_name,
            "post_id": comment.submission.id,
            "parent_id": comment.parent_id,
            "text": truncate_text(comment.body),
            "permalink": comment.permalink,
            "is_submitter": not not comment.is_submitter,
            "is_distinguished": not not comment.distinguished,
            "is_stickied": not not comment.stickied,
        }
