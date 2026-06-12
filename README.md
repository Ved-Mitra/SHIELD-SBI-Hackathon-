# SHIELD: Secure Heuristics, Intelligence, Education, and Live Detection

### 🛡️ SBI Finnovation Hackathon 2026 Submission
**Team Name:** PhishKillers (IIT Jodhpur)  
**Team Members:** Ved Mitra Verma (Leader), Aditya Sharma, Brijesh Thakkar, Aakarsh Sinha, Mayank Tiwari  

---

## 📌 Project Overview
SHIELD is a cloud-native, five-layer, cross-platform security system designed to defend State Bank of India's digital banking ecosystem (YONO) against malicious "look-alike" phishing applications distributed via side-loading and social engineering vectors. 

By unifying hardware-rooted app attestation, real-time machine learning, and zero-cost open-source infrastructure, SHIELD shifts banking security from reactive blacklisting to proactive mathematical certainty.

---

## 🏗️ 5-Layer System Architecture
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

## 👥 Module Ownership & Work Division

### 1. 🧠 URL Risk Engine (`/core-ml`)
* **Owner:** Aditya Sharma
* **Tech Stack:** Python, Hugging Face (BERT), ONNX Runtime, C++ Core Bindings.
* **Responsibilities:**
  * Quantized a transformer-based BERT classifier into **ONNX INT8 format** to process and score pasted URLs directly on local mobile clients under `<100ms`.
  * Engineered a custom, low-latency algorithm to isolate and flag specialized **Devanagari-Latin Unicode homograph** lookalike attacks (e.g., `yonoŝbi.in`).
  * Integrated **SHAP reason codes** to output human-readable threat indicators for SOC analysts.

### 2. 🔐 Three-Gate Login & Client Foundation (`/mobile-kmp`)
* **Owner:** Ved Mitra Verma
* **Tech Stack:** Kotlin Multiplatform (KMP), Compose Multiplatform, Ktor Client, Swift (Fallback).
* **Responsibilities:**
  * Built the unified mobile client codebase targeting dual-platform native runtimes from a single framework utilizing **Codemagic CI**.
  * Structured the cryptographic sequential **Three-Gate authentication pipeline**:
    * **Gate 1 (App Authenticity):** Hardware attestation abstractions mapping Google Play Integrity and Apple App Attest.
    * **Gate 2 (Channel Authenticity):** Strict mutual TLS (mTLS) configurations and SHA-256 certificate pinning via OkHttp/Darwin network engines.
    * **Gate 3 (User Authenticity):** Implemented an un-phishable cryptographic layer using hardware-bound **FIDO2 WebAuthn** protocol constraints to render credential harvesting impossible.

### 3. 🔍 Passive APK Live Scanner (`/android-watchdog`)
* **Owner:** Brijesh Thakkar
* **Tech Stack:** Kotlin, Android SDK, Android Jetpack WorkManager, freeRASP SDK.
* **Responsibilities:**
  * Implemented a background-persistent **WorkManager service** running within the native `androidMain` multiplatform lifecycle.
  * Utilized `PackageManager.GET_SIGNING_CERTIFICATES` to continuously enumerate installed platform apps and calculate **Levenshtein-distance fuzzy string rules** against unauthorized bank brand variations.
  * Set up immediate in-app warning alerts on detecting certificate signature fingerprint mismatches.
  * Embedded **freeRASP** for runtime debugging blocks, emulator termination, and hook framework detection (Frida/Xposed).

### 4. 🔀 Airflow Takedown DAG (`/automation-pipeline`)
* **Owner:** Aakarsh Sinha
* **Tech Stack:** Python, Apache Airflow, REST APIs.
* **Responsibilities:**
  * Programmed high-speed automation workflows using **Apache Airflow Directed Acyclic Graphs (DAGs)** to bypass human manual handling constraints.
  * Configured parallel worker tasks that trigger webhook reporting calls to the Google Play Console Developer API, dynamic hosting provider DMCA takedown schemes, **I4C Cybercrime platform**, and **CERT-In**.
  * Optimized execution profiles to route full system incidents safely to endpoints in roughly **8 seconds**, driving target infrastructure takedown SLA times down to `<4 hours`.

### 5. 🗣️ Bhashini Integration & YONO Guardian Dashboard (`/analytics-dashboard`)
* **Owner:** Mayank Tiwari
* **Tech Stack:** Python (FastAPI), Grafana, Leaflet.js, PostgreSQL/PostGIS, Apache Kafka & Flink.
* **Responsibilities:**
  * Built the localization layer integrating the **Govt. of India’s Bhashini AI Platform**, configuring real-time on-the-fly Neural Machine Translation (NMT) and Text-to-Speech (TTS) safety audio playback in priority regional Indian languages.
  * Deployed the containerized **YONO Guardian Dashboard** via Grafana to present centralized visibility parameters.
  * Integrated spatial heatmaps with **Leaflet.js** and PostGIS to render mock district-level fraud surge alerts generated asynchronously by stream processing engines.

---

## 🛠️ Complete Technical Stack

| Domain | Component Stack |
| :--- | :--- |
| **Mobile Frontend** | Kotlin Multiplatform (KMP), Compose Multiplatform, Ktor Core, freeRASP SDK, Codemagic CI |
| **Backend Framework** | Python (FastAPI), Go (Golang), Apache Airflow DAG Engines |
| **Streaming & Data Processing**| Apache Kafka Event Broker, Apache Flink Streaming Engine |
| **Storage & Caching** | Redis Caching, PostgreSQL with PostGIS extensions, Elasticsearch |
| **Threat Intelligence** | MISP local instance, STIX 2.1 JSON Schema, TAXII 2.1 Protocols |
| **Sovereign Interfacing** | Digital India Bhashini AI API, TRAI DLT Webhook Portals, GSMA CAMARA Gateway |

---

## ⚡ Setup & Installation

### Prerequisites
* **Android Development:** Android Studio Hedgehog+, JDK 17
* **iOS Compilation:** macOS environment with Xcode 16+ and iOS 18.0+ deployment target.
* **Backend Ingestion Components:** Docker and Docker Compose installed

### Docker Ingestion Setup
Spin up the backend telemetry pipeline, caching servers, and database schemas with a single command:
```bash
docker-compose -f infrastructure/docker-compose.yml up -d
