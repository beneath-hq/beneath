import asyncio
import beneath

from config import make_reddit
from datetime import datetime, timedelta

with open("schemas/post_score.graphql", "r") as file:
    POSTS_SCORES_SCHEMA = file.read()

with open("schemas/comment_score.graphql", "r") as file:
    COMMENTS_SCORES_SCHEMA = file.read()

PROJECT_PATH = "examples/reddit"
BATCH_SIZE = 100  # the maximum number of results reddit.info() returns
POLL_DELAY = timedelta(days=2)


def get_fullname_prefix(kind):
    if kind == "comments":
        return "t1_"
    if kind == "posts":
        return "t3_"


def get_schema(kind):
    if kind == "comments":
        return COMMENTS_SCORES_SCHEMA
    if kind == "posts":
        return POSTS_SCORES_SCHEMA


def make_record(item, kind):
    record = {
        "id": item.id,
        "created_on": datetime.utcfromtimestamp(item.created_utc),
        "score": item.score,
    }
    if kind == "posts":
        record["upvote_ratio"] = item.upvote_ratio
    return record


async def get_scores(kind):
    async with make_reddit() as reddit:
        client = beneath.Client()
        await client.start()

        # only used for first run:
        # table = await client.create_table(
        #     table_path=f"{PROJECT_PATH}/r-wallstreetbets-{kind}-scores",
        #     schema=get_schema(kind),
        #     description=f"Scores for {kind} on r/wallstreetbets. Fetched from Reddit {POLL_DELAY.days} days after posting.",
        #     update_if_exists=True,
        # )
        table = await client.find_table(
            f"{PROJECT_PATH}/r-wallstreetbets-{kind}-scores"
        )
        instance = table.primary_instance

        consumer = await client.consumer(
            f"{PROJECT_PATH}/r-wallstreetbets-{kind}",
            batch_size=BATCH_SIZE,
            subscription_path=f"{PROJECT_PATH}/r-wallstreetbets-scores-scraper",
        )

        async for batch in consumer.iterate(
            batches=True,
            stop_when_idle=True,
        ):
            # create the "fullnames" for the reddit request
            # for an explanation of fullnames, see the top of this page: https://www.reddit.com/dev/api/
            fullnames = []
            for item in batch:
                fullnames.append(get_fullname_prefix(kind) + item["id"])

            # ping reddit for scores
            resp = reddit.info(fullnames=fullnames)
            records = [make_record(item, kind) async for item in resp]

            # drop the batch if every item is not at least POLL_DELAY.days days old
            timestamp_of_last_item_in_batch = records[-1].get("created_on")
            if datetime.utcnow() - timestamp_of_last_item_in_batch < POLL_DELAY:
                break

            await instance.write(records)

        await client.stop()


async def main():
    await asyncio.gather(get_scores("comments"), get_scores("posts"))


asyncio.run(main())
