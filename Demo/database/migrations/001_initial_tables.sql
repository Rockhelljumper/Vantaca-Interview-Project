CREATE TABLE dbo.northwind_links (
    link_id UNIQUEIDENTIFIER NOT NULL,
    tenant_id NVARCHAR(64) NOT NULL,
    customer_reference NVARCHAR(128) NOT NULL,
    status NVARCHAR(16) NOT NULL,
    created_at DATETIMEOFFSET(7) NOT NULL CONSTRAINT DF_northwind_links_created_at DEFAULT SYSUTCDATETIME(),
    revoked_at DATETIMEOFFSET(7) NULL,
    CONSTRAINT PK_northwind_links PRIMARY KEY (link_id),
    CONSTRAINT UQ_northwind_links_tenant_customer UNIQUE (tenant_id, customer_reference),
    CONSTRAINT CK_northwind_links_status CHECK (status IN ('active', 'revoked'))
);

CREATE TABLE dbo.linked_accounts (
    linked_account_id BIGINT IDENTITY(1,1) NOT NULL,
    link_id UNIQUEIDENTIFIER NOT NULL,
    northwind_account_id NVARCHAR(100) NOT NULL,
    account_type NVARCHAR(20) NOT NULL,
    last_four CHAR(4) NOT NULL,
    balance_minor BIGINT NOT NULL,
    currency CHAR(3) NOT NULL,
    status NVARCHAR(16) NOT NULL,
    data_version BIGINT NOT NULL CONSTRAINT DF_linked_accounts_data_version DEFAULT 1,
    fetched_at DATETIMEOFFSET(7) NOT NULL,
    checked_at DATETIMEOFFSET(7) NOT NULL,
    last_sync_error NVARCHAR(64) NULL,
    created_at DATETIMEOFFSET(7) NOT NULL CONSTRAINT DF_linked_accounts_created_at DEFAULT SYSUTCDATETIME(),
    updated_at DATETIMEOFFSET(7) NOT NULL CONSTRAINT DF_linked_accounts_updated_at DEFAULT SYSUTCDATETIME(),
    CONSTRAINT PK_linked_accounts PRIMARY KEY (linked_account_id),
    CONSTRAINT FK_linked_accounts_link FOREIGN KEY (link_id) REFERENCES dbo.northwind_links(link_id),
    CONSTRAINT UQ_linked_accounts_link_partner UNIQUE (link_id, northwind_account_id),
    CONSTRAINT CK_linked_accounts_type CHECK (account_type IN ('checking', 'savings')),
    CONSTRAINT CK_linked_accounts_status CHECK (status IN ('open', 'closed')),
    CONSTRAINT CK_linked_accounts_currency CHECK (currency LIKE '[A-Z][A-Z][A-Z]'),
    CONSTRAINT CK_linked_accounts_last_four CHECK (last_four NOT LIKE '%[^0-9]%')
);

CREATE INDEX IX_linked_accounts_link_status
    ON dbo.linked_accounts(link_id, status)
    INCLUDE (northwind_account_id, balance_minor, currency, fetched_at, data_version);

CREATE TABLE dbo.account_transactions (
    account_transaction_id BIGINT IDENTITY(1,1) NOT NULL,
    linked_account_id BIGINT NOT NULL,
    northwind_transaction_id NVARCHAR(100) NOT NULL,
    amount_minor BIGINT NOT NULL,
    currency CHAR(3) NOT NULL,
    description NVARCHAR(256) NOT NULL,
    merchant_category_code CHAR(4) NULL,
    posted_at DATETIMEOFFSET(7) NOT NULL,
    first_seen_at DATETIMEOFFSET(7) NOT NULL CONSTRAINT DF_account_transactions_first_seen DEFAULT SYSUTCDATETIME(),
    last_seen_at DATETIMEOFFSET(7) NOT NULL CONSTRAINT DF_account_transactions_last_seen DEFAULT SYSUTCDATETIME(),
    CONSTRAINT PK_account_transactions PRIMARY KEY (account_transaction_id),
    CONSTRAINT FK_account_transactions_account FOREIGN KEY (linked_account_id) REFERENCES dbo.linked_accounts(linked_account_id) ON DELETE CASCADE,
    CONSTRAINT UQ_account_transactions_account_partner UNIQUE (linked_account_id, northwind_transaction_id),
    CONSTRAINT CK_account_transactions_currency CHECK (currency LIKE '[A-Z][A-Z][A-Z]'),
    CONSTRAINT CK_account_transactions_mcc CHECK (merchant_category_code IS NULL OR merchant_category_code NOT LIKE '%[^0-9]%')
);

CREATE INDEX IX_account_transactions_recent
    ON dbo.account_transactions(linked_account_id, posted_at DESC)
    INCLUDE (northwind_transaction_id, amount_minor, currency, description, merchant_category_code);

CREATE TABLE dbo.sync_runs (
    sync_run_id UNIQUEIDENTIFIER NOT NULL,
    link_id UNIQUEIDENTIFIER NOT NULL,
    northwind_account_id NVARCHAR(100) NULL,
    sync_type NVARCHAR(32) NOT NULL,
    status NVARCHAR(16) NOT NULL,
    started_at DATETIMEOFFSET(7) NOT NULL,
    completed_at DATETIMEOFFSET(7) NULL,
    record_count INT NOT NULL CONSTRAINT DF_sync_runs_record_count DEFAULT 0,
    safe_error_category NVARCHAR(64) NULL,
    CONSTRAINT PK_sync_runs PRIMARY KEY (sync_run_id),
    CONSTRAINT FK_sync_runs_link FOREIGN KEY (link_id) REFERENCES dbo.northwind_links(link_id),
    CONSTRAINT CK_sync_runs_status CHECK (status IN ('running', 'succeeded', 'failed'))
);

