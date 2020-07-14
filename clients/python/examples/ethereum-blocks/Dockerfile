FROM python:3.7
WORKDIR /app
COPY requirements.txt ./
RUN pip install -r requirements.txt
COPY fetch_blocks.py ./
CMD ["python", "fetch_blocks.py", "run"]
