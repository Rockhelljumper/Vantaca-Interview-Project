CREATE OR ALTER VIEW dbo.vw_account_sync_health
AS
SELECT
    l.tenant_id,
    a.northwind_account_id,
    a.checked_at,
    a.fetched_at,
    DATEDIFF(SECOND, a.fetched_at, SYSUTCDATETIME()) AS stale_age_seconds,
    a.last_sync_error,
    a.data_version
FROM dbo.linked_accounts AS a
INNER JOIN dbo.northwind_links AS l ON l.link_id = a.link_id
WHERE l.status = 'active';
