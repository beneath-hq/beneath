# `deployments/`

This directory contains Kubernetes and Helm manifests for deploying to production.

The beneath.dev deployment of Beneath centers around:

- Docker for specifying builds (see `build/`)
- Kubernetes for running code (see `deployments/kube/`)
- Helm for specifying deployments (see `deployments/helm/`)
- Gitlab for triggering builds and deployments (see `.gitlab-ci.yml`)

The sections in this document explains how to set up everything from scratch.

## Step 0: Prepare infrastructure in the cloud console

- Create Postgres and Redis systems (note down DNS names and access credentials)
- Create relevant engine systems (e.g. Bigtable)
- Create a Kubernetes cluster and set up `kubectl` on your local machine

## Step 1: Configure the Kubernetes cluster

Follow the steps given in `deployments/kube/`.

## Step 2: Configure Gitlab

Follow the steps given in `deployments/gitlab/`.

## Step 3: Run builds and deploys

First, go through the `deployments/helm/NAME/values.yaml` files, where many parameters are still hard-coded, and ensure they are up-to-date and match your deployment.

Then go to the most recent pipeline for the `stable` branch in Gitlab (see [https://gitlab.com/beneath-hq/beneath/pipelines](https://gitlab.com/beneath-hq/beneath/pipelines)) and trigger builds and deploys.
