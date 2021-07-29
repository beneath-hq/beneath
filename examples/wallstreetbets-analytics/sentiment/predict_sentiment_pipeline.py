"""
This is a first attempt at deriving sentiment from posts and comments in r/wallstreetbets. It uses an 
out-of-the-box text analyzer called "textblob" (https://textblob.readthedocs.io/en/dev/). It's ok, not great. 
The r/wallstreetbets vernacular is unique and has a bunch of obvious indicators of positivity that textblob 
doesn't pick up on, e.g. rocket emojis and chants of 'to the moon!' A better approach would be to annotate 
some r/wallstreetbets submissions and train a sentiment model specific to WSB.
"""

import beneath
from textblob import TextBlob

POSTS_SENTIMENT_SCHEMA = """
    type Sentiment @schema {
        post_id: String! @key
        " Title polarity on a scale from -1 (negative) to 0 (neutral) to 1 (positive). Computed with the textblob python package. "
        title_polarity: Float!
        " Title subjectivity on a scale from 0 (objective) to 1 (subjective). Computed with the textblob python package. "
        title_subjectivity: Float!
        " Body polarity on a scale from -1 (negative) to 0 (neutral) to 1 (positive). Computed with the textblob python package. "
        text_polarity: Float!
        " Body subjectivity on a scale from 0 (objective) to 1 (subjective). Computed with the textblob python package. "
        text_subjectivity: Float!
    }
"""

COMMENTS_SENTIMENT_SCHEMA = """
    type Sentiment @schema {
        comment_id: String! @key
        " Polarity on a scale from -1 (negative) to 0 (neutral) to 1 (positive). Computed with the textblob python package. "
        polarity: Float!
        " Subjectivity on a scale from 0 (objective) to 1 (subjective). Computed with the textblob python package. "
        subjectivity: Float!
    }
"""


async def predict_post_sentiment(record):
    title_blob = TextBlob(record["title"])
    text_blob = TextBlob(record["text"])
    yield {
        "post_id": record["id"],
        "title_polarity": title_blob.sentiment.polarity,
        "title_subjectivity": title_blob.sentiment.subjectivity,
        "text_polarity": text_blob.sentiment.polarity,
        "text_subjectivity": text_blob.sentiment.subjectivity,
    }


async def predict_comment_sentiment(record):
    text_blob = TextBlob(record["text"])
    yield {
        "comment_id": record["id"],
        "polarity": text_blob.sentiment.polarity,
        "subjectivity": text_blob.sentiment.subjectivity,
    }


if __name__ == "__main__":
    p = beneath.Pipeline(parse_args=True)
    p.description = "Predicts the sentiment for r/wallstreetbets posts and comments"

    posts = p.read_table("examples/reddit/r-wallstreetbets-posts")
    posts_sentiment = p.apply(posts, predict_post_sentiment)
    p.write_table(
        posts_sentiment,
        "r-wallstreetbets-posts-sentiment",
        schema=POSTS_SENTIMENT_SCHEMA,
        description="Sentiment of r/wallstreetbets posts",
    )

    comments = p.read_table("examples/reddit/r-wallstreetbets-comments")
    comments_sentiment = p.apply(comments, predict_comment_sentiment)
    p.write_table(
        comments_sentiment,
        "r-wallstreetbets-comments-sentiment",
        schema=COMMENTS_SENTIMENT_SCHEMA,
        description="Sentiment of r/wallstreetbets comments",
    )

    p.main()
