# Lending Club historical loan analysis and realtime assessment

The code in these folders produces the three data streams in the [Beneath Lending Club project](https://beneath.dev/epg/lending-club):
- [historical loans with performance data](https://beneath.dev/epg/lending-club/loans-history)
- [real-time loans listed on the Lending Club website 4x per day](https://beneath.dev/epg/lending-club/loans)
- [real-time loans enriched with performance predictions](https://beneath.dev/epg/lending-club/loans-enriched)

### Developing the streams

- The historical loan data is uploaded to Beneath via the Jupyter notebook `loans-history/load_historical_loans.ipynb`.
- The real-time loan data is captured with the `loans/fetch_new_loans.py` script.
- The real-time enriched loan data is created with the `loans-enriched/enrich_loans.py` script.
- To make performance predictions, the enriched data stream utilizes the machine learning model that is trained with the `loans-enriched/train_model.ipynb` notebook.

To set yourself up for development:

    python3 -m venv .venv
    source .venv/bin/activate
    pip install -r requirements.txt

To create the Beneath project where the stream is stored (epg is my username - you'll have to use your own):

    beneath project create epg/lending-club

To stage the Beneath services:

    python ./loans/fetch-new-loans.py stage epg/lending-club/fetch-new-loans --read-quota-mb 10000 --write-quota-mb 10000
    python ./loans-enriched/enrich-loans.py stage epg/lending-club/enrich-loans --read-quota-mb 10000 --write-quota-mb 10000
    
The secrets used to connect to Beneath were issued with:

    beneath service issue-secret epg/lending-club/fetch-new-loans --description kubernetes
    beneath service issue-secret epg/lending-club/enrich-loans --description kubernetes

Then apply the service secrets to Kubernetes (we're using a namespace called `models`):

    kubectl create secret generic lending-club-loans-service-secret -n models --from-literal secret=SECRET
    kubectl create secret generic lending-club-loans-enriched-service-secret -n models --from-literal secret=SECRET

To connect to Lending Club's API for newly-listed loans:
- Create a Lending Club account, register as an Investor, and request API access. Save the API key.

Then apply the API key to Kubernetes:

    kubectl create secret generic lending-club-api-key -n models --from-literal secret=LENDING_CLUB_API_KEY

To rebuild the Docker images:

    docker build -t gcr.io/beneath/lending-club-loans:latest .
    docker push gcr.io/beneath/lending-club-loans:latest

    docker build -t gcr.io/beneath/lending-club-loans-enriched:latest .
    docker push gcr.io/beneath/lending-club-loans-enriched:latest
   
To deploy to Kubernetes:

    kubectl apply -f loans/kube.yaml -n models
    kubectl apply -f loans-enriched/kube.yaml -n models

Important: There must ever only be one replica of these scripts running at a time. (Running multiple wouldn't hurt consistency, but would cause duplicates to appear in the streaming view and in the data warehouses.)
