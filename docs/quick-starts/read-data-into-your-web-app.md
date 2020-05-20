---
title: Read data into your web app
description:
menu:
  docs:
    parent: quick-starts
    weight: 200
weight: 200
---

Time required: 3 minutes.

In this quick-start, we read an event stream from Beneath into a web application. You can read any public data stream or any of the private data streams that you have access to.

##### 1. Log in to the Data Terminal
Go to <a href="https://www.beneath.dev">https://www.beneath.dev</a>, and log in. If you don't yet have an account, create one.

##### 2. Create a Read-Only secret
a) Go to your user profile, by clicking on the profile icon in the top right-hand corner of the screen. <br>
<img src="/media/profile-icon.png" width="90px"/>
b) Click on the Secrets tab <br>
c) Click "Create new read-only secret" and enter a description <br>
d) Save your secret!

##### 3. Go to Project &rarr; Stream &rarr; API tab
a) In the Data Terminal, navigate to your desired project<br>
b) Navigate to your desired stream<br>
c) Click on the API tab

##### 4. Copy-paste the Javascript snippet into your front-end code
```javascript
fetch("https://www.beneath.dev/projects/PROJECT_NAME/STREAM_NAME", {
  "Authorization": "Bearer SECRET",
  "Content-Type": "application/json",
})
.then(res => res.json())
.then(data => {
  // TODO: Add your logic here
  console.log(data)
})
```

