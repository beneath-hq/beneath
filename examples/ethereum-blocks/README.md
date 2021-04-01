To test locally:

Set up poetry virtual environment

```bash
poetry install
poetry shell
```

Set a Web3 provider and run a simulation:

```bash
WEB3_PROVIDER_URL=https://cloudflare-eth.com python main.py test
```

For reference, here's the complete list of build and deploy commands for the Kubernetes deployment:

```bash
python main.py stage examples/ethereum/blocks-scraper --read-quota-mb 1000 --write-quota-mb 20000
beneath service issue-secret examples/ethereum/blocks-scraper
kubectl create secret generic ethereum-blocks -n models --from-literal=beneath-secret=XXX
docker build -t gcr.io/beneath/examples-ethereum-blocks:latest .
docker push gcr.io/beneath/examples-ethereum-blocks:latest
kubectl apply -f kube.yaml -n models
```
