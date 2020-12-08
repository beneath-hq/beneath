FROM python:3.7
WORKDIR /app
COPY requirements.txt ./
RUN pip install -r requirements.txt
COPY earthquake_notifications.py ./
CMD ["python", "earthquake_notifications.py", "run", "demos/earthquakes/notify", "--read-quota-mb", "1000", "--write-quota-mb", "0"]
