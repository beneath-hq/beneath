import beneath
import pandas as pd
import re

MENTIONS_POSTS_SCHEMA = """
    type Mention @schema {
        symbol: String! @key
        timestamp: Timestamp! @key
        post_id: String!
        author: String!
        num_mentions_title: Int!
        num_mentions_body: Int!
        num_unique_symbols_title: Int!
        num_unique_symbols_body: Int!
        length_title: Int!
        length_body: Int!
    }
"""

MENTIONS_COMMENTS_SCHEMA = """
    type Mention @schema {
        symbol: String! @key
        timestamp: Timestamp! @key
        comment_id: String!
        parent_id: String!
        post_id: String!
        author: String!
        num_mentions: Int!
        num_unique_symbols: Int!
        length_text: Int!
    }
"""

BLACKLIST = {
    "FOMO",
    "DD",
    "EOD",
    "TA",
    "PT",
    "RSI",
    "HUGE",
    "ATH",
    "USA",
    "AI",
    "IMO",
    "AM",
    "UK",
    "BIG",
    "SO",
    "OR",
    "FOR",
    "ALL",
    "IT",
    "BE",
    "ARE",
    "NOW",
    "ON",
    "ME",
    "CAN",
    "VERY",
    "SI",
    "TV",
    "BY",
    "NEW",
    "OUT",
    "LOVE",
    "GO",
    "PM",
    "NEXT",
    "ANY",
    "ET",
    "HAS",
    "ONE",
    "PLAY",
    "LOW",
    "III",
    "CASH",
    "RNG",
    "GOOD",
    "REAL",
    "SEE",
    "RE",
}

# rudimentary cache for the stock symbols
_symbols = None

# load a list of stock symbols from Beneath
async def get_symbols():
    global _symbols
    if _symbols is None:
        df = await beneath.load_full("examples/financial-reference-data/stock-symbols")
        _symbols = set(df["symbol"]) - BLACKLIST
    return _symbols


def extract_symbol_candidates(text):
    candidates = re.findall(r"((\$[a-zA-Z]{1,5}\b)|(\b[A-Z]{2,5}\b))", text)
    candidates_clean = []
    for x in candidates:
        candidates_clean.append(x[0].lstrip("$").upper())
    return candidates_clean


async def filter_for_real_symbols(df):
    SYMBOLS = await get_symbols()
    return df.loc[[x in SYMBOLS for x in df.index]]


async def find_mentions_posts(record):
    # create df of mentions
    title_mentions = (
        pd.Series(extract_symbol_candidates(record["title"]), dtype="object")
        .value_counts()
        .rename("mentions_in_title")
    )
    body_mentions = (
        pd.Series(extract_symbol_candidates(record["text"]), dtype="object")
        .value_counts()
        .rename("mentions_in_body")
    )
    df = pd.concat([title_mentions, body_mentions], axis=1).fillna(0)
    df = await filter_for_real_symbols(df)

    # compute aggregate statistics
    num_unique_symbols_title = sum(df["mentions_in_title"] > 0)
    num_unique_symbols_body = sum(df["mentions_in_body"] > 0)

    for index, row in df.iterrows():
        yield {
            "symbol": index,
            "timestamp": record["created_on"],
            "post_id": record["id"],
            "author": record["author"],
            "num_mentions_title": row["mentions_in_title"],
            "num_mentions_body": row["mentions_in_body"],
            "num_unique_symbols_title": num_unique_symbols_title,
            "num_unique_symbols_body": num_unique_symbols_body,
            "length_title": len(record["title"]),
            "length_body": len(record["text"]),
        }


async def find_mentions_comments(record):
    # create df of mentions
    df = (
        pd.Series(extract_symbol_candidates(record["text"]), dtype="object")
        .value_counts()
        .rename("mentions")
        .to_frame()
    )
    df = await filter_for_real_symbols(df)

    # compute aggregate statistics
    num_unique_symbols = sum(df["mentions"] > 0)

    for index, row in df.iterrows():
        yield {
            "symbol": index,
            "timestamp": record["created_on"],
            "comment_id": record["id"],
            "parent_id": record["parent_id"],
            "post_id": record["post_id"],
            "author": record["author"],
            "num_mentions": row["mentions"],
            "num_unique_symbols": num_unique_symbols,
            "length_text": len(record["text"]),
        }


if __name__ == "__main__":
    p = beneath.Pipeline(parse_args=True)
    p.description = (
        "Finds stock symbols mentioned in r/wallstreetbets posts and comments"
    )

    posts = p.read_table("examples/reddit/r-wallstreetbets-posts")
    mentions = p.apply(posts, find_mentions_posts)
    p.write_table(
        mentions,
        "r-wallstreetbets-posts-stock-mentions",
        schema=MENTIONS_POSTS_SCHEMA,
        description="Stock mentions in r/wallstreetbets posts",
    )

    comments = p.read_table("examples/reddit/r-wallstreetbets-comments")
    mentions = p.apply(comments, find_mentions_comments)
    p.write_table(
        mentions,
        "r-wallstreetbets-comments-stock-mentions",
        schema=MENTIONS_COMMENTS_SCHEMA,
        description="Stock mentions in r/wallstreetbets comments",
    )

    p.main()
