---
title: Stateful functions
description:
menu:
  docs:
    parent: transforming-data
    weight: 200
weight: 200
---

Applying stateful functions to your Beneath data streams are possible using Apache Beam. Fully integrating Beam into the Beneath platform is on the near-term roadmap.

Please reach out if you are interested in this feature.

```python
from beneath.client import Client
from beneath.beam import ReadFromBeneath, WriteToBeneath
def run(argv=None):
    client = Client(secret="SECRET")
    stream = client.stream("PROJECTNAME", "STREAMNAME")
    pipeline = beam.Pipeline()
    (
       pipeline
       | 'Read'         >> ReadFromBeneath(stream)
       | 'UDF' 		>> YourModel()
       | 'Write'        >> WriteToBeneath(stream)
     )

if __name__ == '__main__':
 run()
```