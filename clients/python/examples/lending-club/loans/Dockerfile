FROM python:3.7
WORKDIR /app
COPY requirements.txt ./
RUN pip install -r requirements.txt
COPY fetch_new_loans.py loans.graphql ./
CMD ["python", "fetch_new_loans.py"]
