---
title: Apply a machine learning model
description: A guide to loading and applying a machine learning model
menu:
  docs:
    parent: quick-starts
    weight: 700
weight: 700
---

This quickstart builds on the previous one to [train and save a machine learning model]({{< ref "/docs/quick-starts/train-machine-learning-model" >}}). Here we cover the next steps: load the model and apply it to a data stream.

Specifically, we'll load our model from a Beneath checkpoint, and use the Pipeline API to derive a new stream. We'll walk through the pipeline script in sections, but here it is in full: [make_predictions.py](https://github.com/beneath-hq/beneath/blob/master/examples/quick-starts/make_predictions.py)

If you haven't yet familiarized yourself with Beneath pipelines, you should start with the quickstart [Use the Pipeline API]({{< ref "/docs/quick-starts/use-pipeline-api" >}}).

## Initialize a Beneath pipeline

At the top of the file, we initialize a Beneath pipeline and define the schema for a new stream of predictions:

```python
import beneath
import pickle

PREDICTIONS_SCHEMA = """
  type Predictions @schema {
    id: Int! @key
    prediction: Float!
  }
"""

# initialize a Beneath Pipeline
p = beneath.Pipeline(parse_args=True, disable_checkpoints=True)
```

## Load and deserialize the model

Next we create a function to load and deserialize the model, which we saved to Beneath in the previous quickstart.

Again we use a Beneath [checkpointer](https://python.docs.beneath.dev/client.html#beneath.Client.checkpointer), but this time to load the model. Every Beneath pipeline embeds a client so the checkpointer is easy to access.

Deserializing the model converts the byte string back into the Python classifier object originally created with `sklearn`. We use `pickle`, which we imported above, to deserialize the model.

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

## Define a function to call the machine learning model

Now that we've loaded our machine learning model, we can use it in a function to make predictions:

```python
# use the ML model to make predictions and emit records that
# conform to the PREDICTIONS_SCHEMA
async def predict_outcome(record):
    clf = await get_clf()

    X = [[record["FEATURE_1"], record["FEATURE_2"], ...]]
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
