from airflow import DAG
from airflow.operators.python import PythonOperator
from airflow.operators.empty import EmptyOperator
from datetime import datetime
import logging
import yaml
import requests

def load_policies():
    with open('/opt/airflow/config/policies.yaml', 'r') as f:
        return yaml.safe_load(f)

def rank_and_filter(**context):
    dag_run = context['dag_run']
    url = dag_run.conf.get('malicious_url', 'UNKNOWN_URL')
    logging.info(f"Ranking URL: {url}")
    
    policies = load_policies()
    
    # Since the mobile device runs the ONNX ML model locally, 
    # any URL that reaches this DAG has ALREADY been flagged as highly malicious by the client.
    # Therefore, we automatically approve the takedown (IMPROVEMENT: can make an even higher parameter ML model that would recheck the URL before takedown just as verification)
    logging.info("URL was flagged by on-device ML model. Bypassing manual SOC review and triggering auto-takedown.")
    return True

def report_to_google(**context):
    url = context['dag_run'].conf.get('malicious_url', 'UNKNOWN_URL')
    logging.info(f"Executing Google Play Developer API webhook for: {url} ...")

def report_to_i4c(**context):
    url = context['dag_run'].conf.get('malicious_url', 'UNKNOWN_URL')
    logging.info(f"Submitting Cybercrime platform report for: {url} ...")

def report_to_cert_in(**context):
    url = context['dag_run'].conf.get('malicious_url', 'UNKNOWN_URL')
    logging.info(f"Submitting CERT-In MISP TAXII incident for: {url} ...")

def notify_takedown_started(**context):
    url = context['dag_run'].conf.get('malicious_url', 'UNKNOWN_URL')
    logging.info(f"Dashboard Update: {url} is now IN_PROGRESS (Takedown Initiated)")

def notify_takedown_completed(**context):
    url = context['dag_run'].conf.get('malicious_url', 'UNKNOWN_URL')
    logging.info(f"Dashboard Update: {url} is now DONE (Takedown Requests Successfully Sent)")

default_args = {
    'owner': 'Ved',
    'start_date': datetime(2026, 1, 1),
    'retries': 1,
}

with DAG(
    'takedown_pipeline',
    default_args=default_args,
    schedule=None, # This DAG is exclusively triggered by the external Kafka consumer via REST API
    catchup=False
) as dag:
    
    ingest = EmptyOperator(task_id='ingest_signals')
    
    rank = PythonOperator(
        task_id='rank_and_filter',
        python_callable=rank_and_filter
    )
    
    prepare = EmptyOperator(task_id='prepare_takedown_packet')
    
    # Parallel worker tasks for takedown reporting (as per README architecture)
    google = PythonOperator(task_id='report_to_google', python_callable=report_to_google)
    i4c = PythonOperator(task_id='report_to_i4c', python_callable=report_to_i4c)
    cert_in = PythonOperator(task_id='report_to_cert_in', python_callable=report_to_cert_in)
    
    update = EmptyOperator(task_id='update_status')
    
    notify_started = PythonOperator(
        task_id='notify_takedown_started',
        python_callable=notify_takedown_started
    )
    
    notify_completed = PythonOperator(
        task_id='notify_takedown_completed',
        python_callable=notify_takedown_completed
    )
    
    # Define DAG execution flow
    ingest >> rank >> notify_started >> prepare
    
    # Fan-out to parallel reporting workers, then Fan-in to update status
    prepare >> [google, i4c, cert_in] >> update
    
    update >> notify_completed
