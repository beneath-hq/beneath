About `await` and `async` in Python
===================================

The Beneath Python client uses the `await` and `async` keywords in Python extensively. These keywords are the recommended way to use Python's `asyncio <https://docs.python.org/3/library/asyncio.html>`_ library, which is becoming the standard for writing concurrent Python code. 

Since `asyncio` hasn't been completely adopted in the Python community yet, this page answers some common questions for your convenience.

Help, I'm getting a `SyntaxError`
---------------------------------

If you're getting a `SyntaxError`, here's the quick fix: Move the code that uses the `await` keyword into a function defined with the `async` keyword, then wrap your calls to that function with `asyncio.run(...)`.

For example:

.. code-block:: python

  import asyncio
  import beneath

  async def main():
      # TODO: paste your code here

  asyncio.run(main())

Why does Beneath use `asyncio`?
-------------------------------

`asyncio` is a library that is built into Python to make I/O-intensive code significantly faster. That includes code that sends or receives data over the network, such as web requests, server handlers, database calls, and more. 

`asyncio` wraps your code in an *event loop* that tries to run your code concurrently. It notices when Python is waiting for data to pass over the wire, and uses that waiting time to run other awaited code. That means your computer won't sit around and wait for a slow network request to finish when there's other work it could do in the meantime.

It's pretty similar to how `async` and `await` works in JavaScript, if you're familiar with that.

How do I use `asyncio`?
-----------------------

If you just remember these basic rules for `asyncio`, you'll be set for the majority of use cases:

#. Outside a function, use `asyncio.run(...)` to call functions defined with `async`
#. Inside a function:
    #. Use `await` to call functions defined with `async`
    #. Add `async` before `def` to functions where you use `await`
    #. Repeat steps 2.1 and 2.2 until outside a function, then see step 1

Do I always need to use `asyncio.run`?
--------------------------------------

In regular Python, yes, but not in Jupyter or the Python shell.

* In Jupyter, all code runs in an `asyncio` event loop, so you can use `await` directly
* You can use `await` directly in the Python shell if you start it with `python -m asyncio`

In regular Python, you should try to minimize the number of times you call `asyncio.run`. It's a good idea to create an `async` main function (like in the example above) that you call with `asyncio.run`, and then use nested `await` and `async` calls from there on.

Where can I learn more?
-----------------------

- Check out the high-level `asyncio` APIs: https://docs.python.org/3/library/asyncio-task.html
- Browse the `asyncio` documentation: https://docs.python.org/3/library/asyncio.html
- Read the `asyncio` walkthrough on Real Python: https://realpython.com/async-io-python/
