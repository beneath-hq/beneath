---
title: Using the REST APIs
description:
menu:
  docs:
    parent: reading-from-beneath
    weight: 190
weight: 190
---
Every stream is available via a REST API. Using your language of choice, you can access the data and import it into your environment. We provide examples below for the Command Line, Python, and Javascript. 

**Command line**
```bash
curl -H "Authorization: SECRET" https://beneath.dev/projects/PROJECTNAME/streams/STREAMNAME
```

**Python**
```python
import requests

r = requests.get('https://beneath.dev/projects/PROJECTNAME/streams/STREAMNAME', headers={"authorization":"Bearer SECRET"})
r.json()
```


**Javascript**
```javascript
fetch("https://beneath.dev/projects/PROJECTNAME/streams/STREAMNAME", {
  "Authorization": "Bearer SECRET",
  "Content-Type": "application/json",
})
.then(res => res.json())
.then(data => {
  // TODO: Add your logic here
  console.log(data)
})
```
