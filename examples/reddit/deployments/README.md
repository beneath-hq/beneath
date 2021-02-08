To set all the environment variables as one Kubernetes secret:

```bash
kubectl create secret generic wallstreetbets-scraper -n models --from-literal=beneath-secret=BENEATH_SECRET \
    --from-literal=reddit-user-agent=REDDIT_USER_AGENT \
    --from-literal=reddit-client-id=REDDIT_CLIENT_ID \
    --from-literal=reddit-client-secret=REDDIT_CLIENT_SECRET \
    --from-literal=reddit-username=REDDIT_USERNAME \
    --from-literal=reddit-password=REDDIT_PASSWORD \
    --from-literal=reddit-subreddit=REDDIT_SUBREDDIT
```