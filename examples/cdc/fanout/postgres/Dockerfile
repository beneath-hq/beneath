FROM python:3.8
RUN pip install poetry==1.1.12
WORKDIR /app
COPY poetry.lock pyproject.toml .
RUN poetry config virtualenvs.create false \
  && poetry install --no-dev --no-interaction --no-ansi
COPY . .
ENTRYPOINT ["python", "main.py"]
