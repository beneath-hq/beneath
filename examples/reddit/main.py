import beneath
from datetime import datetime

from config import SUBREDDIT
from generators import posts

with open("schemas/post.graphql", "r") as file:
    POSTS_SCHEMA = file.read()

def make_stream_name(subreddit, kind):
    name = subreddit.replace("+", "-")
    return f"{name}-{kind}"

def make_subreddit_description(subreddit, kind):
    subs = [f"/r/{sub}" for sub in subreddit.split("+")]
    if len(subs) == 1:
        return subs[0]
    names = ", ".join(subs[:-1]) + " and " + subs[-1]
    return f"Reddit {kind} scraped in real-time from {names}. Some posts may be missing."

if __name__ == "__main__":
    p = beneath.Pipeline(parse_args=True)
    posts = p.generate(posts.generate_posts)
    p.write_stream(
        posts,
        make_stream_name(SUBREDDIT, "posts"),
        schema=POSTS_SCHEMA,
        description=make_subreddit_description(SUBREDDIT, "posts"),
    )
    p.main()

