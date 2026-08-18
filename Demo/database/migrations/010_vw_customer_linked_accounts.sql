CREATE OR ALTER VIEW dbo.vw_customer_linked_accounts
AS
SELECT
    l.tenant_id,
    l.customer_reference,
    a.linked_account_id,
    a.northwind_account_id,
    a.account_type,
    a.last_four,
    a.balance_minor,
    a.currency,
    a.status,
    a.data_version,
    a.fetched_at,
    a.checked_at,
    a.last_sync_error
FROM dbo.linked_accounts AS a
INNER JOIN dbo.northwind_links AS l ON l.link_id = a.link_id
WHERE l.status = 'active';
