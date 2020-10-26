# `deployments/gitlab/`

We use Gitlab to trigger builds and deployments. These docs detail how to set up Gitlab. Refer to `deployments/README.md` for an overview of the deployment lifecycle.

## Connect to Google Container Registry

In order to connect GitLab to the Google Container Registry, we need to create a GCP service account, grant it permissions, and generate a key to give to GitLab.

Create a service account

  gcloud iam service-accounts create gitlab-gcr-service --display-name "GitLab GCR Service Account"

Grant `storage.admin` permissions, which will allow read and write access. These permissions are for Google's Cloud Storage, which GCR uses.

  gcloud projects add-iam-policy-binding beneath --member "serviceAccount:gitlab-gcr-service@beneath.iam.gserviceaccount.com" --role "roles/storage.admin"

Generate key

  gcloud iam service-accounts keys create keyfile.json --iam-account gitlab-gcr-service@beneath.iam.gserviceaccount.com

Set the key as a build environment variable named `GCR_SERVICE_KEY` on the [CI/CD config page in Gitlab](https://gitlab.com/beneath-hq/beneath/-/settings/ci_cd) under "Variables".

Then delete the key from your local directory

  rm key.json

## Connect to the Kubernetes cluster

Connect to the Kubernetes cluster from the [Clusters page in the Gitlab project](https://gitlab.com/beneath-hq/beneath/-/clusters) by following the tutorial [here](https://gitlab.com/help/user/project/clusters/add_remove_clusters.md#existing-kubernetes-cluster). 

Doing so connects makes Gitlab automatically connect to the Kubernetes cluster from our CI/CD pipelines. 
