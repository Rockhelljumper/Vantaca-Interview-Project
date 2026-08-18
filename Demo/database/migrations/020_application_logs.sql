CREATE TABLE dbo.application_logs (
    application_log_id BIGINT IDENTITY(1,1) NOT NULL,
    occurred_at DATETIMEOFFSET(7) NOT NULL,
    application_name NVARCHAR(64) NOT NULL,
    severity VARCHAR(8) NOT NULL,
    event_name NVARCHAR(256) NOT NULL,
    correlation_id NVARCHAR(64) NULL,
    username NVARCHAR(128) NULL,
    api_key_last_four CHAR(4) NULL,
    attributes_json NVARCHAR(MAX) NOT NULL,
    recorded_at DATETIMEOFFSET(7) NOT NULL CONSTRAINT DF_application_logs_recorded_at DEFAULT SYSUTCDATETIME(),
    CONSTRAINT PK_application_logs PRIMARY KEY (application_log_id),
    CONSTRAINT CK_application_logs_severity CHECK (severity IN ('DEBUG', 'INFO', 'WARN', 'ERROR')),
    CONSTRAINT CK_application_logs_attributes_json CHECK (ISJSON(attributes_json) = 1)
);

CREATE INDEX IX_application_logs_occurred
    ON dbo.application_logs(occurred_at DESC)
    INCLUDE (application_name, severity, event_name, correlation_id);

CREATE INDEX IX_application_logs_correlation
    ON dbo.application_logs(correlation_id, occurred_at DESC)
    INCLUDE (application_name, severity, event_name)
    WHERE correlation_id IS NOT NULL;

CREATE INDEX IX_application_logs_errors
    ON dbo.application_logs(severity, occurred_at DESC)
    INCLUDE (application_name, event_name, correlation_id, username)
    WHERE severity IN ('WARN', 'ERROR');
