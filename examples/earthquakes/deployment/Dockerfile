FROM python:3.8
RUN pip install poetry==1.1.4
WORKDIR /app
COPY poetry.lock pyproject.toml /app/
RUN poetry config virtualenvs.create false \
  && poetry install --no-dev --no-interaction --no-ansi
COPY . .
CMD ["python", "main.py", "run", "examples/earthquakes/generate-earthquakes", "--read-quota-mb", "1000", "--write-quota-mb", "1000"]
