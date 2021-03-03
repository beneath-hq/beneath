# Credit goes to https://github.com/ynqa/pandavro

from collections import OrderedDict
from datetime import datetime
from typing import List, Union

import numpy as np
import pandas as pd
from pandas import DatetimeTZDtype

TO_AVRO_TYPE = {
    bool: "boolean",
    int: "long",
    float: "double",
    str: "string",
    bytes: "bytes",
    datetime: {"type": "long", "logicalType": "timestamp-millis"},
    np.bool_: "bool",
    np.int8: "int",
    np.int16: "int",
    np.int32: "int",
    np.uint8: "int",
    np.uint16: "int",
    np.uint32: "int",
    pd.Int8Dtype: "int",
    pd.Int16Dtype: "int",
    pd.Int32Dtype: "int",
    pd.UInt8Dtype: "int",
    pd.UInt16Dtype: "int",
    pd.UInt32Dtype: "int",
    np.int64: "long",
    np.uint64: "long",
    pd.Int64Dtype: "long",
    pd.UInt64Dtype: "long",
    np.unicode_: "string",
    np.float32: "float",
    np.float64: "double",
    np.datetime64: {"type": "long", "logicalType": "timestamp-millis"},
    DatetimeTZDtype: {"type": "long", "logicalType": "timestamp-millis"},
    pd.Timestamp: {"type": "long", "logicalType": "timestamp-millis"},
}


def infer_avro(records: Union[List[dict], pd.DataFrame], key_fields: List[str]):
    if not isinstance(records, pd.DataFrame):
        records = pd.DataFrame.from_records(records)
    fields = _infer_df_fields(records, key_fields, {})
    schema = {"type": "record", "name": "Root", "fields": fields}
    return schema


def _infer_df_fields(df, key_fields, nested_record_names):
    fields = []
    for key, dtype in df.dtypes.items():
        if dtype is np.dtype("O"):
            inferred_type = _infer_dtype_complex(df, key, nested_record_names)
        else:
            inferred_type = _infer_dtype_simple(dtype)
        if inferred_type is None:
            raise TypeError(f"Cannot infer schema for (sub-)type {repr(dtype)}")
        if not _is_array_type(inferred_type):
            if key not in key_fields:
                inferred_type = ["null", inferred_type]
        fields.append({"name": key, "type": inferred_type})
    return fields


def _infer_dtype_simple(dtype):
    if dtype in TO_AVRO_TYPE:
        return TO_AVRO_TYPE[dtype]
    if hasattr(dtype, "type") and dtype.type in TO_AVRO_TYPE:
        return TO_AVRO_TYPE[dtype.type]
    return None


def _infer_dtype_complex(df, field, nested_record_names):
    NoneType = type(None)
    bool_types = {bool, NoneType}
    string_types = {str, NoneType}
    bytes_types = {bytes, NoneType}
    record_types = {dict, OrderedDict, NoneType}
    array_types = {list, NoneType}

    base_field_types = set(df[field].apply(type))

    # String type - checking here makes it the default for all-None columns
    if base_field_types.issubset(string_types):
        return "string"
    # Bool type - double checking here because pandas uses np.dtype('O') for bools with None
    if base_field_types.issubset(bool_types):
        return "boolean"
    # Bytes type
    if base_field_types.issubset(bytes_types):
        return "bytes"
    # Record type
    if base_field_types.issubset(record_types):
        records = df.loc[~df[field].isna(), field].reset_index(drop=True)
        if field in nested_record_names:
            nested_record_names[field] += 1
        else:
            nested_record_names[field] = 0
        record_name = field + "_record" + str(nested_record_names[field])
        record_fields = (
            _infer_df_fields(pd.DataFrame.from_records(records), [], nested_record_names),
        )
        return {
            "type": "record",
            "name": record_name,
            "fields": record_fields,
        }
    # Array type
    if base_field_types.issubset(array_types):
        ds = pd.Series(df.loc[~df[field].isna(), field].sum(), name=field).reset_index(drop=True)
        if ds.empty:
            # Defaults to array of strings
            item_type = "string"
        else:
            item_type = _infer_df_fields(ds.to_frame(), [], nested_record_names)[0]["type"]
        return {"type": "array", "items": item_type}


def _is_array_type(inferred_type):
    return isinstance(inferred_type, dict) and inferred_type["type"] == "array"
