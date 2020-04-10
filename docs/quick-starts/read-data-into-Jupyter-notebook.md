---
title: Read data into Jupyter notebook
description:
menu:
  docs:
    parent: quick-starts
    weight: 100
weight: 100
---

Time required: 5 minutes.

In this quick-start, we read an event stream from Beneath into a Jupyter notebook. You can read any public data stream or any of the private data streams that you have access to.

##### 1. Install the Beneath Python SDK
Install the Python SDK from your command line:
```bash
pip install beneath
```

##### 2. Log in to the Data Terminal
Go to [https://www.beneath.dev](https://www.beneath.dev), and log in. If you don't yet have an account, create one.

##### 3. Create a Read-Only secret
a) Navigate to your user profile by clicking on the profile icon in the top right-hand corner of the screen. <br>
<img src="/media/profile-icon.png" width="70px"/>
b) Click on the Secrets tab <br>
c) Click "Create new read-only secret" and enter a description <br>
d) Save your secret!

##### 4. Authorize your local environment
From your command line,
```bash
beneath auth SECRET
``` 
Now your secret has been placed in a hidden file on your Desktop called ".beneath"

##### 5. Go to desired Project &rarr; Stream &rarr; API tab
a) In the Data Terminal, navigate to your desired project<br>
b) Navigate to your desired stream<br>
c) Click on the API tab

##### 6. Copy-paste the Python snippet into a Jupyter notebook
Many data workers choose [Jupyter notebooks](https://jupyter.org/) for ad-hoc analyses. A few short lines of Python code are all you need to import Beneath data into your notebook environment.

```python
from beneath import Client
client = Client()
stream = client.stream(project_name="PROJECT_NAME", stream_name="STREAM_NAME")
df = stream.read()
```

##### 7. Enjoy the easy-to-use API
TODO: read-from-cursor

