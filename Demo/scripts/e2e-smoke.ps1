param(
    [string]$ApiBaseUrl = "http://localhost:18080",
    [string]$WebBaseUrl = "http://localhost:13000",
    [string]$MockBaseUrl = "http://localhost:8081"
)

$ErrorActionPreference = "Stop"
$tenantHeaders = @{
    "X-Demo-Tenant" = "tenant_demo"
    "Content-Type"  = "application/json"
}

function Assert-True {
    param([bool]$Condition, [string]$Message)
    if (-not $Condition) {
        throw "Smoke assertion failed: $Message"
    }
    Write-Host "PASS  $Message"
}

function Get-PartnerTransferCount {
    $response = Invoke-WebRequest -Uri "$MockBaseUrl/v1/transfers?page=1&api_key=northwind_mock_local_key" -UseBasicParsing
    $items = $response.Content | ConvertFrom-Json
    return @($items).Count
}

$health = Invoke-RestMethod -Uri "$ApiBaseUrl/healthz"
Assert-True ($health.status -eq "ok" -and $health.database -eq "connected") "Go API and SQL Server are healthy"

$page = Invoke-WebRequest -Uri $WebBaseUrl -UseBasicParsing
Assert-True ($page.StatusCode -eq 200) "Next.js serves the interview UI"

# Establish a known read-model baseline even when a rebuilt in-memory mock is
# paired with the intentionally retained SQL demo volume.
$null = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/demo/sync" -Headers $tenantHeaders -Body (@{ scenario = "" } | ConvertTo-Json)
$accountsBefore = (Invoke-RestMethod -Uri "$ApiBaseUrl/api/accounts" -Headers $tenantHeaders).accounts
Assert-True ($accountsBefore.Count -ge 2) "SQL read model contains synthetic linked accounts"

$snapshotBefore = Invoke-RestMethod -Uri "$ApiBaseUrl/api/accounts/acc_1029/transactions?refresh=false" -Headers $tenantHeaders
$refresh = Invoke-RestMethod -Uri "$ApiBaseUrl/api/accounts/acc_1029/transactions?refresh=true" -Headers $tenantHeaders
Assert-True ($refresh.refresh_started -eq $true) "Recent-transaction read schedules an asynchronous comparison"
Start-Sleep -Milliseconds 750
$unchanged = Invoke-RestMethod -Uri "$ApiBaseUrl/api/accounts/acc_1029/transactions?refresh=false" -Headers $tenantHeaders
Assert-True ([int64]$unchanged.version -eq [int64]$snapshotBefore.version) "Matching partner data does not advance the content version"

$external = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/demo/accounts/acc_1029/external-activity" -Headers $tenantHeaders -Body "{}"
$null = Invoke-RestMethod -Uri "$ApiBaseUrl/api/accounts/acc_1029/transactions?refresh=true" -Headers $tenantHeaders
$changed = $unchanged
for ($attempt = 0; $attempt -lt 20; $attempt++) {
    Start-Sleep -Milliseconds 150
    $changed = Invoke-RestMethod -Uri "$ApiBaseUrl/api/accounts/acc_1029/transactions?refresh=false" -Headers $tenantHeaders
    if ([int64]$changed.version -gt [int64]$unchanged.version) {
        break
    }
}
Assert-True ([int64]$changed.version -eq ([int64]$unchanged.version + 1)) "External activity advances the SQL content version exactly once"
Assert-True (@($changed.transactions | Where-Object id -eq $external.transaction.id).Count -eq 1) "Asynchronous refresh exposes the external transaction after commit"

$syncStatus = 0
try {
    $null = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/demo/sync" -Headers $tenantHeaders -Body (@{ scenario = "503" } | ConvertTo-Json)
} catch {
    $syncStatus = [int]$_.Exception.Response.StatusCode
}
Assert-True ($syncStatus -eq 502) "Exhausted Northwind 503 is surfaced as a bounded upstream failure"
$degraded = (Invoke-RestMethod -Uri "$ApiBaseUrl/api/accounts" -Headers $tenantHeaders).accounts
Assert-True (@($degraded).Count -eq @($accountsBefore).Count) "Northwind failure preserves the last-known-good SQL snapshot"
Assert-True (@($degraded | Where-Object { $_.freshness.state -ne "degraded" }).Count -eq 0) "Northwind failure marks account freshness degraded"
$null = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/demo/sync" -Headers $tenantHeaders -Body (@{ scenario = "" } | ConvertTo-Json)

$normalRequestID = "smoke-normal-$([guid]::NewGuid().ToString('N'))"
$normalBody = @{
    request_id     = $normalRequestID
    from_account_id = "acc_1029"
    to_account_id   = "acc_2042"
    amount          = "10.00"
    currency        = "USD"
    scenario        = ""
} | ConvertTo-Json
$normal = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/transfers" -Headers $tenantHeaders -Body $normalBody
$normalDuplicate = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/transfers" -Headers $tenantHeaders -Body $normalBody
Assert-True ($normal.status -eq "PENDING") "Definitive Northwind response remains pending, not complete"
Assert-True ($normalDuplicate.id -eq $normal.id) "Repeated Vantaca request returns the durable transfer intent"

$posted = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/demo/transfers/$($normal.id)/advance" -Headers $tenantHeaders -Body (@{ status = "POSTED"; deliveries = 2 } | ConvertTo-Json)
$returned = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/demo/transfers/$($normal.id)/advance" -Headers $tenantHeaders -Body (@{ status = "RETURNED"; deliveries = 1 } | ConvertTo-Json)
Assert-True ($posted.status -eq "POSTED") "Duplicate POSTED webhook deliveries reconcile idempotently"
Assert-True ($returned.status -eq "RETURNED") "A posted transfer can later become returned"

$partnerCountBefore = Get-PartnerTransferCount
$unknownRequestID = "smoke-unknown-$([guid]::NewGuid().ToString('N'))"
$unknownBody = @{
    request_id      = $unknownRequestID
    from_account_id = "acc_1029"
    to_account_id   = "acc_2042"
    amount           = "11.00"
    currency         = "USD"
    scenario         = "post-commit-timeout"
} | ConvertTo-Json
$unknown = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/transfers" -Headers $tenantHeaders -Body $unknownBody
$partnerCountAfterFirst = Get-PartnerTransferCount
$unknownDuplicate = Invoke-RestMethod -Method Post -Uri "$ApiBaseUrl/api/transfers" -Headers $tenantHeaders -Body $unknownBody
$partnerCountAfterDuplicate = Get-PartnerTransferCount
Assert-True ($unknown.status -eq "UNKNOWN" -and $unknown.error_category -eq "ambiguous_outcome") "Post-commit timeout is durable UNKNOWN"
Assert-True ($partnerCountAfterFirst -eq ($partnerCountBefore + 1)) "Ambiguous attempt issued one partner POST"
Assert-True ($unknownDuplicate.id -eq $unknown.id -and $partnerCountAfterDuplicate -eq $partnerCountAfterFirst) "Repeated ambiguous request issues no second partner POST"

Write-Host "`nEnd-to-end demo smoke test passed."
