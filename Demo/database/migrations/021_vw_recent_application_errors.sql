CREATE OR ALTER VIEW dbo.vw_recent_application_errors
AS
SELECT
    application_log_id,
    occurred_at,
    application_name,
    severity,
    event_name,
    correlation_id,
    username,
    api_key_last_four,
    attributes_json
FROM dbo.application_logs
WHERE severity IN ('WARN', 'ERROR');
