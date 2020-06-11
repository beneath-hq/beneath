# Lending Club historical loan analysis and realtime assessment

- The code in these folders produces two data streams: one is the raw loans from the Lending Club website, and another is an enriched stream that includes predictions for whether or not the loan borrower will default
- The streams are deployed in the project at [beneath.dev/epg/lending-club](https://beneath.dev/epg/lending-club)

### Developing the streams

- The historical loan data is uploaded to Beneath via the Jupyter notebook `loans/load_historical_loans.ipynb`.
- The new loan data is captured with the `loans/fetch_new_loans.py` script.
- The (quick-and-dirty!) machine learning model is trained with the `loans-enriched/train_model.ipynb` notebook.
- The enriched data stream is created with the `loans-enriched/enrich_loans.py` script.

The `loans/fetch_new_loans.py` and `loans-enriched/enrich_loans.py` scripts are run in Kubernetes in a namespace called `models`.

To set yourself up for development:

    python3 -m venv .venv
    source .venv/bin/activate
    pip install -r requirements.txt

To create the Beneath project where the stream is stored:

    beneath project create epg/lending-club

To stage the Beneath streams:

    beneath stream stage epg/lending-club/loans
    beneath stream stage epg/lending-club/loans-enriched

To connect to Lending Club's API for newly-listed loans:
- Create a Lending Club account, register as an Investor, and request API access. Save the API key.

Then apply the API key to Kubernetes:

    kubectl create secret generic lending-club-api-key -n models --from-literal secret=LENDING_CLUB_API_KEY

To connect to Beneath, create a service and issue a service secret:

    beneath service create epg/lending-club-service --read-quota-mb 100 --write-quota-mb 2000
    beneath service update-permissions epg/lending-club-service epg/lending-club/loans --read --write 
    beneath service update-permissions epg/lending-club-service epg/lending-club/loans-enriched --read --write 
    beneath service issue-secret epg/lending-club-service --description kubernetes

Then apply the service secret to Kubernetes:

    kubectl create secret generic lending-club-service-secret -n models --from-literal secret=SECRET

To rebuild the Docker images:

    docker build -t gcr.io/beneath/lending-club-loans:latest .
    docker push gcr.io/beneath/lending-club-loans:latest

    docker build -t gcr.io/beneath/lending-club-loans-enriched:latest .
    docker push gcr.io/beneath/lending-club-loans-enriched:latest
   
To deploy to Kubernetes:

    kubectl apply -f loans/kube.yaml -n models
    kubectl apply -f loans-enriched/kube.yaml -n models

Important: There must ever only be one replica of the script running at a time. (Running multiple wouldn't hurt consistency, but would cause duplicates to appear in the streaming view and in the data warehouses.)
