# airflow-takedown-DAG (prototype)

## Scope
- Automate takedown workflows for threat signals.
- Accept inputs from threat-intel database and policy rules.
- Produce audit logs and status updates for the dashboard.

## Inputs
- Indicators: domains, URLs, IPs, handles.
- Confidence score and severity.
- Evidence bundle (links, screenshots, hashes).

## Outputs
- Takedown request records.
- Status updates per target.
- Audit trail for review.

## Minimal DAG outline
- `ingest_signals`
- `rank_and_filter`
- `prepare_takedown_packet`
- `send_takedown_request`
- `update_status`
- `notify_dashboard`

## Ops notes (prototype)
- Keep DAGs and configs in-repo.
- Use local Airflow for development.
- Store secrets in env vars only.

## Proposed structure
```
airflow-takedown-DAG/
  README.md
  dags/
    takedown_pipeline.py
  config/
    policies.yaml
  scripts/
    seed_sample_signals.py
```

## References
- Airflow DAGs: https://airflow.apache.org/docs/apache-airflow/stable/core-concepts/dags.html

