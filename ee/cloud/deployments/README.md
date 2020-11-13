# `ee/cloud/deployments/`

This directory contains Kubernetes and Helm manifests for deploying beneath.dev to production.

The beneath.dev deployment of Beneath centers around:

- Docker for specifying builds (see `ee/build/`)
- Kubernetes for running code (see `ee/cloud/deployments/kube/`)
- Helm for specifying deployments (see `ee/cloud/deployments/helm/`)
- Gitlab for triggering builds and deployments (see `.gitlab-ci.yml` in root, `ee/`, and `ee/cloud/`)

The sections in this document explains how to set up everything from scratch.

## Step 0: Prepare infrastructure in the cloud console

- Create Postgres and Redis systems (note down DNS names and access credentials)
- Create relevant engine systems (e.g. Bigtable)
- Create a Kubernetes cluster and set up `kubectl` on your local machine

## Step 1: Configure the Kubernetes cluster

Follow the steps given in `ee/cloud/deployments/kube/`.

## Step 2: Configure Gitlab

Follow the steps given in `ee/cloud/deployments/gitlab/`.

## Step 3: Run builds and deploys

First, go through the `ee/cloud/deployments/helm/*_values.yaml` files, and make sure it's configured correctly.

Then go to the most recent pipeline for the `stable` branch in Gitlab (see [https://gitlab.com/beneath-hq/beneath/pipelines](https://gitlab.com/beneath-hq/beneath/pipelines)) and trigger builds and deploys.
