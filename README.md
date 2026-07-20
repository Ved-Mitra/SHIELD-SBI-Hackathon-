# SHIELD: Secure Heuristics, Intelligence, Education, and Live Detection

### 🛡️ SBI Finnovation Hackathon 2026 Submission
**Team Name:** PhishKillers (IIT Jodhpur)  
**Team Members:** Ved Mitra Verma, Aditya Sharma

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
                │  L1: Perimeter Defense │ (ONNX MobileBERT URL Engine, Unicode Homograph Detector)
                └───────────┬────────────┘
                            │
                ┌───────────▼────────────┐
                │   L2: Device Defense   │ (Kotlin Multiplatform App, WorkManager Passive Scanner)
                └───────────┬────────────┘
                            │
                ┌───────────▼────────────┐
                │ L3: Backend Intel Engine│ (Apache Kafka, Go Microservices, Airflow Takedown DAG)
                └───────────┬────────────┘
                            │
                ┌───────────▼────────────┐
                │ L4: Ecosystem Command  │ (Grafana Dashboard, Threat-Intel Database)
                └────────────────────────┘
                
---

## Mobile App (Client Defense)

### 1. URL Risk Engine (`/core-ml`)
* **Tech Stack:** Kotlin, Hugging Face (MobileBERT), ONNX Runtime, Ktor Client.
* **Responsibilities:**
  * Executed **ONNX INT8 format** MobileBERT inference to process and score pasted URLs directly on local mobile clients in <30ms.
  * Engineered a custom algorithm to isolate and flag **Devanagari-Latin Unicode homograph** attacks.
  * Implemented telemetry reporting via **Ktor (`UrlReportClient`)** utilizing strict `withTimeout(5000L)` bounds.

### 2. Three-Gate Login & Client Foundation (`/mobile-kmp`)
* **Tech Stack:** Kotlin Multiplatform (KMP), Compose Multiplatform, Coroutines.
* **Responsibilities:**
  * Built the unified mobile client codebase targeting dual-platform native runtimes from a single framework.
  * Structured the UI state for the **Three-Gate authentication pipeline** (Device Attestation, mTLS, FIDO2).

### 3. Passive APK Live Scanner (`/android-watchdog`)
* **Tech Stack:** Kotlin, Android SDK, Android Jetpack WorkManager, SharedPreferences.
* **Responsibilities:**
  * Implemented a background-persistent **WorkManager service (`ScannerWorker`)** evaluating installed apps against unauthorized bank brand variations.
  * Engineered **SharedPreferences persistence** for the scanner's toggle state and language selection.

### 4. Multi-Lingual Threat Alarm System (`/audio-alerts`)
* **Tech Stack:** Kotlin, Android MediaPlayer, Local MP3 Assets.
* **Responsibilities:**
  * Engineered a zero-latency **Multi-Lingual Audio Alarm System**, playing back pre-recorded regional language warnings for visually impaired users.

---

## Backend Infrastructure

### 1. Three-Gate Login Perimeter (`backend/backend/three-gate-login`)
* **Tech Stack:** Go (Golang), Redis, RS256 JWT, Envoy Proxy.
* **Responsibilities:**
  * **Gate 1 (Device Authenticity):** Hardware attestation validation mapping Google Play Integrity.
  * **Gate 2 (Channel Authenticity):** Strict mutual TLS (mTLS) configurations via an Envoy proxy.
  * **Gate 3 (User Authenticity):** Hardware-bound **FIDO2 WebAuthn** protocol constraints.

### 2. Risk URL API Engine & Real-Time Threat Intelligence
* **Tech Stack:** Go (Golang), PostgreSQL, Apache Kafka.
* **Responsibilities:**
  * High-throughput REST API endpoint that receives localized threat intel from mobile clients and instantly publishes to Kafka topics (`url-events`).
  * Maintains an immutable audit log of all brute-force attempts and a deduplicated database of active phishing threats.

### 3. Automated Takedown Pipeline (`backend/backend/airflow-takedown-DAG`)
* **Tech Stack:** Python, Apache Airflow, PostgreSQL.
* **Responsibilities:**
  * Airflow Directed Acyclic Graphs (DAGs) simulate takedown task orchestration for Google SafeBrowsing, I4C, and CERT-In.

### 4. Dashboard (Grafana)
* **Tech Stack:** Grafana, PostgreSQL.
* **Responsibilities:**
  * Containerized SOC (Security Operations Center) dashboard providing real-time visibility into global fraud surges and takedown status.

---

## Complete Technical Stack

| Domain | Component Stack |
| :--- | :--- |
| **Mobile Frontend** | Kotlin Multiplatform (KMP), Compose Multiplatform, Coroutines |
| **On-Device ML** | Hugging Face MobileBERT, ONNX Runtime (C++ Bindings) |
| **Backend Microservices** | Go (Golang) 1.26, Python, Ktor |
| **Streaming & Data**| Apache Kafka Event Broker, PostgreSQL 15, Redis 7 |
| **Automation & Visualization** | Apache Airflow, Grafana, Prometheus, Loki |

---

## Setup & Installation

### Prerequisites
* **Android Development:** Android Studio Hedgehog+, JDK 17
* **iOS Compilation:** macOS environment with Xcode 16+ and iOS 18.0+ deployment target.
* **Backend Components:** Docker and Docker Compose installed

### 1. Backend Setup

