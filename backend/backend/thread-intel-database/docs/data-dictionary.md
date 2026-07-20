# Threat-Intel Database Data Dictionary

## Table: `threat_intel`

This table serves as the central repository for all identified threats against the SBI network (phishing domains, malicious IPs, fake social media handles).

| Column Name       | Data Type               | Constraints                  | Description |
|-------------------|-------------------------|------------------------------|-------------|
| `id`              | `UUID`                  | `PRIMARY KEY`                | Unique identifier for the intelligence record. |
| `indicator_type`  | `indicator_type_enum`   | `NOT NULL`                   | The type of threat. Enum: `domain`, `url`, `ip`, `handle`, `hash`. |
| `indicator_value` | `TEXT`                  | `NOT NULL`                   | The actual malicious value (e.g., `extremely-malicious-sbi-clone.in`). |
| `source`          | `source_enum`           | `NOT NULL`, Default: `manual`| The origin of the report. Enum: `crawler`, `manual`, `partner`, `device_ml`. |
| `confidence`      | `INTEGER`               | `0-100`, Default: `50`       | Confidence score that this is a true positive. |
| `severity`        | `severity_enum`         | `NOT NULL`, Default: `medium`| Threat severity. Enum: `low`, `medium`, `high`, `critical`. |
| `evidence`        | `JSONB`                 | None                         | Flexible unstructured JSON data storing screenshots, raw HTML, network logs, or device telemetry. |
| `status`          | `status_enum`           | `NOT NULL`, Default: `new`   | Current lifecycle state. Enum: `new`, `in_review`, `takedown_requested`, `resolved`, `false_positive`. |
| `first_seen`      | `TIMESTAMP WITH TZ`     | Default: `CURRENT_TIMESTAMP` | When this threat was first identified by the system. |
| `last_seen`       | `TIMESTAMP WITH TZ`     | Default: `CURRENT_TIMESTAMP` | When this threat was most recently detected active. |
| `created_at`      | `TIMESTAMP WITH TZ`     | Default: `CURRENT_TIMESTAMP` | Database record creation time. |
| `updated_at`      | `TIMESTAMP WITH TZ`     | Auto-updates via Trigger     | Database record last modification time. |

### Enums Details
* **Source**: `device_ml` specifically represents reports generated automatically by the Android on-device Risk-URL-Engine (using the ONNX model).
* **Status**: `takedown_requested` indicates the Airflow Takedown DAG has successfully initiated external requests to Google Safe Browsing / CERT-In / I4C.
