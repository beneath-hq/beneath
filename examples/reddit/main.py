import beneath

from config import SUBREDDIT
from generators import posts, comments

with open("schemas/post.graphql", "r") as file:
    POSTS_SCHEMA = file.read()

with open("schemas/comment.graphql", "r") as file:
    COMMENTS_SCHEMA = file.read()


def make_table_name(subreddit, kind):
    name = subreddit.replace("+", "-")
    return f"r-{name}-{kind}"


def make_subreddit_description(subreddit, kind):
    subs = [f"/r/{sub}" for sub in subreddit.split("+")]
    names = subs[0]
    if len(subs) > 1:
        names = ", ".join(subs[:-1]) + " and " + subs[-1]
    return (
        f"Reddit {kind} scraped in real-time from {names}. Some {kind} may be missing."
    )


if __name__ == "__main__":
    p = beneath.Pipeline(parse_args=True, disable_checkpoints=True)
    p.description = "Scrapes posts and comments from Reddit"

    posts = p.generate(posts.generate_posts)
    p.write_table(
        posts,
        make_table_name(SUBREDDIT, "posts"),
        schema=POSTS_SCHEMA,
        description=make_subreddit_description(SUBREDDIT, "posts"),
    )

    comments = p.generate(comments.generate_comments)
    p.write_table(
        comments,
        make_table_name(SUBREDDIT, "comments"),
        schema=COMMENTS_SCHEMA,
        description=make_subreddit_description(SUBREDDIT, "comments"),
    )

    p.main()
