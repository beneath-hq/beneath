FROM python:3.7
WORKDIR /app
COPY requirements.txt ./
RUN pip install -r requirements.txt
COPY fetch_blocks.py blocks-stable.graphql blocks-unstable.graphql ./
CMD ["python", "fetch_blocks.py"]
