To run locally:

Set up poetry virtual environment

```bash
poetry install
poetry shell
```

Stage the pipeline with a new service, and run

```bash
python clock.py stage SERVICE_PATH --read-quota-mb 1000 --write-quota-mb 10000
python clock.py run SERVICE_PATH
```

Where `SERVICE_PATH` = `USERNAME/PROJECT/YOUR_NEW_SERVICE_NAME`
