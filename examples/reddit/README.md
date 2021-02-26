To run locally:

Set up poetry virtual environment

```bash
poetry install
poetry shell
```

Stage the streams and a service for the pipeline, then run it

```bash
REDDIT_CLIENT_ID="" REDDIT_CLIENT_SECRET="" REDDIT_USER_AGENT="" REDDIT_USERNAME="" REDDIT_PASSWORD="" REDDIT_SUBREDDIT="" python main.py stage USERNAME/PROJECT/SUBREDDIT-scraper --write-quota-mb 10000
REDDIT_CLIENT_ID="" REDDIT_CLIENT_SECRET="" REDDIT_USER_AGENT="" REDDIT_USERNAME="" REDDIT_PASSWORD="" REDDIT_SUBREDDIT="" python main.py run USERNAME/PROJECT/SUBREDDIT-scraper
```