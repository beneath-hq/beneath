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

## Log in to the Data Terminal
Go to the [Terminal](https://beneath.dev/?noredirect=1), and log in. If you don't yet have an account, create one.

## Create a Read-Only secret

- Go to your user profile, by clicking on the profile icon in the top right-hand corner of the screen.
- Click on the Secrets tab
- Click "Create new read-only secret" and enter a description
- Save your secret!

## Navigate to a data stream's API tab

- The Beneath directory structure is USER/PROJECT/STREAM
- In the [Terminal](https://beneath.dev/?noredirect=1), navigate to your desired stream, and click on the API tab

## Copy-paste the Javascript snippet into your front-end code
```javascript
fetch("https://data.beneath.dev/v1/USER/PROJECT/STREAM", {
  "Authorization": "Bearer SECRET",
  "Content-Type": "application/json",
})
.then(res => res.json())
.then(data => {
  // TODO: Add your logic here
  console.log(data)
})
```
