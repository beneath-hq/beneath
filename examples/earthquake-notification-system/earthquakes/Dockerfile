FROM python:3.7
WORKDIR /app
COPY requirements.txt ./
RUN pip install -r requirements.txt
COPY earthquakes.py ./
CMD ["python", "earthquakes.py", "run", "demos/earthquakes/get-earthquakes", "--read-quota-mb", "1000", "--write-quota-mb", "1000"]
