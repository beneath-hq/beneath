To run locally:

Set up poetry virtual environment

```bash
poetry install
poetry shell
```

Stage the tables and a service for the pipeline, then run it

```bash
python clock.py stage USERNAME/PROJECT/clock --read-quota-mb 1000 --write-quota-mb 2000
python clock.py run USERNAME/PROJECT/clock
```
