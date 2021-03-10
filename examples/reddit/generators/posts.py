from datetime import datetime

from config import reddit, SUBREDDIT, truncate_text


async def generate_posts(p):
    sub = await reddit.subreddit(SUBREDDIT)
    async for post in sub.stream.submissions(skip_existing=True):
        link = (
            post.url
            if post.url.replace("https://www.reddit.com", "") != post.permalink
            else None
        )
        yield {
            "created_on": datetime.utcfromtimestamp(post.created_utc),
            "id": post.id,
            "author": post.author.name if post.author else "[deleted]",
            "subreddit": post.subreddit.display_name,
            "title": post.title,
            "text": truncate_text(post.selftext),
            "link": link,
            "permalink": post.permalink,
            "flair": post.link_flair_text,
            "is_over_18": not not post.over_18,
            "is_original_content": not not post.is_original_content,
            "is_self_post": not not post.is_self,
            "is_distinguished": not not post.distinguished,
            "is_spoiler": not not post.spoiler,
            "is_stickied": not not post.stickied,
        }
