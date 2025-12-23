CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id VARCHAR(255) NOT NULL,
    from_account_id UUID NOT NULL,
    to_account_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency VARCHAR(10) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_transactions_from_account ON transactions(from_account_id);

CREATE INDEX idx_transactions_to_account ON transactions(to_account_id);

CREATE INDEX idx_transactions_status ON transactions(status);

CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);

CREATE INDEX idx_transactions_type ON transactions(type);

CREATE INDEX idx_transactions_from_status ON transactions(from_account_id, status);