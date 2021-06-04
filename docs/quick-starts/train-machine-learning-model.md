---
title: Train a machine learning model
description: A guide to training and saving a machine learning model
menu:
  docs:
    parent: quick-starts
    weight: 600
weight: 600
---

As a [unified data system]({{< ref "/docs/concepts/unified-data-system" >}}), Beneath is especially suited to predictive analytics. You can train a machine learning model on historical data, save the model, then apply the model to streaming data to make predictions in real-time.

This quick start shows how to train a model and use Beneath checkpoints to save it. Here's the associated code: [train_model.ipynb](https://github.com/beneath-hq/beneath/blob/master/examples/quick-starts/train_model.ipynb)

## Install the Beneath SDK

If you haven't already, follow the [Install the Beneath SDK]({{< ref "/docs/quick-starts/install-sdk" >}}) quick start to install and authenticate Beneath on your local computer.

## Train your model

To start, load data from Beneath and train a machine learning model using the tool of your choice. Here we load our training data into a notebook:

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

The next step is to convert the model object into a data format that can be transmitted and stored. Here we use [Python's `pickle` module](https://realpython.com/python-pickle-module/) to serialize our `clf` model into a byte string:

```python
import pickle

s = pickle.dumps(clf)
```

## Checkpoint your model

In Beneath, checkpoints are metadata that can be retrieved whenever a data processor starts up. At the beginning of any machine learning application, the first step is to load the model into memory before processing new data.

To use checkpoints, first establish a connection to Beneath:

```python
client = beneath.Client()
await client.start()
```

Next create a "[checkpointer](https://python.docs.beneath.dev/client.html#beneath.Client.checkpointer)" and save our serialized classifier to it:

```python
checkpointer = await client.checkpointer(project_path="USERNAME/PROJECT_NAME")

await checkpointer.set("clf_serialized", s)
```

When we're done with the checkpointer, we close the connection:

```python
await client.stop()
```

## Next steps

Your new machine learning model is now saved to Beneath. The next quick start shows how to load the model and apply it to a data stream.
