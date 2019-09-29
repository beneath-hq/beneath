# README

Kubernetes is tricky. For everything not deployed completely through Gitlab, this file should contain a complete log of commands executed against the cluster.

### Namespaces

Create the production namespace like this:

    kubectl create namespace production

### tiller

To use Helm to manage packages, we must run tiller. It runs in the kube-system namespace.

    kubectl apply -f kube/tiller-service-account.yaml
    helm init --upgrade --service-account tiller

### ingress

Nginx ingress handles load balancing and routing. It runs in the default namespace. We installed it using Helm:

    helm install stable/nginx-ingress --name nginx-ingress --set rbac.create=true

### cert-manager

Cert-manager automatically handles TLS (HTTPS) for all public-facing ingresses. It runs in the cert-manager namespace, but can issue certificates for ingresses in all namespaces.

    kubectl apply -f https://raw.githubusercontent.com/jetstack/cert-manager/release-0.10/deploy/manifests/00-crds.yaml
    kubectl create namespace cert-manager
    kubectl label namespace cert-manager certmanager.k8s.io/disable-validation=true
    helm repo add jetstack https://charts.jetstack.io
    helm repo update
    helm install --name cert-manager --namespace cert-manager --version v0.10.1 jetstack/cert-manager
    kubectl apply -f kube/cert-manager-cluster-issuer-prod.yaml --namespace cert-manager
