"""
Beneath connector

This module implements writing the results of a Beam pipeline to Beneath's servers.

Here is an example of WriteToBeneath's usage in a Beam pipeline:

  # set up beneath client
  client = Client(secret='SECRET')

  # get the table of interest
  table = client.get_table("PROJECT_NAME", "TABLE_NAME")

  # create pipeline
  pipeline = beam.Pipeline()
  (p
    | beam.Create(self._generate())
    | WriteToBeneath(table)
  )

  # run pipeline
  pipeline.run()
"""

import apache_beam as beam


class _GatewayWriteFn(beam.DoFn):

    _BATCH_SIZE = 1000

    def __init__(self, table, instance_id):
        if table is None:
            raise Exception("Error! The provided table is not valid")
        if instance_id is None:
            raise Exception("Error! The provided instance ID is not valid")
        self.table = table
        self.instance_id = instance_id
        self.bundle = None

    def __getstate__(self):
        return {
            "table": self.table,
            "instance_id": self.instance_id,
            "bundle": self.bundle,
        }

    def __setstate__(self, obj):
        self.table = obj["table"]
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
            self.table.write(records=self.bundle, instance_id=self.instance_id)
            self.bundle = []


class WriteToBeneath(beam.PTransform):
    def __init__(self, table, instance_id=None):
        self.table = table
        self.instance_id = instance_id if instance_id else table.current_instance_id

    def expand(self, pvalue):
        p = pvalue | "Write" >> beam.ParDo(_GatewayWriteFn(self.table, self.instance_id))
        return p
