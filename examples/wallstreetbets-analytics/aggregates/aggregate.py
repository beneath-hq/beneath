import asyncio
import beneath
from datetime import datetime, timedelta


async def main():
    client = beneath.Client()
    table = await client.find_table(
        "examples/wallstreetbets-analytics/stock-mentions-rollup-daily"
    )
    yesterday = (datetime.now() - timedelta(days=1)).strftime("%Y-%m-%d")
    mentions = await beneath.query_warehouse(
        f"""
with
    vars as (
        select 
            timestamp("{yesterday}") as date,
            .05 as sentiment_cutoff,
    ),
    stock_mentions_posts_calc_sentiment as (
        select 
            symbol, 
            timestamp, 
            (title_polarity*num_mentions_title+text_polarity*num_mentions_body)/(num_mentions_title+num_mentions_body) as polarity,
            (title_subjectivity*num_mentions_title+text_subjectivity*num_mentions_body)/(num_mentions_title+num_mentions_body) as subjectivity,
        from `examples/wallstreetbets-analytics/r-wallstreetbets-posts-stock-mentions` m, vars
        join `examples/wallstreetbets-analytics/r-wallstreetbets-posts-sentiment` s on m.post_id=s.post_id
        where timestamp_trunc(timestamp, day) = vars.date
    ),
    stock_mentions_posts as (
        select
            symbol,
            timestamp_trunc(timestamp, day) as day,
            count(*) as num_mentions,
            countif(polarity >= vars.SENTIMENT_CUTOFF) as num_positive,
            countif(polarity < vars.SENTIMENT_CUTOFF and polarity > -vars.SENTIMENT_CUTOFF) as num_neutral,
            countif(polarity <= -vars.SENTIMENT_CUTOFF) as num_negative,
            avg(polarity) as avg_polarity,
            avg(subjectivity) as avg_subjectivity
        from stock_mentions_posts_calc_sentiment, vars
        group by symbol, timestamp_trunc(timestamp, day)
    ),
    stock_mentions_comments as (
        select 
            symbol, 
            timestamp_trunc(timestamp, day) as day, 
            count(*) as num_mentions,
            countif(polarity >= vars.SENTIMENT_CUTOFF) as num_positive,
            countif(polarity < vars.SENTIMENT_CUTOFF and polarity > -vars.SENTIMENT_CUTOFF) as num_neutral,
            countif(polarity <= -vars.SENTIMENT_CUTOFF) as num_negative,
            avg(polarity) as avg_polarity,
            avg(subjectivity) as avg_subjectivity
        from `examples/wallstreetbets-analytics/r-wallstreetbets-comments-stock-mentions` m, vars
        join `examples/wallstreetbets-analytics/r-wallstreetbets-comments-sentiment` s on m.comment_id=s.comment_id
        where timestamp_trunc(timestamp, day) = vars.date
        group by symbol, timestamp_trunc(timestamp, day)
    )
select 
    coalesce(p.symbol, c.symbol) as symbol,
    coalesce(p.day, c.day) as day,
    ifnull(p.num_mentions, 0) + ifnull(c.num_mentions,0) as num_mentions,
    ifnull(p.num_positive, 0) + ifnull(c.num_positive,0) as num_positive,
    ifnull(p.num_neutral, 0) + ifnull(c.num_neutral,0) as num_neutral,
    ifnull(p.num_negative, 0) + ifnull(c.num_negative,0) as num_negative,
    (ifnull(p.num_mentions*p.avg_polarity, 0) + ifnull(c.num_mentions*c.avg_polarity, 0))/(ifnull(p.num_mentions, 0) + ifnull(c.num_mentions,0)) as avg_polarity,
    (ifnull(p.num_mentions*p.avg_subjectivity, 0) + ifnull(c.num_mentions*c.avg_subjectivity, 0))/(ifnull(p.num_mentions, 0) + ifnull(c.num_mentions,0)) as avg_subjectivity
from stock_mentions_posts p
full join stock_mentions_comments c on p.symbol = c.symbol and p.day = c.day
order by symbol, day
"""
    )

    await client.start()
    await table.write(mentions.to_dict("records"))
    await client.stop()


asyncio.run(main())
