---
title: Lending Club loan analysis
description: Use Beneath, Kubernetes, and Metabase to build an end-to-end application with real-time stream processing and business intelligence
menu:
  docs:
    parent: example-applications
    weight: 100
weight: 100
---

IN PROGRESS

In this walk-through, we'll create an end-to-end analytics application with Beneath, Kubernetes, and Metabase. We have an interesting data source in the [Lending Club](https://www.lendingclub.com/) loan data, and we want to do some classic things: 
- Get all the data (both historical and real-time) and store it in our own system
- Connect a BI tool so that we can easily explore and visualize the data
- Develop a model to predict which data points are good and bad
- Apply the model to new data and generate real-time predictions
- Expose an API so we can act on our predictions

Together, these activities encompass a modern analytics application. Using Beneath as your core storage and processing system, you'll see that these things are made simple and fall into place.

## What we'll build

- First, we'll create this [stream of historical and real-time loan data](https://beneath.dev/epg/lending-club/loans).
- Then we'll create this [Metabase dashboard](https://metabase.demo.beneath.dev/dashboard/1).
- Finally, we'll create this [stream of loan data, enriched with performance predictions](https://beneath.dev/epg/lending-club/loans-enriched).

For reference, here's all [the code](https://gitlab.com/beneath-hq/beneath/-/tree/master/clients/python/examples/lending-club).

## Prerequisites

