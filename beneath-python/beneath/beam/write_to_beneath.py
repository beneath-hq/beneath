"""
Beneath connector

This module implements writing the results of a Beam pipeline to Beneath's servers.

Here is an example of WriteToBeneath's usage in a Beam pipeline:

  # set up beneath client
  client = Client(secret='SECRET')

  # get the stream of interest
  stream = client.get_stream("PROJECT_NAME", "STREAM_NAME")

  # create pipeline
  pipeline = beam.Pipeline()
  (p
    | beam.Create(self._generate())
    | WriteToBeneath(stream)
  )

  # run pipeline
  pipeline.run()
"""

import apache_beam as beam

class _GatewayWriteFn(beam.DoFn):
  def __init__(self, stream):
    if stream is None:
      raise Exception("Error! The provided stream is not valid")
    self.stream = stream
    self.bundle = None

  def __getstate__(self):
    return {
      "stream": self.stream,
      "bundle": self.bundle,
    }

  def __setstate__(self, obj):
    self.stream = obj["stream"]
    self.bundle = obj["bundle"]

  def start_bundle(self):
    self.bundle = []

  def process(self, row):
    self.bundle.append(row)

  def finish_bundle(self):
    self.stream.write_records(self.stream.current_instance_id, self.bundle)
    self.bundle = None


class WriteToBeneath(beam.PTransform):
  def __init__(self, stream):
    self.stream = stream

    # check to see if the stream is a batch or stream (aka the data is bounded or unbounded)
    if self.stream.batch == True:
      # TODO: if batch, 1) "prepare new batch" call gRPC, get back a "new" instance id
      # 2) add data to that instance
      # 3) promote the instance to be the "current instance"
      pass
    else:
      # if streaming, lookup and write to current instance id (as is currently happening)
      pass

  def expand(self, pvalue):
    stream = self.stream
    return (
      pvalue
      | beam.ParDo(_GatewayWriteFn(stream))
    )
