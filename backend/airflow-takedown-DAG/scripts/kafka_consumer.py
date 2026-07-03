import os
import json
import requests
from confluent_kafka import Consumer

# Airflow REST API setup
AIRFLOW_URL  = os.getenv("AIRFLOW_URL",  "http://airflow-webserver:8080/api/v1/dags/takedown_pipeline/dagRuns")
AIRFLOW_USER = os.getenv("AIRFLOW_USER", "airflow")
AIRFLOW_PASS = os.getenv("AIRFLOW_PASS", "airflow")
KAFKA_BROKER = os.getenv('KAFKA_BROKER_URL', 'localhost:9092')

def start_consumer():
    c = Consumer({
        'bootstrap.servers': KAFKA_BROKER,
        'group.id': 'airflow-trigger-group',
        'auto.offset.reset': 'latest'
    })

    c.subscribe(['url-events'])
    print(f"Listening for phishing events on Kafka ({KAFKA_BROKER})...")

    # C-10 fix: wrap in try/finally so consumer is always closed on exit
    try:
        while True:
            msg = c.poll(1.0)
            if msg is None:
                continue
            if msg.error():
                print(f"Consumer error: {msg.error()}")
                continue

            # C-9 fix: handle malformed/non-JSON messages gracefully
            try:
                payload = json.loads(msg.value().decode('utf-8'))
            except (json.JSONDecodeError, UnicodeDecodeError) as e:
                print(f"Bad message, skipping: {e}")
                continue

            print(f"Received phishing event: {payload}")

            trigger_payload = {
                "conf": {
                    "malicious_url": payload.get("url"),
                    "device_id": payload.get("device_id")
                }
            }

            # C-10 fix: add timeout and raise_for_status so failures are visible
            try:
                resp = requests.post(
                    AIRFLOW_URL,
                    json=trigger_payload,
                    auth=(AIRFLOW_USER, AIRFLOW_PASS),
                    timeout=10
                )
                resp.raise_for_status()
                print(f"Triggered DAG: {resp.status_code}")
            except requests.exceptions.Timeout:
                print("DAG trigger timed out — Airflow may be slow. Will retry on next event.")
            except requests.exceptions.HTTPError as e:
                print(f"DAG trigger failed with HTTP error: {e.response.status_code} - {e.response.text}")
            except requests.exceptions.RequestException as e:
                print(f"DAG trigger failed: {e}")
    finally:
        c.close()
        print("Kafka consumer closed.")

if __name__ == '__main__':
    start_consumer()

