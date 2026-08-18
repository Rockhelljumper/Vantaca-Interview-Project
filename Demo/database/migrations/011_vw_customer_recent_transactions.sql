CREATE OR ALTER VIEW dbo.vw_customer_recent_transactions
AS
SELECT
    l.tenant_id,
    a.northwind_account_id,
    a.data_version,
    a.fetched_at,
    a.checked_at,
    a.last_sync_error,
    t.northwind_transaction_id,
    t.amount_minor,
    t.currency,
    t.description,
    t.merchant_category_code,
    t.posted_at
FROM dbo.account_transactions AS t
INNER JOIN dbo.linked_accounts AS a ON a.linked_account_id = t.linked_account_id
INNER JOIN dbo.northwind_links AS l ON l.link_id = a.link_id
WHERE l.status = 'active';
