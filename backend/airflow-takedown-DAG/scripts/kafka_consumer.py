import os
import json
import requests
from confluent_kafka import Consumer

# Airflow REST API setup
AIRFLOW_URL = os.getenv("AIRFLOW_URL", "http://localhost:8080/api/v1/dags/takedown_pipeline/dagRuns")
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
    
    while True:
        msg = c.poll(1.0)
        if msg is None:
            continue
        if msg.error():
            print(f"Consumer error: {msg.error()}")
            continue
            
        payload = json.loads(msg.value().decode('utf-8'))
        print(f"Received phishing event: {payload}")
        
        # Trigger Airflow DAG
        trigger_payload = {
            "conf": {
                "malicious_url": payload.get("url"),
                "device_id": payload.get("device_id")
            }
        }
        
        resp = requests.post(
            AIRFLOW_URL,
            json=trigger_payload,
            auth=(AIRFLOW_USER, AIRFLOW_PASS)
        )
        print(f"Triggered DAG: {resp.status_code} - {resp.text}")

if __name__ == '__main__':
    start_consumer()
