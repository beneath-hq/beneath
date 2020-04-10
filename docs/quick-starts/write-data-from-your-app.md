---
title: Write data from your app
description:
menu:
  docs:
    parent: quick-starts
    weight: 300
weight: 300
---

Time required: 5 minutes.

In this quick-start, we write an event stream to Beneath. The event stream might originate from your application or from some external API that you would like to consume.

##### 1. Install the Beneath Python SDK
If you haven't already, install the Python SDK from your command line:
```bash
pip install beneath
```

##### 2. Create your schema
In a blank text file, create the schema for your data stream. Save the file with a ".graphql" extension. Here is a template to follow:

```graphql
"The stream's description goes here. This is a stream schema that can be used as a template."
type SchemaTemplate
  @stream(name: "stream-template-1")
  @key(fields: ["my_integer_index"])
{
  "Example integer field. This is also the stream's index, as the key defines above."
  my_integer_index: Int!

  "Example timestamp field"
  my_timestamp: Timestamp!

  "Example bytes field"
  my_bytes: Bytes32!

  "Example really big number field"
  my_big_number: Numeric!

  "Example string field. This field is not required, as denoted by the lack of exclamation point."
  my_optional_string: String
}
```
Every stream requires a schema. If needed, there is more information about the Beneath schema language [here](/docs/schema-language).

##### 3. Create your stream
Use the CLI to create your stream:

```bash
beneath root-stream create -f my_schema.graphql -p ExampleProject --manual false --batch false 
```
In this case, as the stream's data will be written programatically and continuously from your application, we have specified that the stream is neither "manual" nor "batch."

##### 4. Initialize the Python client and connect to your stream
```python
client = beneath.Client(secret="SECRET")
example_stream = client.stream(project_name="example-project", stream_name="example-stream-1")
```

##### 5. Generate your data
However your application generates the data of interest, capture the data and ensure it aligns with your stream's schema. If the data does not conform to the Beneath stream's schema, Beneath will reject the write request.

```python
example_data = {
  "my_integer_index": 1,
  "my_timestamp": datetime.utcfromtimestamp(SOMETIMESTAMP),
  "my_bytes": bytes(SOMEBYTES),
  "my_big_number": Decimal(99999999999999),
  "my_optional_string": "not required but I'm providing one anyways"
}
```

##### 6. Write your data to Beneath
The write function accepts a list of data records. Here, we write a list of size 1.

```python
example_stream.write([example_data])
```

##### 7. Check out the Data Terminal to see data arrive in realtime 
Navigate to your stream in the Data Terminal and go to the Explore tab to see your data arrive in realtime. The url will look like: https://www.beneath.dev/project/PROJECTNAME/stream/STREAMNAME/explore

##### Further examples
More examples of the Python SDK in action can be found <a href="https://gitlab.com/_beneath/beneath-core/-/tree/master/clients%2Fpython%2Fexamples" class="link">here</a>.
