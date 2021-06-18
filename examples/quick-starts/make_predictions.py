import beneath
import pickle

PREDICTIONS_SCHEMA = """
  type Predictions @schema {
    id: Int! @key
    prediction: Float!
  }
"""

# create a Beneath Pipeline
# we define it globally so we can use p.client.checkpointer() in get_clf()
p = beneath.Pipeline(parse_args=True, disable_checkpoints=True)

# rudimentary cache to store the predictive model
_clf = None

# load the ML model from Beneath
async def get_clf():
    global _clf
    if _clf is None:
        checkpointer = await p.client.checkpointer(
            project_path="USERNAME/PROJECT_NAME",
            metatable_name="predictive-model",
        )

        s = await checkpointer.get("clf_serialized")
        _clf = pickle.loads(s)
    return _clf


# use the ML model to make predictions and emit records that conform to the PREDICTIONS_SCHEMA
async def predict_outcome(record):
    clf = await get_clf()
    X = [[record["FEATURE_1"], record["FEATURE_1"], ...]]
    prediction = clf.predict_proba(X)[0][0]
    yield {
        "id": record["id"],
        "prediction": prediction,
    }


if __name__ == "__main__":
    p.description = "A pipeline that makes predictions based on the features table"

    # consume the features table
    features = p.read_table("USERNAME/PROJECT_NAME/features")

    # derive a new table and write to Beneath
    predictions = p.apply(features, predict_outcome)
    p.write_table(
        predictions,
        "predictions",
        schema=PREDICTIONS_SCHEMA,
    )

    # run the pipeline
    p.main()
