-- =========================================================================
-- SHIELD Core Banking Reference Database Schema
-- =========================================================================

CREATE TABLE IF NOT EXISTS banking_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    mpin VARCHAR(64) NOT NULL, -- Stored as plain text here for hackathon reference, but should be hashed in production
    account_number VARCHAR(20) UNIQUE NOT NULL,
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =========================================================================
-- Seed Data: 5 Mock Users
-- =========================================================================

INSERT INTO banking_users (username, mpin, account_number, phone_number, is_admin) VALUES
-- Admin User
('admin_ved', '1234', '00000000001', '+91-9876543210', TRUE),
('admin_aditya', '5678', '11223344556', '+91-1112223334', TRUE),

-- Regular Users
('aakarsh_sinha', '4321', '99887766554', '+91-5556667778', FALSE),
('mayank_tiwari', '1111', '33445566778', '+91-9998887776', FALSE),
('brijesh_thakkar', '9999', '55667788990', '+91-4445556667', FALSE)
ON CONFLICT (username) DO NOTHING;
