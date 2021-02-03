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

- A stream for [historical loans with performance data](https://beneath.dev/epg/lending-club/stream:loans-history)
- A stream for [real-time loans listed on the Lending Club website 4x per day](https://beneath.dev/epg/lending-club/stream:loans)
- A stream for [real-time loans enriched with performance predictions](https://beneath.dev/epg/lending-club/stream:loans-enriched).
- Then we'll create this [Metabase dashboard](https://metabase.demo.beneath.dev/dashboard/1).

For reference, here's all [the code](https://gitlab.com/beneath-hq/beneath/-/tree/master/clients/python/examples/lending-club).

## Prerequisites

In the interest of getting-to-the-point and focusing on Beneath, we'll assume you can get set-up with the following:

- A [Kubernetes](https://kubernetes.io/docs/home/) cluster. There are [many ways to set up Kubernetes](https://kubernetes.io/docs/setup/). Personally, we use Google's managed Kubernetes offering, [GKE](https://cloud.google.com/kubernetes-engine).
- [Docker Desktop](https://docs.docker.com/desktop/). So you can build and deploy your container images.
- A [Metabase](https://www.metabase.com/) installation. There are quicker [ways to get up-and-running](https://www.metabase.com/docs/latest/operations-guide/installing-metabase.html), but for our production deployment, we’ve opted to run [Metabase on Kubernetes](https://www.metabase.com/docs/latest/operations-guide/running-metabase-on-kubernetes.html).
- A [Beneath account](https://beneath.dev?noredirect=1). Also [install the SDK](/docs/quick-starts/install-the-sdk/) to follow along.
- A [Lending Club Developer account](https://www.lendingclub.com/developers/api-overview). Get yourself an API key in order to fetch loans from their servers.

Alright, let's start!

## Uploading historical data

The first thing I'd like to do is load historical loan data into Beneath. Lending Club provides a bunch of csv files on its website with historical loans and how they've performed over time (i.e. have people paid back their loans or not). I've downloaded the csv files to my local computer.

Our goal is to explore this data in the business intelligence tool that we'll connect, and we'll train a machine learning model to predict loan performance.

### Create a Project

The first thing to do before writing data to Beneath is decide on the Project directory that will hold the Stream. With the Beneath CLI, within my user account `epg`, I create a project named `lending-club`:

```bash
beneath project create epg/lending-club
```

### Create a Stream

Next, we need to create a Stream by defining the stream’s schema, _staging_ the stream, and _staging_ an _instance_. In a blank file, we define the schema and name it `loans_history.graphql`. The full file is [here](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-history/loans_history.graphql), but this is a short version of what it looks like:

```graphql
" Loans listed on the Lending Club platform. Loans are listed each day at 6AM, 10AM, 2PM, and 6PM (PST). Historical loans include the borrower's payment outcome. "
type Loan
  @stream
  @key(fields: ["id"])
{
  "A unique LC assigned ID for the loan listing."
  id: Int!

  "The date when the borrower's loan was issued."
  issue_d: Timestamp

  "LC assigned loan grade"
  grade: String!

  ... (omitted a few fields) ...

  "Current status of the loan."
  loan_status: String
}

```

Then, we _stage_ the stream by providing the stream path in the form of `USERNAME/PROJECT/YOUR-NEW-STREAM-NAME` and the schema file. Here I name the new stream `loans-history`:

```python
import beneath
client = beneath.Client()

username = "epg"
project_name = "lending-club"
stream_name = "loans-history"

STREAM_PATH = f"{username}/{project_name}/{stream_name}"
SCHEMA = open("loans_history.graphql", "r").read()

stream = await client.create_stream(
    stream_path=STREAM_PATH,
    schema=SCHEMA,
    update_if_exists=True,
)
```

Lastly, streams have _instances_, which allow for versioning. We actually write data, not to a stream, but to a stream instance. We stage our first stream instance like so:

```python
instance = await stream.create_instance(version=0, make_primary=True, update_if_exists=True)
```

### Write csv files to Beneath

Now that the stream has been created on Beneath, we can write data to it. Here's a [simple Python script](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-history/load_historical_loans.ipynb) that uploads the data in those csv files to Beneath.

The script ensures that the schema from the input files matches the schema of the Beneath stream. If the schema doesn’t match, Beneath will reject the write.

This is the part of the script that performs the actual write:

```python
async with instance.writer() as w:
  await w.write(data.to_dict('records'))
```

### Look at the Beneath Console to validate the writes

Now the data is stored on Beneath. In the Beneath Console, I can double check my work by going to:

- https://beneath.dev/epg/lending-club/stream:loans-history to browse and query the data I just uploaded
- https://beneath.dev/epg/lending-club/stream:loans-history/-/monitoring to validate that the correct amount of data was just written

## Writing real-time data

Now that we’ve loaded historical data into Beneath, we want to continually fetch the real-time loan data. Lending Club releases these loans and their data four times each day.

### Write ETL script

We write a little [ETL script](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans/fetch_new_loans.py) to ping the Lending Club API and write the resulting data to Beneath.

This script revolves around a Beneath _generator_ which is defined in this snippet:

```python
beneath.easy_generate_stream(
  generate_fn=generate_loans,
  output_stream_path=STREAM,
  output_stream_schema=SCHEMA,
)
```

A Beneath stream _generator function_ does a few things: keeps track of state, performs ETL logic like calling the API of an external data source, and yields records.

The ability to keeping track of state is super handy. Here, we keep track of which is the most recently listed loan that we've processed:

```python
await p.set_checkpoint("latest_list_d", max_list_d.isoformat())
```

And we get the state like so:

```python
latest_list_d = await p.get_checkpoint("latest_list_d", default="1970-01-01T00:00:00.000000+00:00")
```

Keeping this state ensures that, in the event of failure, we don't write repeat data to Beneath. It allows us to stay truly _synced_ with our external data source.

We ping the Lending Club API here:

```python
headers = {"Authorization": LENDING_CLUB_API_KEY}
params = {"showAll": "true"}
req = requests.get("https://api.lendingclub.com/api/investor/v1/loans/listing", headers=headers, params=params)
json = req.json()
```

And because the `easy_generate_stream` function is a generator, we must `yield` (not `return`) records to be written to Beneath:

```python
yield new_loans
```

### Create a Beneath Service

In order to deploy our ETL code and authenticate it to Beneath, we need to create a Beneath [Service](/docs/managing-resources/resources/). A Service has its own secret, its own minimally-viable permissions, and its own quota. We'll create a `fetch-new-loans` service, which will solely have access to write to the `epg/lending-club/loans` stream.

We stage the service and issue its secret with these CLI commands:

```bash
python ./fetch-new-loans.py stage epg/lending-club/fetch-new-loans --read-quota-mb 10000 --write-quota-mb 10000
beneath service issue-secret epg/lending-club/fetch-new-loans --description kubernetes
```

Save the secret so that we can register it in the Kubernetes environment.

### Deploy a Kubernetes Cronjob

Our ETL script looks at our Kubernetes environment for two secrets: the `fetch-new-loans` service secret (which we just created), and the Lending Club API Key. We add these secrets to our Kubernetes environment (with names that match what's in our `kube.yaml` file in the next step) like so:

```bash
kubectl create secret generic lending-club-loans-service-secret -n models --from-literal secret=SECRET
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

Our script is now running 4x a day, and writing data to Beneath. But let's just check the Console at https://beneath.dev/epg/lending-club/stream:loans to validate that these writes are actually happening. By selecting the `Log` view, we can see the time of each write in the `Time ago` column.

## Connect Metabase

Now that we've uploaded both historical and real-time data to Beneath, let's connect a business intelligence tool so that we can easily explore the data. Additionally, Metabase allows us to make dashboards public, so we can share them with teammates or outside parties.

Assuming you have Metabase installed, you can link it to Beneath by connecting through BigQuery, which is one of the source destinations where Beneath stores your data under-the-hood.

### Add Database

In the admin panel, click “Add Database” and fill out the form with the following values\*:

![metabase config](/media/docs/metabase-config.png)

\*Currently, if you’d like to connect your own BI tool, you’ll have to [reach out to us](/contact) to provide you with authentication details so you can connect to the Beneath data warehouse. We know this isn’t ideal and are working on making this fully self-service.

### Use Metabase to explore data, run analytics queries, or create a dashboard

Voila! You’ve now connected Metabase to your Beneath data, and you can take advantage of Metabase’s easy-to-use data exploration and it’s friendly SQL interface. Play around with Metabase for 30 minutes and you’ll realize how easy it is. We’ve created [this dashboard](https://metabase.demo.beneath.dev/dashboard/1).

## Train a machine learning model on historical data

After exploring our data, our next step is to enrich our real-time loan data stream with predictions.

With Beneath, you can train a quick machine learning model by reading your data into a Jupyter notebook and training your model in-memory on your local computer. (This is the quick way to do it -- we’ll cover more robust ways in other tutorials).

You can look at the full training script [here](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-enriched/train_model.ipynb). But here’s the outline of it:

### Read data into a Jupyter notebook

Read in your data to a Jupyter notebook. Your data will be ready-to-go in a Pandas dataframe.

```python
import beneath
df = await beneath.easy_read(“epg/lending_club/loans”)
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

Now that we have our machine learning model in the model.pkl file, we can include it in our [Dockerfile](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-enriched/Dockerfile) and access it in our Kubernetes environment.

## Enrich the data stream

Next we'll apply our machine learning model to every new Lending Club loan in real-time. This will give us an instant assessment of whether or not the loan is an attractive loan to buy.

In a Beneath service, we'll read the `epg/lending-club/loans` stream, apply our model, and output a new stream called `epg/lending-club/loans-enriched`.

### Define new schema

First, we prepare our output data stream by defining its schema in [this file](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-enriched/loans_enriched.graphql). We’re using the same schema as the raw loans stream, but this time we’re adding a column for our predictions.

### Write stream processing script

Next, we create our stream processing script that we’ll run in a Kubernetes container. It’s called enrich_loans.py and you can find it [here](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-enriched/enrich_loans.py).

The script revolves around the `easy_derive_stream` function, which reads a Beneath stream, performs a computation on the stream, and outputs another Beneath stream:

```python
beneath.easy_derive_stream(
  input_stream_path=INPUT_STREAM,
  apply_fn=process_loan,
  output_stream_path=OUTPUT_STREAM,
  output_stream_schema=OUTPUT_SCHEMA,
)
```

Take a look in the `process_loan` function, and you'll see that we make our prediction from the machine learning model like so:

```python
clf = joblib.load('model.pkl')
y_pred = clf.predict(X)[0]
```

Further in the `process_loan` function, after formatting the `enriched_loan` record to match the schema we designed for the output stream, we _yield_ the record:

```python
yield enriched_loan
```

The script is simple! Now we need to deploy it to Kubernetes via a Beneath service.

### Create a Beneath Service

We create the service and issue its secret with these CLI commands:

```bash
python ./loans-enriched/enrich-loans.py stage epg/lending-club/enrich-loans --read-quota-mb 10000 --write-quota-mb 10000
beneath service issue-secret epg/lending-club/enrich-loans --description kubernetes
```

Save the secret so that we can register it in the Kubernetes environment.

### Deploy to Kubernetes

Like we did for the loans stream, we need to deploy it to Kubernetes. We follow the same steps as before.

We add our service secret to our Kubernetes environment like so:

```bash
kubectl create secret generic lending-club-loans-enriched-service-secret -n models --from-literal secret=SECRET
```

Now we need to create [this Dockerfile](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-enriched/Dockerfile). We specify that the service uses the `delta` strategy. With the delta strategy, every time Kubernetes spins up the service, the service will process all records that it hasn't seen before. Then the service will shut down until the Kubernetes cronjob spins it back up.

Here's the important line:

```dockerfile
CMD ["python", "enrich_loans.py", "run", "epg/lending-club/enrich-loans", "--strategy", "delta", "--read-quota-mb", "1000", "--write-quota-mb", "1000"]
```

We create our Kubernetes cronjob with [this yaml file](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/lending-club/loans-enriched/kube.yaml) and these commands:

```bash
docker build -t gcr.io/beneath/lending-club-loans-enriched:latest .
docker push gcr.io/beneath/lending-club-loans-enriched:latest
kubectl apply -f loans-enriched/kube.yaml -n models
```

Now our stream processing service is live! In Kubernetes, you can trigger a one-off job to ensure that the script is working as designed.

### Go to the Beneath Console to validate the writes

Again, to double-check our work, we can check out the stream at https://beneath.dev/epg/lending-club/stream:loans-enriched to make sure it's writing data as we'd expect.

## Consume predictions with the API options

In the Console, head on over to the stream's API tab at https://beneath.dev/epg/lending-club/stream:loans-enriched/-/api. You can consume your predictions with any of the available API options. Or you could issue conditional buy orders within the enrich_loans.py script!

## We're done!

We've built an end-to-end analytics application. We leveraged Kubernetes for deploying our containers, we leveraged Metabase to allow other people to easily visualize and query the data, and we centered our application around Beneath for its data storage and stream processing library.

Head on over to the [Console](https://beneath.dev?noredirect=1) to explore some of the public projects on Beneath or [reach out to us](/contact) if you'd like to talk through your next project. We're very happy to help you get going.
