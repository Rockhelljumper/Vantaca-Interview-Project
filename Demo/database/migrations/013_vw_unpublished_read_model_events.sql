CREATE OR ALTER VIEW dbo.vw_unpublished_read_model_events
AS
SELECT
    event_id,
    linked_account_id,
    aggregate_version,
    event_type,
    created_at,
    publish_attempts
FROM dbo.read_model_outbox
WHERE published_at IS NULL;
