CREATE TABLE account(

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    balance BIGINT NOT NULL CHECK (balance >= 0) DEFAULT 0,
    user_name VARCHAR(120) NOT NULL,
    user_cpf_cnpj VARCHAR(14) UNIQUE NOT NULL, 
    user_email VARCHAR(120) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    blocked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP

);



CREATE INDEX idx_account_cpf_cnpj ON account(user_cpf_cnpj);
CREATE INDEX idx_account_email ON account(user_email);
CREATE INDEX idx_account_blocked ON account(blocked);

