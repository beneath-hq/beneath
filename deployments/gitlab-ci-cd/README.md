# GitLab CI/CD

In order to connect GitLab to the Google Container Registry, we need to create a GCP service account, grant it permissions, and generate a key to give to GitLab.

Create a service account
```
gcloud iam service-accounts create gitlab-gcr-service --display-name "GitLab GCR Service Account"
```

Grant "storage.admin" permissions, which will allow read and write access. These permissions are for Google's Cloud Storage, which GCR uses.
```
gcloud projects add-iam-policy-binding beneath --member "serviceAccount:gitlab-gcr-service@beneath.iam.gserviceaccount.com" --role "roles/storage.admin"
```

Generate key
```
gcloud iam service-accounts keys create keyfile.json --iam-account gitlab-gcr-service@beneath.iam.gserviceaccount.com
```

Set the key in Gitlab in https://gitlab.com/beneath-hq/beneath/-/settings/ci_cd under “Variables” ```GCR_SERVICE_KEY```


Then delete the key from your local directory
```
rm key.json
```
