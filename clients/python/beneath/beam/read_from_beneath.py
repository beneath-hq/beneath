"""
  This module implements reading data from Beneath into a Beam pipeline
"""

import apache_beam as beam


class ReadFromBeneath(beam.PTransform):
    def __init__(self, table):
        self.table = table

    def expand(self, pvalue):
        table = self.table
        query = "select * from `{}`".format(table.bigquery_table)
        source = beam.io.BigQuerySource(query=query, use_standard_sql=True)
        # can probably get query schema with source.schema
        return pvalue | beam.io.Read(source)
