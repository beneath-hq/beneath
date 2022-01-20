import asyncio
import beneath


async def main():
    client = beneath.Client()

    await client.create_table(
        "examples/wallstreetbets-analytics/stock-metrics-24h-by-symbol",
        schema="""
" Stocks metrics computed over the last 24 hours "
type StockMetrics24h @schema {
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

    await client.create_table(
        "examples/wallstreetbets-analytics/stock-metrics-24h-by-timestamp",
        schema="""
" Top stocks of each day "
type StockMetrics24h @schema {
    day: Timestamp! @key
    symbol: String! @key
    num_mentions: Int32!
    num_positive: Int32!
    num_neutral: Int32!
    num_negative: Int32!
    num_mentions_rank: Int32!
}
        """,
        update_if_exists=True,
    )


asyncio.run(main())