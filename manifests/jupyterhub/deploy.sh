# add helm repo
helm repo add jupyterhub https://jupyterhub.github.io/helm-chart/
helm repo update

# helm names
RELEASE=jupyterhub
NAMESPACE=jupyterhub

# run upgrade
helm upgrade --install $RELEASE jupyterhub/jupyterhub \
  --namespace $NAMESPACE  \
  --version 0.7.0 \
  --values helm/config.yaml
