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

In this quick-start, we write a data stream to Beneath. The stream might originate from your application or from some external API that you would like to consume.

##### 1. Install the Beneath Python SDK
If you haven't already, install the Python SDK from your command line:
```bash
pip install beneath
```

##### 2. Log in to the Data Terminal
Go to [https://www.beneath.dev](https://www.beneath.dev), and log in. If you don't yet have an account, create one.

##### 3. Create a Command-Line secret
a) Navigate to your user profile by clicking on the profile icon in the top right-hand corner of the screen. <br>
<img src="/img/profile-icon.png" width="70px"/>
b) Click on the Secrets tab <br>
c) Click "Create new command-line secret" and enter a description <br>
d) Save your secret!

##### 4. Authorize your local environment
From your command line,
```bash
beneath auth SECRET
``` 
Now your secret has been placed in a hidden file on your Desktop called ".beneath"

##### 5. Create a Beneath project
On Beneath, every data stream lives in a project. Just like, on GitHub, every code file lives in a repository.

Additionally, every project is assigned to an organization. This ensures that all project resources have someone assigned for billing purposes.

**When you sign up for Beneath, you are automatically assigned an organization with the same name as your username.** The organization is automatically assigned to a Free billing plan.

On your command line:
```bash
beneath project create PROJECT_NAME -o ORGANIZATION_NAME 
```

This tutorial won't cover adding a pretty display name for your project, a description for your project, a web address you'd like your project to be affiliated with, and a photo url for your project. But that's all possible.

##### 6. Create your schema

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
beneath root-stream create -f template.graphql -p ExampleProject --manual false --batch false 
```
Great! You just created a Beneath stream. The next step is to write data to the stream.

In this case, by specifying the stream is neither "manual" nor "batch", we've denoted that the stream's data will be written programatically and continuously from your application.

##### 8. Initialize the Python client and connect to your stream
Either in your application code, or, to test, in a Jupyter notebook, connect to your stream with the following Python code. Use your command-line secret.

```python
import beneath
client = beneath.Client(secret="SECRET")
stream = await client.find_stream("USERNAME/example-project/stream-template-1")
```

##### 9. Generate your data
Generate or capture the data of interest and ensure it aligns with your stream's schema that you defined above. If the data does not conform to the defined schema, Beneath will reject the write request.

```python
from datetime import datetime
import secrets
import sys

def generate_record(n):
  return {
    "my_integer_index": n,
    "my_timestamp": datetime.now(),
    "my_bytes": secrets.token_bytes(4),
    "my_big_number": secrets.randbelow(sys.maxsize),
    "my_optional_string": None
  }

n = 25
records = [generate_record(n) for _ in range(n)]
```

##### 10. Write your data to Beneath
The write function accepts a list of data records. Here, we write a list of size 1. We use the await syntax so that we do not block our application while we wait for Beneath's response.

```python
await stream.write(records)
```

##### 11. Check out the Data Terminal to see data arrive in realtime 
Navigate to your stream in the Data Terminal and go to the Explore tab to see your data arrive in realtime. The url will look like: https://www.beneath.dev/USERNAME/PROJECTNAME/STREAMNAME

##### Further examples
More examples of the Python SDK in action can be found <a href="https://gitlab.com/_beneath/beneath-core/-/tree/master/clients%2Fpython%2Fexamples" class="link">here</a>.