#### 1. Cryptographic Keys & Certificates Setup
Before spinning up the containers, you must generate the RS256 JWT key pairs and the mTLS certificates for the three gates. Helper scripts are provided.

Run the following commands from the root of the repository:

```bash
# 1. Generate Gate-1 JWT Keys
chmod +x backend/backend/three-gate-login/gate-1/scripts/gen-gate1-keys.sh
cd backend/backend/three-gate-login/gate-1 && ./scripts/gen-gate1-keys.sh && cd ../../../../

# 2. Generate Gate-2 JWT Keys
chmod +x backend/backend/three-gate-login/gate-2/scripts/gen-gate2-keys.sh
cd backend/backend/three-gate-login/gate-2 && ./scripts/gen-gate2-keys.sh && cd ../../../../

# 3. Generate Gate-2 mTLS Certificates (Server & Client CA)
chmod +x backend/backend/three-gate-login/gate-2/scripts/gen-certs.sh
cd backend/backend/three-gate-login/gate-2 && ./scripts/gen-certs.sh && cd ../../../../

# 4. Distribute Public Keys for Cross-Gate Token Verification
# Gate-2 needs Gate-1's public key to verify the Gate-1 JWT
mkdir -p backend/backend/three-gate-login/gate-2/certs/gate1
cp backend/backend/three-gate-login/gate-1/certs/gate1/public.pem backend/backend/three-gate-login/gate-2/certs/gate1/public.pem

# Gate-3 needs Gate-2's public key to verify the Gate-2 JWT
mkdir -p backend/backend/three-gate-login/gate-3/certs/gate2
cp backend/backend/three-gate-login/gate-2/certs/gate2/public.pem backend/backend/three-gate-login/gate-3/certs/gate2/public.pem
```

*(Note: The client certificate generated in step 3 must be converted to a `.p12` file and placed in the Android app for local testing. See the Mobile App README for instructions).*

#### 2. Google Service Account (Mock)
Gate-1 requires a Google Service Account to verify Play Integrity tokens. For local development, create a mock JSON file to prevent the container from crashing:

```bash
mkdir -p backend/backend/three-gate-login/secrets
echo '{"type":"service_account","project_id":"ci-mock","private_key_id":"mock","private_key":"","client_email":"ci@mock.iam.gserviceaccount.com","client_id":"0","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token"}' > backend/backend/three-gate-login/secrets/three-gate-login-service-account.json
```

#### 3. Docker Setup
Spin up the backend pipeline, caching servers, and database schemas with a single command:
```bash
docker compose -f backend/infrastructure/docker-compose.yml up -d
docker compose -f backend/infrastructure/DAG/docker-compose.yml up -d
```

### 2. Mobile App Setup

#### Network Configuration (Important)
Before compiling, you must update the hardcoded backend IP addresses to point to your local machine (or `10.0.2.2` if using the Android Emulator). Update the IP in the following files:
* `mobile/shared/src/androidMain/kotlin/org/example/shield/gate/Constants.kt`
* `mobile/shared/src/commonMain/kotlin/org/example/shield/scanner/url/UrlReportClient.kt`

#### Gate-2 mTLS Setup (Local Development)
To test Gate-2 locally, you must generate a mock client certificate (`client.p12`) and place it in the Android resources directory. The app expects the keystore password to be `"password"`.

Run the following terminal commands to generate the certificate:

```bash
# 1. Create the raw res directory if it doesn't exist
mkdir -p mobile/androidApp/src/main/res/raw

# 2. Generate a new private key and a self-signed certificate
openssl req -x509 -newkey rsa:2048 -keyout client.key -out client.crt -days 365 -nodes -subj "/CN=shield-client-dev"

# 3. Bundle the key and certificate into a PKCS12 format (.p12) with password "password"
openssl pkcs12 -export -out mobile/androidApp/src/main/res/raw/client.p12 -inkey client.key -in client.crt -passout pass:password

# 4. Clean up the intermediate files
rm client.key client.crt
```
Once the IP addresses are updated and the `client.p12` file is in `mobile/androidApp/src/main/res/raw/`, you can build and run the app.

### Building the Project
Simply open the project directory in Android Studio and run a gradle sync. To build the debug APK via terminal:
```bash
cd mobile
./gradlew :androidApp:assembleDebug
```


---

## Exposed Ports

| Service | Port | Description |
| :--- | :--- | :--- |
| **Grafana** | `3000` | Analytics and Observability Dashboard |
| **PostgreSQL** | `5432` | Primary Database (`intel_db`) |
| **Redis** | `6379` | Nonce and session caching |
| **Gate 1** | `8081` | Device Authenticity API |
| **Gate 3** | `8082` | FIDO2 WebAuthn API |
| **Risk URL Engine**| `8083` | Phishing Report Ingestion API |
| **Kafka** | `9092` | Event stream broker |

---

## Future Enhancements
* **Apache Flink Analytics Engine:** Integrating Flink on top of Kafka to process complex sliding-window event patterns (CEP) and detect coordinated geographic botnets in real-time.
* **PostGIS and Spatial Heatmaps:** Enabling PostGIS extensions in PostgreSQL to visually render active cyber-attacks across districts in India based on live GPS coordinates.
* **Actual Airflow Takedown API Calls:** Replacing the Python logging stubs in the Airflow DAG with live, authenticated REST `requests` to the Google Play Developer API, CERT-In MISP TAXII servers, and TRAI DLT portals.
