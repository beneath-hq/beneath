import asyncio
import beneath


async def main():
    client = beneath.Client()
    await client.create_table(
        "examples/wallstreetbets-analytics/stock-mentions-rollup-daily",
        schema="""
" A rollup table of stock mentions "
type MentionDay @schema {
    symbol: String! @key
    day: Timestamp! @key
    num_mentions: Int32!
    num_positive: Int32!
    num_neutral: Int32!
    num_negative: Int32!
    avg_polarity: Float32!
    avg_subjectivity: Float32!
}
        """,
        update_if_exists=True,
    )


asyncio.run(main)