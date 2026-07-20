# Airflow Takedown DAG

## Overview
This module automates the regulatory reporting and domain takedown process for SHIELD. It listens to Kafka for new phishing threats and triggers workflows to report them.

## Workflow
1. A Kafka consumer (`scripts/kafka_consumer.py`) listens to the `url-events` topic.
2. When a phishing URL is received, it triggers the `takedown_pipeline` DAG in Airflow via the Airflow REST API.
3. The DAG evaluates the threat and simulates sending automated takedown requests to authorities like Google SafeBrowsing, CERT-In, and I4C.

## Current Implementation Details
- **Kafka Consumer**: Written in Python using `confluent-kafka`. Hardened with robust error handling, JSON decoding protection, HTTP timeouts, and graceful shutdown on SIGTERM.
- **Airflow DAG**: `dags/takedown_pipeline.py`. Expects configuration payload containing URL details.

## Running Locally
This module requires an Airflow environment and the Kafka broker to be running.
The Kafka consumer requires:
- `KAFKA_BROKER_URL` (default: `localhost:9092`)
- `KAFKA_TOPIC` (default: `url-events`)
- `AIRFLOW_URL`
- `AIRFLOW_USER`
- `AIRFLOW_PASSWORD`
