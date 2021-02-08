from datetime import datetime

from config import reddit, SUBREDDIT, MAX_CHARACTERS


async def generate_posts(p):
    sub = await reddit.subreddit(SUBREDDIT)
    async for post in sub.stream.submissions():
        yield {
            "created_on": datetime.utcfromtimestamp(post.created_utc),
            "id": post.id,
            "author": post.author.name,
            "subreddit": post.subreddit.display_name,
            "title": post.title,
            "text": post.selftext[:MAX_CHARACTERS] + " [CONTENT TRUNCATED]"
            if len(post.selftext) > MAX_CHARACTERS
            else post.selftext,
            "link": post.url
            if post.url.replace("https://www.reddit.com", "") != post.permalink
            else None,
            "permalink": post.permalink,
            "flair": post.link_flair_text,
            "is_over_18": not not post.over_18,
            "is_original_content": not not post.is_original_content,
            "is_self_post": not not post.is_self,
            "is_distinguished": not not post.distinguished,
            "is_spoiler": not not post.spoiler,
            "is_stickied": not not post.stickied,
        }
