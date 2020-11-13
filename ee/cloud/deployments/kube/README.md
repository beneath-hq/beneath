# `deployments/kube/`
​
This file should keep a complete log of configuration of the Kubernetes cluster. Our ambition is that after configuring the cluster, we should not have to directly modify Kubernetes from the command-line – everything should happen through helm deployments triggered from Gitlab.
​​
### ingress
​
Nginx ingress handles load balancing and routing.
​
Create an external IP in Google Cloud:
​
    gcloud compute addresses create beneath-ip-1 --project=beneath --region=us-east1

[Install Helm 3](https://helm.sh/docs/intro/install/) if you don't already have it installed. Make sure you are not using Helm 2!
​
Add ingress-nginx charts to Helm:
​
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
    helm repo update
​
Create namespace for ingress:
​
    kubectl create namespace nginx

Create nginx ingress in cluster. (Change `loadBalancerIP` to the external IP you just provisioned.)
​
    helm upgrade --wait --install --namespace nginx ingress-nginx ingress-nginx/ingress-nginx --set controller.service.loadBalancerIP="35.231.110.138" --set controller.service.externalTrafficPolicy=Local

NOTE: There are two different NGINX ingress implementations, [ingress-nginx](https://github.com/kubernetes/ingress-nginx) and [nginx-ingress](https://github.com/nginxinc/kubernetes-ingress). Some men just want to watch the world burn. We use the *former*, make sure always to refer to the [correct documentation](https://kubernetes.github.io/ingress-nginx/)!
​
### cert-manager
​
Cert-manager automatically handles TLS (HTTPS) for all public-facing ingresses. We run it in the `cert-manager` namespace, but it can issue certificates for ingresses in all namespaces.

Refer to https://cert-manager.io/docs/installation/kubernetes/ for up-to-date instructions.
​
Add helm repo:

    helm repo add jetstack https://charts.jetstack.io
    helm repo update
​
Create namespace:

    kubectl create namespace cert-manager
    
Deploy cert-manager:

    kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.15.0/cert-manager.crds.yaml
    helm install cert-manager jetstack/cert-manager --namespace cert-manager --version v0.15.0

Create cluster-wide issuer:

    kubectl apply -n cert-manager -f cert-manager/cluster-issuer-prod.yaml 

### Optional: test progress

You can apply the `test/test-resources.yaml` manifest to deploy a hello-world app to the cluster with HTTPS enabled.

    kubectl apply -f test/test-resources.yaml

Remember to delete it:

    kubectl delete -f test/test-resources.yaml

### Create production namespace
​
Create the production namespace like this:
​
    kubectl create namespace production

### Backend config file

The `backend` Helm deployment relies on the existance of a `backend-config` secret in the `production` namespace. It's initialized like this (assumes a config file `production.yaml` is available):

```
kubectl create secret generic backend-config --from-file production.yaml --namespace production
```

Below is an approximate list of places from which the secrets used in `production.yaml` were sourced:

- Postgres credentials: Cloud SQL admin console
- Session secret: randomly generated
- Stripe credentials: Stripe dashboard
- Github credentials: The [OAuth apps section](https://github.com/organizations/beneath-hq/settings/applications) of the [Github project settings]
- Google credentials: The [Google developer console](https://console.developers.google.com/apis/credentials?authuser=1&project=beneath&supportedpurview=project) (note that it's different from the Cloud console!)

### Google Cloud service account for pods

The Helm charts relies on the existance of a service account in `key.json` in the `beneath-sa-key` secret in the `production` namespace. It's created like this:

    gcloud iam service-accounts create beneath-service --display-name "Beneath Service Account (used by services in Kubernetes)"
    gcloud iam service-accounts list
    gcloud iam service-accounts keys create key.json --iam-account beneath-service@beneath.iam.gserviceaccount.com
    kubectl create secret generic beneath-sa-key --from-file key.json --namespace production
    rm key.json

    gcloud projects add-iam-policy-binding beneath --member serviceAccount:beneath-service@beneath.iam.gserviceaccount.com --role roles/pubsub.admin
    gcloud projects add-iam-policy-binding beneath --member serviceAccount:beneath-service@beneath.iam.gserviceaccount.com --role roles/bigtable.admin
    gcloud projects add-iam-policy-binding beneath --member serviceAccount:beneath-service@beneath.iam.gserviceaccount.com --role roles/redis.admin
    gcloud projects add-iam-policy-binding beneath --member serviceAccount:beneath-service@beneath.iam.gserviceaccount.com --role roles/bigquery.admin

### Postgres

For admin purposes, we sometimes need to connect to the postgres database. We do so as the `postgres` user with:

  gcloud sql connect beneath-postgres-1 -u postgres

All resources (tables, views, sequences, etc.) should be created and owned by the `control` user. For example, running `\dt TABLE_NAME` must always reveal `control` as the `owner`.

The `postgres` user has all the same privileges as `control` (initially, we ran `GRANT control TO postgres;` from the `control` user).

NOTE: Be very careful when administering the database! To minimize the chance of mistakes: Connect only with `psql` (not a GUI), write your queries in a text editor before running them, and don't keep the connection open for longer than necessary.

### Redis

To connect to the Redis database (Cloud Memorystore) in productin, run:

    kubectl run -i --tty redisbox --image=gcr.io/google_containers/redis:v1 -- sh

Then within the container, run:

    redis-cli -h 10.255.145.163
