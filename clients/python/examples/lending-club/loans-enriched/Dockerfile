FROM python:3.7
WORKDIR /app
COPY requirements.txt ./
RUN pip install -r requirements.txt
COPY enrich_loans.py loans_enriched.graphql model.pkl ./
CMD ["python", "enrich_loans.py", "run", "epg/lending-club/enrich-loans", "--strategy", "delta", "--read-quota-mb", "1000", "--write-quota-mb", "1000"]
