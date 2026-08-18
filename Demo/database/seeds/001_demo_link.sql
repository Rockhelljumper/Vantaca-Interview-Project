IF NOT EXISTS (
    SELECT 1
    FROM dbo.northwind_links
    WHERE tenant_id = 'tenant_demo'
      AND customer_reference = 'customer_demo'
)
BEGIN
    INSERT INTO dbo.northwind_links (
        link_id,
        tenant_id,
        customer_reference,
        status
    )
    VALUES (
        '11111111-1111-4111-8111-111111111111',
        'tenant_demo',
        'customer_demo',
        'active'
    );
END;
