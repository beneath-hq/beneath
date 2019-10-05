# README

### Secrets

This chart relies on the existance of the secret `control-secrets` in the `production` namespace. It's initialized like this:

    kubectl create secret generic control-secrets --namespace production --dry-run -o yaml \
      --from-literal pg-user=INSERT \
      --from-literal pg-password=INSERT \
      --from-literal session-secret=INSERT \
      --from-literal github-auth-id=INSERT \
      --from-literal github-auth-secret=INSERT \
      --from-literal google-auth-id=INSERT \
      --from-literal google-auth-secret=INSERT \
      | kubectl apply -f -

### Service account

This chart relies on the existance of a service account in `key.json` in the `control-sa-key` secret in the `production` namespace. It's created like this:

    gcloud beta iam service-accounts create control-service --display-name "Beneath Control Service Account"
    gcloud iam service-accounts list
    gcloud iam service-accounts keys create key.json --iam-account control-service@beneathcrypto.iam.gserviceaccount.com
    kubectl create secret generic control-sa-key --from-file key.json --namespace production
    rm key.json

    gcloud projects add-iam-policy-binding beneathcrypto --member serviceAccount:control-service@beneathcrypto.iam.gserviceaccount.com --role roles/pubsub.admin
    gcloud projects add-iam-policy-binding beneathcrypto --member serviceAccount:control-service@beneathcrypto.iam.gserviceaccount.com --role roles/bigtable.admin
    gcloud projects add-iam-policy-binding beneathcrypto --member serviceAccount:control-service@beneathcrypto.iam.gserviceaccount.com --role roles/redis.admin
    gcloud projects add-iam-policy-binding beneathcrypto --member serviceAccount:control-service@beneathcrypto.iam.gserviceaccount.com --role roles/bigquery.admin

### Postgres in production

You can access the production Postgres instance by running the Google Cloud SQL proxy on your local machine. 

- Follow the instructions [here](https://cloud.google.com/sql/docs/postgres/quickstart-proxy-test) to get the `cloud_sql_proxy` binary on your local machine
- Run the proxy with `./cloud_sql_proxy -instances=beneathcrypto:us-east1:beneath-network=tcp:5432`
- Connect to the database on `localhost` using a credentials for the production database (only for administrators)
