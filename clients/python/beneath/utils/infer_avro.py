# Credit goes to https://github.com/ynqa/pandavro

from collections import OrderedDict
from datetime import datetime
import json
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
    np.bool_: "boolean",
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


def infer_avro(records: Union[List[dict], pd.DataFrame]):
    records = _values_to_df(records)
    fields = _infer_df_fields(records, {})
    schema = {"type": "record", "name": "Root", "fields": fields}
    return json.dumps(schema)


def _infer_df_fields(df, nested_record_names):
    fields = []
    for key, dtype in df.dtypes.items():
        if dtype is np.dtype("O"):
            inferred_type = _infer_dtype_complex(df, key, nested_record_names)
        else:
            inferred_type = _infer_dtype_simple(dtype)
        if inferred_type is None:
            raise TypeError(f"Cannot infer schema for (sub-)type {repr(dtype)}")
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
    record_types = {dict, OrderedDict, NoneType}
    array_types = {list, NoneType}

    field_types = set(df[field].apply(type))

    # Error if more than one non-None type
    if len(field_types) > 2 or (len(field_types) == 2 and NoneType not in field_types):
        raise TypeError(
            f"Cannot identify schema type for field '{field}'"
            f" because it has multiple distinct types: '{field_types}'"
        )

    # Default to ["null", "string"] for fields where all values are None
    if len(field_types) == 1 and NoneType in field_types:
        return ["null", "string"]

    # The inferred type
    inferred_type = None

    # Try inferring record
    if field_types.issubset(record_types):
        if field in nested_record_names:
            nested_record_names[field] += 1
        else:
            nested_record_names[field] = 0
        record_name = field + "_record" + str(nested_record_names[field])
        records = df[field].dropna().reset_index(drop=True)
        record_fields = _infer_df_fields(_values_to_df(records), nested_record_names)
        inferred_type = {
            "type": "record",
            "name": record_name,
            "fields": record_fields,
        }
    # Try inferring array
    elif field_types.issubset(array_types):
        if NoneType in field_types:
            raise TypeError(
                f"Cannot create schema for field '{field}' because some values are None and "
                "Beneath doesn't support nullable arrays (hint: replace None values with empty "
                "lists)"
            )
        ds = pd.Series(df.loc[~df[field].isna(), field].sum(), name=field).reset_index(drop=True)
        if ds.empty:
            # Defaults to array of strings
            item_type = "string"
        else:
            ds = _values_to_df(ds.to_frame())
            item_type = _infer_df_fields(ds, nested_record_names)[0]["type"]
        return {"type": "array", "items": item_type}
    # Infer by looking up in TO_AVRO_TYPE
    else:
        for field_type in field_types:
            if field_type is not NoneType:
                inferred_type = TO_AVRO_TYPE.get(field_type)

    # Error if no match for inferred_type
    if inferred_type is None:
        raise TypeError(
            f"Cannot identify schema type for field '{field}'"
            f" with Python types: '{field_types}'"
        )

    # Make nullable if applicable
    if NoneType in field_types:
        inferred_type = ["null", inferred_type]

    return inferred_type


def _values_to_df(values):
    if isinstance(values, pd.DataFrame):
        df = values
    else:
        df = pd.DataFrame.from_records(values)
    # df = df.where(pd.notnull(df), None)
    df = df.replace({np.nan: None})
    return df
