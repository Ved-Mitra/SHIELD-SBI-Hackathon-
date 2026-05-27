# Dashboard (prototype)

## Scope
- Show threat intel, takedown status, and system health.
- Provide filters and drill-down on indicators.

## Data sources
- Threat intel DB (read-only queries).
- Airflow takedown status (latest task state).
- Gate-2 auth status (basic metrics only).

## Minimal features
- List view: indicators with status and severity.
- Detail view: evidence and timeline.
- Takedown queue: pending and in-progress.
- Health cards: DAG last run, failures.

## API notes (prototype)
- Backend API should expose read-only endpoints.
- Cache responses lightly to reduce DB load.

## Proposed structure
```
Dashboard/
  README.md
  src/
  public/
```

## References
- Dashboard design patterns: https://www.nngroup.com/articles/dashboard-design/

