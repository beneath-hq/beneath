import base64
from datetime import datetime
from decimal import *
import apache_beam as beam

"""
  hex: string of format "0x..." -> bytes
  base64: string of base64 -> bytes
  numeric: string containing an integer number (it actually might be an int) -> Decimal
  timestamp: string containing something like "2019-08-06 22:25:53 UTC" -> datetime
"""

# set precision of Decimal to 128 as default of 28 is too small
getcontext().prec = 128

class ConvertFromJSONTypes(beam.PTransform):
  def __init__(self, params):
    self.params = params

  def convert_to_bytes(self, element):
    if not isinstance(element, str):
      return element
    
    if (len(element) > 1) and (element[0:2] == '0x'):
      return bytes.fromhex(element[2:])

    return base64.b64decode(element)

  def convert_to_int(self, element):
    return int(element)

  def convert_to_ts(self, element):
    if not isinstance(element, str):
      return element

    return datetime.strptime(element, '%Y-%m-%d %H:%M:%S.%f %Z')

  def convert_to_numeric(self, element):
    if isinstance(element, str):
      return Decimal(element)
    if isinstance(element, int):
      return Decimal(element)

    return element


  def _convert_types(self, row):
    for field, to_type in self.params.items():
      if row[field] is None:
        continue
      elif to_type == "bytes":
        row[field] = self.convert_to_bytes(row[field])
      elif to_type == "int":
        row[field] = self.convert_to_int(row[field])
      elif to_type == "timestamp":
        row[field] = self.convert_to_ts(row[field])
      elif to_type == "numeric":
        row[field] = self.convert_to_numeric(row[field])

    return row

  def expand(self, pvalue):   
    return (
      pvalue 
      | beam.Map(self._convert_types)
    )
