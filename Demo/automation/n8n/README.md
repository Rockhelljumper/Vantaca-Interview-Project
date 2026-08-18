# Optional n8n orchestration

The base demo does not require n8n. Start the optional operational control plane with:

```powershell
docker compose --profile automation up --build -d --wait
```

Open `http://localhost:5678`, complete local n8n setup, and import `/workflows/northwind-reconciliation.json` from the mounted container path (or import the repository file through the UI).

The workflow runs every five minutes and invokes one bounded Go reconciliation endpoint. Go—not n8n—selects accounts/transfers, calls Northwind, applies retry policy, owns state transitions, and writes SQL.

This is a demo of centralized recurring-operation history. Production adoption, credentials, access, retention, upgrades, and on-call ownership remain Vantaca platform decisions.
