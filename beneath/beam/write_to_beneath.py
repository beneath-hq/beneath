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

  _BATCH_SIZE = 1000

  def __init__(self, stream, instance_id):
    if stream is None:
      raise Exception("Error! The provided stream is not valid")
    if instance_id is None:
      raise Exception("Error! The provided instance ID is not valid")
    self.stream = stream
    self.instance_id = instance_id
    self.bundle = None

  def __getstate__(self):
    return {
      "stream": self.stream,
      "instance_id": self.instance_id,
      "bundle": self.bundle,
    }

  def __setstate__(self, obj):
    self.stream = obj["stream"]
    self.instance_id = obj["instance_id"]
    self.bundle = obj["bundle"]

  def start_bundle(self):
    self.bundle = []

  def process(self, row):
    self.bundle.append(row)
    if len(self.bundle) >= self._BATCH_SIZE:
      self._flush()

  def finish_bundle(self):
    self._flush()

  def _flush(self):
    if len(self.bundle) != 0:
      self.stream.write(records=self.bundle, instance_id=self.instance_id)
      self.bundle = []

class WriteToBeneath(beam.PTransform):
  def __init__(self, stream, instance_id=None):
    self.stream = stream
    self.instance_id = instance_id if instance_id else stream.current_instance_id

  def expand(self, pvalue):
    p = (
      pvalue
      | 'Write' >> beam.ParDo(_GatewayWriteFn(self.stream, self.instance_id))
    )
    return p
