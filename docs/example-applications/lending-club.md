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
- A [Metabase](https://www.metabase.com/) installation. There are quicker [ways to get up-and-running](https://www.metabase.com/docs/latest/operations-guide/installing-metabase.html), but for our production deployment, we’ve opted to run [Metabase on Kubernetes](https://www.metabase.com/docs/latest/operations-guide/running-metabase-on-kubernetes.html).
- A [Beneath account](https://beneath.dev?noredirect=1). Also [install the SDK](docs/quick-starts/install-the-sdk/) to follow along.
- A [Lending Club Developer account](https://www.lendingclub.com/developers/api-overview). Get yourself an API key in order to fetch loans from their servers.

Alright, let's start!

## Uploading historical data

The first thing I'd like to do is load historical loan data into Beneath. This historical data will seed our dataset, which we'll soon augment with real-time data. Lending Club provides a bunch of csv files on its website with historical loans and how they've performed over time (i.e. have people paid back their loans or not). I've downloaded the csv files to my local computer.

### Create a Project

The first thing to do before writing data to Beneath is decide on the Project that will hold the Stream. With the Beneath CLI, within my user account `epg`, I create a project named `lending-club`:
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

### Look at the Beneath Terminal to validate the writes

Now the data is stored on Beneath. In the Beneath Terminal, I can double check my work by going to:
- https://beneath.dev/epg/lending-club/loans to browse and query the data I just uploaded
- https://beneath.dev/epg/lending-club/loans/-/monitoring to validate that the correct amount of data was just written


## Writing real-time data

Now that we’ve loaded historical data into Beneath, we want to get all the real-time loan data. Lending Club releases these loans and their data four times each day. So, we write a little ETL script to ping the Lending Club API and write the resulting data to Beneath.

### Create a Beneath Service

Using Beneath, any piece of code that you deploy to run continuously or on a certain schedule is considered a Service. A Service runs ‘in the wild’ and, accordingly, should use its own secret, should leverage its own minimally-viable permissions, and should be granted its own quota. All this ensures that, if a bad actor finds your secret, they won’t be able to abuse it for their own purposes.

We create the service, grant it permissions, and issue its secret with these CLI commands:

```bash
beneath service create epg/lending-club-service --read-quota-mb 100 --write-quota-mb 2000
beneath service update-permissions epg/lending-club-service epg/lending-club/loans --read --write
beneath service issue-secret epg/lending-club-service --description kubernetes
```

### Create a Kubernetes Cronjob

My ETL script doesn’t have to run continuously, but only spin-up/spin-down at 6am, 12pm, 3pm, and 6pm PST every day. So this calls for a Cronjob, which we can deploy to our Kubernetes cluster with this yaml file and these commands:

```bash
docker build -t gcr.io/beneath/lending-club-loans:latest .
docker push gcr.io/beneath/lending-club-loans:latest
kubectl apply -f loans/kube.yaml -n models
```

Whenever the script spins up, it will look for our secrets in the Kubernetes environment, so we add them like so:
```bash
kubectl create secret generic lending-club-service-secret -n models --from-literal secret=SECRET
kubectl create secret generic lending-club-api-key -n models --from-literal secret=SECRET
```

### Go to the Beneath Terminal to validate the writes

Awesome. Now our script is running 4x a day, and writing data to Beneath. Again, we can check out the stream in the Terminal to validate that these writes are actually happening. By selecting the Log view (which is the default), we can see the time of each write in the “Time ago” column.

## Connect Metabase

Once you have Metabase installed, you can connect it to your data stored on Beneath by connecting through BigQuery, which is one of the source destinations where Beneath stores your data under-the-hood.

### Add Database

In the admin panel, click “Add Database” and fill out the form with the following values:

![metabase config](/media/docs/metabase-config.png)

*Currently, if you’d like to connect your own BI tool, you’ll have to reach out to us to provide you with authentication details so you can connect to the Beneath data warehouse. We know this isn’t ideal and are working on making this fully self-service.

### Use Metabase to explore data, run analytics queries, or create a dashboard

Voila! You’ve now connected Metabase to your Beneath data, and you can take advantage of Metabase’s easy-to-use data exploration and it’s friendly SQL interface. Play around with Metabase for 30 minutes and you’ll realize how easy it is. We’ve created this dashboard.

## Train a machine learning model on the Lending Club historical data

After exploring our data, our next step is to enrich the data stream with predictions.

With Beneath, you can train a quick machine learning model by reading your data into a Jupyter notebook and training your model in-memory on your local computer. (This is the quick way to do it -- we’ll cover more robust ways in other tutorials).

You can look at the full training script here. But here’s the outline of it:

Read in your data to a Jupyter notebook. Your data will be ready-to-go in a Pandas dataframe.

```python
import beneath
client = beneath.Client()
df = await client.easy_read(“epg/lending_club/loans”)
```

Define your input features and your target variable. Split your data into a training set and test set. Train your model.

```python
X = df[['term', 'int_rate', 'loan_amount', 'annual_inc', 
        'acc_now_delinq', 'dti', 'fico_range_high', 'open_acc', 'pub_rec', 'revol_util']]
Y = df[['loan_status_binary']]
X_train, X_test, y_train, y_test = train_test_split(X, Y, test_size=0.3, random_state=2020)
clf = LogisticRegression(random_state=0).fit(X_train, y_train)
```
Save the model

```python
import joblib
joblib.dump(clf, 'model.pkl')
```

Now that we have our machine learning model (in the form of the model.pkl file), we can apply it to every new loan that’s issued on the Lending Club platform.

## Apply the ML model to every new loan listing to predict which loans will default and should be avoided and which loans will not default and are safe to buy. We deploy the model to Kubernetes.

With Beneath, you create and deploy a Service for your stream processing needs. The Service, deployed to Kubernetes, will read from a Beneath stream, apply a function, then write to an output Beneath stream. The beauty of Services is that, with Beneath’s Python SDK, they are entirely stateless: you don’t have to worry about keeping in sync with your input data stream, even if your compute goes down.

First, we’ll prepare our output data stream by defining its schema. We’re using the same schema as the raw loans stream, but this time we’re adding a column for our predictions:

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

  …

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

Next, we’ll create our file that we’ll run in a Kubernetes container. It’s called enrich_loans.py and you can find it here. In particular, note these functions from the Beneath SDK:
```python
client.process_forever()
```
Even if the compute goes down, the Service will spin back up and continue right where it left off! Under the hood, your progress is stored in Beneath, so you don’t have to worry about it and your Kubernetes deployment needs no storage of its own.

Like we did for the loans stream, we need to deploy it to Kubernetes. Again, we need to create a Beneath Service, issue its secret, and store it in the Kubernetes environment.


## REST API allows you to consume the results of each prediction

In another application, we can consume the results of the loans_enriched stream. Or, within, our enrich_loans.py script we can issue buy orders based on the predictions. 



## Conclusion

End-to-end application.