In the interest of getting-to-the-point and focusing on Beneath, we'll assume you can set yourself up with the following:
- A [Kubernetes](https://kubernetes.io/docs/home/) cluster. There are [many ways to set up Kubernetes](https://kubernetes.io/docs/setup/). Personally, we use Google's managed Kubernetes offering, [GKE](https://cloud.google.com/kubernetes-engine).
- [Docker Desktop](https://docs.docker.com/desktop/). So you can build and deploy your container images.
- A [Metabase](https://www.metabase.com/) installation. There are quicker [ways to get up-and-running](https://www.metabase.com/docs/latest/operations-guide/installing-metabase.html), but for our production deployment, we’ve opted to run [Metabase on Kubernetes](https://www.metabase.com/docs/latest/operations-guide/running-metabase-on-kubernetes.html).
- A [Beneath account](https://beneath.dev?noredirect=1). Also [install the SDK](/docs/quick-starts/install-the-sdk/) to follow along.
- A [Lending Club Developer account](https://www.lendingclub.com/developers/api-overview). Get yourself an API key in order to fetch loans from their servers.

Alright, let's start!

## Uploading historical data

The first thing I'd like to do is load historical loan data into Beneath. This historical data will seed our dataset, which we'll soon augment with real-time data. Lending Club provides a bunch of csv files on its website with historical loans and how they've performed over time (i.e. have people paid back their loans or not). I've downloaded the csv files to my local computer.

### Create a Project

The first thing to do before writing data to Beneath is decide on the Project directory that will hold the Stream. With the Beneath CLI, within my user account `epg`, I create a project named `lending-club`:
```bash
beneath project create epg/lending-club
```

### Stage a Stream

Next, we need to create a Stream by defining the stream’s schema and *staging* the stream. In a blank file, we define the schema and name it loans.graphql. The full file is [here](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans/loans.graphql), but this is a short version of what it looks like:

```graphql
" Loans listed on the Lending Club platform. Loans are listed each day at 6AM, 10AM, 2PM, and 6PM (PST). Historical loans include the borrower's payment outcome. "
type Loan
  @stream
  @key(fields: ["id"])
{
  "A unique LC assigned ID for the loan listing."
  id: Int!

  "The date when the borrower's application was listed on the platform."
  list_d: Timestamp

  "The date when the borrower's loan was issued."
  issue_d: Timestamp
  
  "LC assigned loan grade"
  grade: String!

  ... (omitted a few fields) ...

  "Current status of the loan."
  loan_status: String
}

```

Then, we *stage* the stream by providing the stream path in the form of `USERNAME/PROJECT/YOUR-NEW-STREAM-NAME` and the schema file. Here I name the new stream `loans`:

```bash
beneath stream stage epg/lending-club/loans -f loans.graphql
```

### Write csv files to Beneath

Now that the stream has been created on Beneath, we can write data to it. Here's a [simple Python script](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans/load_historical_loans.ipynb) that uploads the data in those csv files to Beneath.

The script ensures that the schema from the input files matches the schema of the Beneath stream. If the schema doesn’t match, Beneath will reject the write.

In case you don't look at the full Python script, here are the lines that matter:
```python
import beneath
client = beneath.Client()
stream = client.find_stream(“epg/lending-club/loans”)
stream.write(data.to_dict(‘records’))
```

### Look at the Beneath Console to validate the writes

Now the data is stored on Beneath. In the Beneath Console, I can double check my work by going to:
- https://beneath.dev/epg/lending-club/loans to browse and query the data I just uploaded
- https://beneath.dev/epg/lending-club/loans/-/monitoring to validate that the correct amount of data was just written


## Writing real-time data

Now that we’ve loaded historical data into Beneath, we want to continually fetch the real-time loan data. Lending Club releases these loans and their data four times each day. 

### Write ETL script
We write a little [ETL script](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans/fetch_new_loans.py) to ping the Lending Club API and write the resulting data to Beneath.

These are the Beneath-relevant lines of Python code for writing data:
```python
import beneath
client = beneath.Client()
stream = await client.find_stream("epg/lending-club/loans")
await stream.write(loans)
```

### Create a Beneath Service

In order to deploy our ETL code and authenticate it to Beneath, we need to create a Beneath [Service](/docs/managing-resources/resources/). A Service has its own secret, its own minimally-viable permissions, and its own quota. We'll create a Lending Club Service, which will solely have access to write to the `epg/lending-club/loans` stream.

We create the service, grant it permissions, and issue its secret with these CLI commands:

```bash
beneath service create epg/lending-club-service --read-quota-mb 100 --write-quota-mb 2000
beneath service update-permissions epg/lending-club-service epg/lending-club/loans --read --write
beneath service issue-secret epg/lending-club-service --description kubernetes
```

Save the secret so that we can register it in the Kubernetes environment.

### Deploy a Kubernetes Cronjob
Our ETL script looks at our Kubernetes environment for two secrets: the Lending Club Service Secret (which we just created), and the Lending Club API Key. We add these secrets to our Kubernetes environment like so:
```bash
kubectl create secret generic lending-club-service-secret -n models --from-literal secret=SECRET
kubectl create secret generic lending-club-api-key -n models --from-literal secret=SECRET
```

Our ETL script only needs to spin-up/spin-down at 6am, 12pm, 3pm, and 6pm PST every day, so this calls for a Cronjob. We can deploy the Cronjob to our Kubernetes cluster with [this yaml file](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans/kube.yaml) and these commands:

```bash
docker build -t gcr.io/beneath/lending-club-loans:latest .
docker push gcr.io/beneath/lending-club-loans:latest
kubectl apply -f loans/kube.yaml -n models
```

Now our ETL script is live! In order to check that everything is wired up correctly, you can run a one-off Cronjob. 

### Go to the Beneath Console to validate the writes

Our script is now running 4x a day, and writing data to Beneath. But let's just check the Console at https://beneath.dev/epg/lending-club/loans to validate that these writes are actually happening. By selecting the `Log` view, we can see the time of each write in the `Time ago` column.

## Connect Metabase

Now that we've uploaded both historical and real-time data to Beneath, let's connect a business intelligence tool so that anyone (inside or outside your organization) can play around with the data.

Assuming you have Metabase installed, you can link it to Beneath by connecting through BigQuery, which is one of the source destinations where Beneath stores your data under-the-hood.

### Add Database

In the admin panel, click “Add Database” and fill out the form with the following values*:

![metabase config](/media/docs/metabase-config.png)

*Currently, if you’d like to connect your own BI tool, you’ll have to [reach out to us](/contact) to provide you with authentication details so you can connect to the Beneath data warehouse. We know this isn’t ideal and are working on making this fully self-service.

### Use Metabase to explore data, run analytics queries, or create a dashboard

Voila! You’ve now connected Metabase to your Beneath data, and you can take advantage of Metabase’s easy-to-use data exploration and it’s friendly SQL interface. Play around with Metabase for 30 minutes and you’ll realize how easy it is. We’ve created [this dashboard](https://metabase.demo.beneath.dev/dashboard/1).

## Train a machine learning model on historical data

After exploring our data, our next step is to enrich the data stream with predictions.

With Beneath, you can train a quick machine learning model by reading your data into a Jupyter notebook and training your model in-memory on your local computer. (This is the quick way to do it -- we’ll cover more robust ways in other tutorials).

You can look at the full training script [here](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-enriched/train_model.ipynb). But here’s the outline of it:

### Read data into a Jupyter notebook

Read in your data to a Jupyter notebook. Your data will be ready-to-go in a Pandas dataframe.

```python
import beneath
client = beneath.Client()
df = await client.easy_read(“epg/lending_club/loans”)
```

### Define features and train your model

Define your input features and your target variable. Split your data into a training set and test set. Train your model.

```python
X = df[['term', 'int_rate', 'loan_amount', 'annual_inc', 
        'acc_now_delinq', 'dti', 'fico_range_high', 'open_acc', 'pub_rec', 'revol_util']]
Y = df[['loan_status_binary']]
X_train, X_test, y_train, y_test = train_test_split(X, Y, test_size=0.3, random_state=2020)
clf = LogisticRegression(random_state=0).fit(X_train, y_train)
```

### Save the model

With the [joblib library](https://joblib.readthedocs.io/en/latest/generated/joblib.dump.html), we can easily and efficiently save a ML model to a `pkl` file:

```python
import joblib
joblib.dump(clf, 'model.pkl')
```

Now that we have our machine learning model in the model.pkl file,  we can include it in our [Dockerfile](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-enriched/Dockerfile) and access it in our Kubernetes environment.

## Enrich the data stream

Next we'll apply our machine learning model to every new Lending Club loan in real-time. We'll get an instant assessment of whether or not the loan is an attractive loan to buy or not.

In a Beneath Service, we'll read in the `epg/lending-club/loans` stream, apply our model, and output a new stream called `epg/lending-club/loans-enriched`.

### Stage the output stream

First, we’ll prepare our output data stream by defining its schema in [this file](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-enriched/loans_enriched.graphql). We’re using the same schema as the raw loans stream, but this time we’re adding a column for our predictions:

```graphql
" Loans listed on the Lending Club platform. Loans are listed each day at 6AM, 10AM, 2PM, and 6PM (PST). Historical loans include the borrower's payment outcome. "
type Loan
  @stream
  @key(fields: ["id"])
{
  "A unique LC assigned ID for the loan listing."
  id:	Int!

  "The date when the borrower's application was listed on the platform."
  list_d:	Timestamp

  "The date when the borrower's loan was issued."
  issue_d: Timestamp
  
  "LC assigned loan grade"
  grade:	String!

  ... (omitted a few fields) ...

  "Current status of the loan."
  loan_status: String

  "Predicted loan status, computed using a logistic regression model"
  loan_status_predicted: String!
}

```

Then we stage the stream:
```bash
beneath stream stage epg/lending-club/loans-enriched -f loans_enriched.graphql
```

### Write stream processing script
Next, we’ll create our stream processing script that we’ll run in a Kubernetes container. It’s called enrich_loans.py and you can find it [here](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-enriched/enrich_loans.py). In particular, note these functions from the Beneath SDK:
```python
TODO
```

### Create a Beneath Service
Even if the compute goes down, the Service will spin back up and continue right where it left off! Under the hood, your progress is stored in Beneath, so you don’t have to worry about it and your Kubernetes deployment needs no storage of its own.

We create the service, grant it permissions, and issue its secret with these CLI commands:

```bash
TODO
```

Save the secret so that we can register it in the Kubernetes environment.


### Deploy to Kubernetes
Like we did for the loans stream, we need to deploy it to Kubernetes. We follow the same steps as before.

We add our Service secret to our Kubernetes environment like so:
```bash
TODO
```

We create our Kubernetes deployment with [this yaml file](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-enriched/kube.yaml) and these commands:
```bash
docker build -t gcr.io/beneath/lending-club-loans-enriched:latest .
docker push gcr.io/beneath/lending-club-loans-enriched:latest
kubectl apply -f loans-enriched/kube.yaml -n models
```

Now our stream processing script is live!

### Go to the Beneath Console to validate the writes
Again, to double-check our work, we can check out the stream at https://beneath.dev/epg/lending-club/loans-enriched to make sure it's writing data as we'd expect.


## Consume predictions with the API options

In the Console, head on over to the stream's API tab at https://beneath.dev/epg/lending-club/loans-enriched/-/api. You can consume your predictions with any of the available API options. Or, within, our enrich_loans.py script we can issue buy orders based on the predictions. 

## We're done!

We've built an end-to-end analytics application. We leveraged Kubernetes for deploying our containers, we leveraged Metabase to allow other people to easily visualize and query the data, and we centered our application around Beneath for its data storage and stream processing.

Head on over to the [Console](https://beneath.dev?noredirect=1) to explore some of the public projects or [reach out to us](/contact) if you'd like to talk through your next project. We're very happy to help you get going.
