CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE subscriptions (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    start_date DATE NOT NULL,
    end_date DATE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_user_id ON subscriptions (user_id);
CREATE INDEX idx_service_name ON subscriptions (service_name);