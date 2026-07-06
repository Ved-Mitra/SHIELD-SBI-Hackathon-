# SHIELD: Secure Heuristics, Intelligence, Education, and Live Detection

### 🛡️ SBI Finnovation Hackathon 2026 Submission
**Team Name:** PhishKillers (IIT Jodhpur)  
**Team Members:** Ved Mitra Verma (Leader), Aditya Sharma, Brijesh Thakkar, Aakarsh Sinha, Mayank Tiwari  

---

## Project Overview
SHIELD is a cloud-native, five-layer, cross-platform security system designed to defend State Bank of India's digital banking ecosystem (YONO) against malicious "look-alike" phishing applications distributed via side-loading and social engineering vectors. 

By unifying hardware-rooted app attestation, real-time machine learning, and zero-cost open-source infrastructure, SHIELD shifts banking security from reactive blacklisting to proactive mathematical certainty.

---

## 5-Layer System Architecture
				┌────────────────────────┐
                │   L0: Threat Origin    │ (SMS Smishing, WhatsApp Clones, Fake Store APKs)
                └───────────┬────────────┘
                            │
                ┌───────────▼────────────┐
                │  L1: Perimeter Defense │ (FastAPI URL Scoring Engine, Unicode Homograph Detector)
                └───────────┬────────────┘
                            │
                ┌───────────▼────────────┐
                │   L2: Device Defense   │ (Kotlin Multiplatform App, WorkManager Passive Scanner)
                └───────────┬────────────┘
                            │
                ┌───────────▼────────────┐
                │ L3: Backend Intel Engine│ (Apache Kafka, Flink Analytics, Airflow Takedown DAG)
                └───────────┬────────────┘
                            │
                ┌───────────▼────────────┐
                │ L4: Ecosystem Command  │ (YONO Guardian Dashboard, CERT-In MISP TAXII, CAMARA APIs)
                └────────────────────────┘
                
---

> **Note:** This repository contains the **Cloud-Native Backend Infrastructure** for SHIELD. The mobile client (KMP) and on-device ML models are hosted in separate branch.


### 1. Three-Gate Login Perimeter (`/backend/three-gate-login`)
* **Tech Stack:** Go (Golang), Redis, RS256 JWT, Envoy Proxy.
* **Responsibilities:**
  * **Gate 1 (Device Authenticity):** Hardware attestation validation mapping Google Play Integrity and Apple App Attest.
  * **Gate 2 (Channel Authenticity):** Strict mutual TLS (mTLS) configurations via an Envoy proxy to eliminate Man-in-the-Middle attacks.
  * **Gate 3 (User Authenticity):** Hardware-bound **FIDO2 WebAuthn** protocol constraints to render credential harvesting and phishing impossible.

### 2. Risk URL API Engine (`/backend/risk-url-engine`)
* **Tech Stack:** Go (Golang), Apache Kafka.
* **Responsibilities:**
  * High-throughput REST API endpoint that receives localized threat intel from mobile clients.
  * Instantly publishes confirmed malicious zero-day phishing links into the real-time event stream.

### 3. Real-Time Threat Intelligence (`/backend/auth-intel-database` & `/thread-intel-database`)
* **Tech Stack:** Go (Golang), PostgreSQL, Apache Kafka, Zookeeper.
* **Responsibilities:**
  * Event-driven consumer microservices that listen to `auth-events` and `url-events` Kafka topics.
  * Maintains an immutable audit log of all brute-force attempts and a deduplicated database of active phishing threats.

### 4. Automated Takedown Pipeline (`/backend/airflow-takedown-DAG`)
* **Tech Stack:** Python, Apache Airflow, PostgreSQL.
* **Responsibilities:**
  * Configured Airflow Directed Acyclic Graphs (DAGs) to bypass human manual handling constraints.
  * Currently simulates takedown task orchestration for Google SafeBrowsing, I4C, and CERT-In. Updates the PostgreSQL database upon completion to drive dashboard state.

### 5. Dashboard (Grafana)
* **Tech Stack:** Grafana, PostgreSQL.
* **Responsibilities:**
  * Containerized SOC (Security Operations Center) dashboard providing real-time visibility into global fraud surges, brute force IP blocks, and phishing takedown status.

---

## 🚀 To Be Done (Future / Enhancements)
While the core zero-trust authentication perimeter is fully functional, the following components are scheduled for future backend implementation:

* **Apache Flink Analytics Engine:** Integrating Flink on top of Kafka to process complex sliding-window event patterns (CEP) and detect coordinated geographic botnets in real-time.
* **PostGIS and Spatial Heatmaps:** Enabling PostGIS extensions in PostgreSQL and building a custom dashboard map to visually render active cyber-attacks across districts in India based on live GPS coordinates.
* **Actual Airflow Takedown API Calls:** Replacing the Python logging stubs in the Airflow DAG with live, authenticated REST `requests` to the Google Play Developer API, CERT-In MISP TAXII servers, and TRAI DLT portals.

---

## Complete Technical Stack

| Domain | Component Stack |
| :--- | :--- |
| **Backend Microservices** | Go (Golang) 1.26, Python |
| **Streaming & Data Processing**| Apache Kafka Event Broker, Zookeeper |
| **Storage & Caching** | Redis 7, PostgreSQL 15 |
| **API Proxy** | Envoy Proxy (mTLS Termination) |
| **Automation & Visualization** | Apache Airflow, Grafana |

---

## Setup & Installation

### Prerequisites
* **Backend Components:** Docker and Docker Compose installed

### Docker Setup
Spin up the backend pipeline, caching servers, and database schemas with a single command:
```bash
docker compose -f infrastructure/docker-compose.yml up -d
docker compose -f infrastructure/DAG/docker-compose.yml up -d
```