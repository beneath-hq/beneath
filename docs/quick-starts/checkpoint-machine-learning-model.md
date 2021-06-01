---
title: Checkpoint a machine learning model
description: A guide to saving and loading a machine learning model
menu:
  docs:
    parent: quick-starts
    weight: 700
weight: 700
---

As a [unified data system]({{< ref "/docs/concepts/unified-data-system" >}}), Beneath is especially suited to predictive analytics. You can train a machine learning model on historical data, then apply the model to streaming data to make predictions in real-time.

After training your model, you'll need somewhere to save it. This quick start shows how to use Beneath checkpointing to save and load your model.

## Install the Beneath SDK

If you haven't already, follow the [Install the Beneath SDK]({{< ref "/docs/quick-starts/install-sdk" >}}) quick start to install and authenticate Beneath on your local computer.

# Part 1: Create and save a machine learning model

To start, load data from Beneath and train a machine learning model using the tool of your choice.

## Train your model

Here we load our training data into a notebook:

```python
import beneath

features = await beneath.load_full("USERNAME/PROJECT_NAME/features")
outcomes = await beneath.load_full("USERNAME/PROJECT_NAME/outcomes")

X_train, y_train = ... # create a training set
```

And use [`sklearn`](https://scikit-learn.org/stable/) to fit a classifier:

```python
from sklearn import LogisticRegression

clf = LogisticRegression().fit(X_train, y_train)
```

## Serialize your model

Before checkpointing, the model object first needs to be converted into a data format that can be transmitted and stored (the process known as [serialization](https://en.wikipedia.org/wiki/Serialization)). Here we use [Python's `pickle` module](https://realpython.com/python-pickle-module/) to serialize our `clf` model into a byte string:

```python
import pickle

s = pickel.dumps(clf)
```

## Checkpoint your model

Beneath checkpoints are metadata that can be retrieved whenever a data processor starts up. At the beginning of any machine learning application, the first step is to load the model into memory before processing new data.

To use checkpoints, first establish a connection to Beneath:

```python
client = beneath.Client()
await client.start()
```

Next create a "checkpointer." Here we create a metastream named `predictive-model` and save ("set") our serialized classifier:

```python
checkpointer = await client.checkpointer(
  project_path="USERNAME/PROJECT_NAME",
  metastream_name="predictive-model",
  metastream_description="Stores the model used to make predictions"
)

await checkpointer.set("clf_serialized", s)
```

When we're done with the checkpointer, we close the connection:

```python
client.stop()
```

# Part 2: Make predictions

Now that we've built and saved our machine learning model, let's turn to the pipeline where we'll call the model to make predictions. We'll walk through the script in sections, but here it is in full: [make_predictions.py](https://github.com/beneath-hq/beneath/blob/master/examples/quick-starts/checkpoint-machine-learning-model/make_predictions.py).

If you haven't yet familiarized yourself with Beneath pipelines, you should start with the [Create a Pipeline quickstart]({{< ref "/docs/quick-starts/create-pipeline" >}}).

## Create a Beneath pipeline

At the top of the file, we create a Beneath pipeline and define the schema for our stream of predictions:

```python
import beneath
import pickle

PREDICTIONS_SCHEMA = """
  type Predictions @schema {
    id: Int! @key
    prediction: Float!
  }
"""

# create a Beneath Pipeline
# we define it globally so we can use p.client.checkpointer()
# in get_clf()
p = beneath.Pipeline(parse_args=True, disable_checkpoints=True)
```

## Fetch and deserialize the model

Next we create a function to fetch and deserialize the machine learning model from Beneath.

We use the checkpointer to fetch ("get") the model. Note that every Beneath pipeline embeds a client, so we don't need to create a client independently, like we did when we saved the model in Part 1.

Deserializing the model converts the byte string back into the Python classifier object we originally created with `sklearn`. We use `pickle`, which we imported above, to deserialize the model.

```python
# rudimentary cache to store the predictive model
_clf = None

# load the ML model from Beneath
async def get_clf():
    global _clf
    if _clf is None:
        checkpointer = await p.client.checkpointer(
            project_path="USERNAME/PROJECT_NAME",
            metastream_name="predictive-model",
        )

        s = await checkpointer.get("clf_serialized")
        _clf = pickle.loads(s)
    return _clf
```

## Define a function to make predictions

Now that we've loaded our machine learning model, we can use it in a function to make predictions:

```python
# use the ML model to make predictions and emit records that
# conform to the PREDICTIONS_SCHEMA
async def predict_outcome(record):
    clf = await get_clf()

    X = [[record["FEATURE_1"], record["FEATURE_1"], ...]]
    prediction = clf.predict_proba(X)[0][0]

    yield {
        "id": record["id"],
        "prediction": prediction,
    }
```

## Create the pipeline to make predictions

Here we construct the Beneath pipeline that reads our stream of features, applies the `predict_outcome` function, and outputs the results to a new stream named `predictions`:

```python
if __name__ == "__main__":
    p.description = "A pipeline that makes predictions \
    based on the features stream"

    # consume the features stream
    features = p.read_stream("USERNAME/PROJECT_NAME/features")

    # derive a new stream and write to Beneath
    predictions = p.apply(features, predict_outcome)
    p.write_stream(
        predictions,
        "predictions",
        schema=PREDICTIONS_SCHEMA,
    )

    # run the pipeline
    p.main()
```

Run the pipeline, then check out the web console to see the predictions stream in!

# More info

Find all the code for this walkthrough in the [quickstart repo](https://github.com/beneath-hq/beneath/blob/master/examples/quick-starts/checkpoint-machine-learning-model/). For more info about the Beneath client checkpointer, check out the [Python client API reference](https://python.docs.beneath.dev/client.html#beneath.Client.checkpointer).
