# add helm repo
helm repo add jupyterhub https://jupyterhub.github.io/helm-chart/
helm repo update

# run upgrade
helm upgrade --install jupyterhub jupyterhub/jupyterhub \
  --namespace jupyterhub  \
  --version 0.7.0 \
  --values helm/config.yaml

# 0.9-174bbd5

# !pip install psycopg2-binary
# !pip install google-cloud-bigquery
# !pip install squarify
