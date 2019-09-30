import base64
from datetime import datetime
import apache_beam as beam

"""
  hex: string of format "0x..." -> bytes
  base64: string of base64 -> bytes
  numeric: string containing an integer number(it actually might be an int) -> int
  timestamp: string containing something like "2019-08-06 22:25:53 UTC" -> datetime
"""

class ConvertFromJSONTypes(beam.PTransform):
  def __init__(self, params):
    self.params = params

  def convert_to_bytes(self, element):
    # decode hexadecimal
    if element[0:2] == '0x':
      return bytes.fromhex(element[2:])
    # decode base64
    else:
      return base64.b64decode(element)

  def convert_to_int(self, element):
    return int(element)

  def convert_to_ts(self, element):
    return datetime.strptime(element, '%Y-%m-%d %H:%M:%S.%f %Z')

  def _convert_types(self, row):
    for field, to_type in self.params.items():
      if to_type == "bytes":
          row[field] = self.convert_to_bytes(row[field])
      if to_type == "int":
          row[field] = self.convert_to_int(row[field])
      if to_type == "timestamp":
          row[field] = self.convert_to_ts(row[field])
    return row

  def expand(self, pvalue):   
    return (
      pvalue 
      | beam.Map(self._convert_types)
    )