CREATE INDEX IX_sync_runs_link_started
    ON dbo.sync_runs(link_id, started_at DESC);

CREATE TABLE dbo.read_model_outbox (
    event_id UNIQUEIDENTIFIER NOT NULL,
    linked_account_id BIGINT NOT NULL,
    aggregate_version BIGINT NOT NULL,
    event_type NVARCHAR(64) NOT NULL,
    created_at DATETIMEOFFSET(7) NOT NULL CONSTRAINT DF_read_model_outbox_created_at DEFAULT SYSUTCDATETIME(),
    published_at DATETIMEOFFSET(7) NULL,
    publish_attempts INT NOT NULL CONSTRAINT DF_read_model_outbox_attempts DEFAULT 0,
    CONSTRAINT PK_read_model_outbox PRIMARY KEY (event_id),
    CONSTRAINT FK_read_model_outbox_account FOREIGN KEY (linked_account_id) REFERENCES dbo.linked_accounts(linked_account_id),
    CONSTRAINT UQ_read_model_outbox_account_version_type UNIQUE (linked_account_id, aggregate_version, event_type)
);

CREATE INDEX IX_read_model_outbox_unpublished
    ON dbo.read_model_outbox(published_at, created_at)
    INCLUDE (linked_account_id, aggregate_version, event_type, publish_attempts);

CREATE TABLE dbo.transfers (
    internal_transfer_id UNIQUEIDENTIFIER NOT NULL,
    tenant_id NVARCHAR(64) NOT NULL,
    request_id NVARCHAR(100) NOT NULL,
    from_account_id NVARCHAR(100) NOT NULL,
    to_account_id NVARCHAR(100) NOT NULL,
    from_display NVARCHAR(64) NOT NULL,
    to_display NVARCHAR(64) NOT NULL,
    amount_minor BIGINT NOT NULL,
    currency CHAR(3) NOT NULL,
    status NVARCHAR(40) NOT NULL,
    partner_transfer_id NVARCHAR(100) NULL,
    last_error_category NVARCHAR(64) NULL,
    created_at DATETIMEOFFSET(7) NOT NULL,
    updated_at DATETIMEOFFSET(7) NOT NULL,
    CONSTRAINT PK_transfers PRIMARY KEY (internal_transfer_id),
    CONSTRAINT UQ_transfers_tenant_request UNIQUE (tenant_id, request_id),
    CONSTRAINT CK_transfers_distinct_accounts CHECK (from_account_id <> to_account_id),
    CONSTRAINT CK_transfers_amount CHECK (amount_minor > 0),
    CONSTRAINT CK_transfers_currency CHECK (currency LIKE '[A-Z][A-Z][A-Z]'),
    CONSTRAINT CK_transfers_status CHECK (status IN ('INTENT_RECORDED', 'PENDING', 'POSTED', 'FAILED', 'RETURNED', 'UNKNOWN'))
);

CREATE UNIQUE INDEX UX_transfers_partner_id
    ON dbo.transfers(partner_transfer_id)
    WHERE partner_transfer_id IS NOT NULL;

CREATE INDEX IX_transfers_tenant_created
    ON dbo.transfers(tenant_id, created_at DESC)
    INCLUDE (request_id, status, partner_transfer_id, amount_minor, currency);

CREATE TABLE dbo.transfer_status_history (
    transfer_status_history_id BIGINT IDENTITY(1,1) NOT NULL,
    internal_transfer_id UNIQUEIDENTIFIER NOT NULL,
    status NVARCHAR(40) NOT NULL,
    source NVARCHAR(32) NOT NULL,
    safe_detail NVARCHAR(128) NULL,
    recorded_at DATETIMEOFFSET(7) NOT NULL CONSTRAINT DF_transfer_history_recorded_at DEFAULT SYSUTCDATETIME(),
    CONSTRAINT PK_transfer_status_history PRIMARY KEY (transfer_status_history_id),
    CONSTRAINT FK_transfer_history_transfer FOREIGN KEY (internal_transfer_id) REFERENCES dbo.transfers(internal_transfer_id),
    CONSTRAINT CK_transfer_history_status CHECK (status IN ('INTENT_RECORDED', 'PENDING', 'POSTED', 'FAILED', 'RETURNED', 'UNKNOWN'))
);

CREATE INDEX IX_transfer_history_transfer_time
    ON dbo.transfer_status_history(internal_transfer_id, recorded_at);

CREATE TABLE dbo.webhook_inbox (
    webhook_receipt_id UNIQUEIDENTIFIER NOT NULL,
    dedupe_key NVARCHAR(240) NOT NULL,
    partner_transfer_id NVARCHAR(100) NOT NULL,
    reported_status NVARCHAR(40) NOT NULL,
    payload_sha256 CHAR(64) NOT NULL,
    received_at DATETIMEOFFSET(7) NOT NULL,
    processed_at DATETIMEOFFSET(7) NULL,
    processing_outcome NVARCHAR(64) NULL,
    CONSTRAINT PK_webhook_inbox PRIMARY KEY (webhook_receipt_id),
    CONSTRAINT UQ_webhook_inbox_dedupe UNIQUE (dedupe_key)
);

CREATE INDEX IX_webhook_inbox_unprocessed
    ON dbo.webhook_inbox(processed_at, received_at)
    INCLUDE (partner_transfer_id, reported_status);
