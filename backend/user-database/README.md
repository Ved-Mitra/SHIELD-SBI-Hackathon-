# User Database

## Overview
This module contains the initial SQL schema for the core banking user database used by Gate-3 for storing WebAuthn/FIDO2 credentials.

## Current Implementation Details
- **Schema File**: `schema.sql`
- **Tables**: `users`, `webauthn_credentials`
- **Integration**: The schema is automatically mounted into the `shield-db` PostgreSQL container via `docker-compose.yml` (`/docker-entrypoint-initdb.d/0-user-schema.sql`).
