---
title: Using the Python library
description:
menu:
  docs:
    parent: writing-to-beneath
    weight: 200
weight: 200
---
One of the primary use cases for the Beneath Python SDK is to create models and write output back to Beneath.

##### Installation
Run on your command line:
```bash
pip install beneath
```

##### Instantiate Beneath client

```python
import beneath
client = beneath.Client(secret="SECRET")
```

##### Connect to data stream
```python
example_stream = client.stream(project_name="example-project", stream_name="example-stream-1")
```

##### Write data to stream
```python
example_stream.write([example_data]) 
```
