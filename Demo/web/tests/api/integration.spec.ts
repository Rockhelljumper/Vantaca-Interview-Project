import {
  expect,
  request as playwrightRequest,
  test,
  type APIRequestContext,
  type APIResponse,
} from "@playwright/test";
import SwaggerParser from "@apidevtools/swagger-parser";
import path from "node:path";

const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:18080";
const MOCK_BASE_URL = process.env.MOCK_BASE_URL ?? "http://localhost:8081";
const DEMO_ADMIN_KEY = process.env.DEMO_ADMIN_KEY ?? "demo_admin_local_only";
const SWAGGER_BASE_URL = process.env.SWAGGER_BASE_URL ?? "http://localhost:18090";
const SWAGGER_ORIGIN = new URL(SWAGGER_BASE_URL).origin;
const TENANT = "tenant_demo";

type JsonObject = Record<string, any>;

test.describe.configure({ mode: "serial" });

let api: APIRequestContext;
let anonymousApi: APIRequestContext;
let mock: APIRequestContext;

test.beforeAll(async () => {
  api = await playwrightRequest.newContext({
    baseURL: API_BASE_URL,
    extraHTTPHeaders: {
      "Content-Type": "application/json",
      "X-Demo-Tenant": TENANT,
    },
  });
  anonymousApi = await playwrightRequest.newContext({ baseURL: API_BASE_URL });
  mock = await playwrightRequest.newContext({ baseURL: MOCK_BASE_URL });
});

test.afterAll(async () => {
  await Promise.all([api.dispose(), anonymousApi.dispose(), mock.dispose()]);
});

async function expectJson(response: APIResponse, status: number): Promise<JsonObject> {
  expect(response.status(), await response.text()).toBe(status);
  expect(response.headers()["content-type"]).toContain("application/json");
  return response.json() as Promise<JsonObject>;
}

async function synchronize(scenario = ""): Promise<APIResponse> {
  return api.post("/api/demo/sync", { data: { scenario } });
}

async function accounts(): Promise<JsonObject[]> {
  const body = await expectJson(await api.get("/api/accounts"), 200);
  return body.accounts;
}

async function transactions(accountID = "acc_1029"): Promise<JsonObject> {
  return expectJson(
    await api.get(`/api/accounts/${accountID}/transactions?refresh=false`),
    200,
  );
}

async function waitForTransactions(
  accountID: string,
  predicate: (payload: JsonObject) => boolean,
  timeout = 12_000,
): Promise<JsonObject> {
  const deadline = Date.now() + timeout;
  let latest = await transactions(accountID);
  while (!predicate(latest) && Date.now() < deadline) {
    await new Promise((resolve) => setTimeout(resolve, 150));
    latest = await transactions(accountID);
  }
  expect(predicate(latest), JSON.stringify(latest)).toBe(true);
  return latest;
}

async function partnerTransferCount(): Promise<number> {
  let total = 0;
  for (let page = 1; page <= 20; page += 1) {
    const response = await mock.get(
      `/v1/transfers?page=${page}&api_key=northwind_mock_local_key`,
    );
    const pageItems = (await expectJson(response, 200)) as unknown as JsonObject[];
    total += pageItems.length;
    if (pageItems.length < 50) return total;
  }
  throw new Error("Partner transfer pagination exceeded the test safety bound.");
}

function requestID(label: string): string {
  return `pw-${label}-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function transferBody(label: string, scenario = "", amount = "10.00"): JsonObject {
  return {
    request_id: requestID(label),
    from_account_id: "acc_1029",
    to_account_id: "acc_2042",
    amount,
    currency: "USD",
    scenario,
  };
}

test("health, correlation, security headers, tenant boundary, and demo metadata", async () => {
  const healthResponse = await anonymousApi.get("/healthz", {
    headers: { "X-Correlation-ID": "playwright-health-001" },
  });
  const health = await expectJson(healthResponse, 200);
  expect(health).toEqual({ status: "ok", database: "connected" });
  expect(healthResponse.headers()["x-correlation-id"]).toBe("playwright-health-001");
  expect(healthResponse.headers()["cache-control"]).toBe("no-store");
  expect(healthResponse.headers()["x-content-type-options"]).toBe("nosniff");
  expect(healthResponse.headers()["referrer-policy"]).toBe("no-referrer");

  const forbidden = await expectJson(await anonymousApi.get("/api/accounts"), 403);
  expect(forbidden.error).toBe("tenant_forbidden");
  expect(forbidden.correlation_id).toMatch(/^[a-f0-9]{24}$/);

  const info = await expectJson(await api.get("/api/demo/info"), 200);
  expect(info).toMatchObject({
    mode: "demo",
    customer_id: "customer_demo",
    northwind_authoritative: true,
    read_model: "SQL Server 2022",
    demo_controls_enabled: true,
    unsigned_webhook_mode: true,
    transfer_submission_enabled: true,
  });
  expect(info.production_blockers).toHaveLength(4);
  expect(info.demo_assumptions).toHaveLength(4);
});

test("Northwind mock contract is reachable and requires its synthetic API key", async () => {
  await expectJson(await mock.get("/healthz"), 200);
  const unauthorized = await expectJson(await mock.get("/v1/accounts?page=1"), 401);
  expect(unauthorized.error).toBeTruthy();
  const partnerAccounts = (await expectJson(
    await mock.get("/v1/accounts?page=1&api_key=northwind_mock_local_key"),
    200,
  )) as unknown as JsonObject[];
  expect(partnerAccounts).toHaveLength(3);
});

test("both OpenAPI contracts validate and the local Swagger explorer serves them", async () => {
  const coreSpecPath = path.resolve("..", "api", "internal", "httpapi", "openapi.yaml");
  const mockSpecPath = path.resolve(
    "..",
    "mock",
    "northwind",
    "internal",
    "mockapi",
    "openapi.yaml",
  );
  const coreSpec = (await SwaggerParser.validate(coreSpecPath)) as any;
  const mockSpec = (await SwaggerParser.validate(mockSpecPath)) as any;
  expect(coreSpec.openapi).toBe("3.0.3");
  expect(mockSpec.openapi).toBe("3.0.3");
  expect(Object.keys(coreSpec.paths)).toHaveLength(11);
  expect(Object.keys(mockSpec.paths)).toHaveLength(7);

  const liveCore = await anonymousApi.get("/openapi.yaml");
  expect(liveCore.status()).toBe(200);
  expect(liveCore.headers()["content-type"]).toContain("application/yaml");
  expect(await liveCore.text()).toContain("title: Vantaca Northwind Integration Demo API");
  const liveMock = await mock.get("/openapi.yaml");
  expect(liveMock.status()).toBe(200);
  expect(liveMock.headers()["content-type"]).toContain("application/yaml");
  expect(await liveMock.text()).toContain("title: Northwind Connect Deterministic Mock API");

  const corePreflight = await anonymousApi.fetch("/api/accounts", {
    method: "OPTIONS",
    headers: {
      Origin: SWAGGER_ORIGIN,
      "Access-Control-Request-Method": "GET",
      "Access-Control-Request-Headers": "X-Demo-Tenant",
    },
  });
  expect(corePreflight.status()).toBe(204);
  expect(corePreflight.headers()["access-control-allow-origin"]).toBe(SWAGGER_ORIGIN);
  const mockPreflight = await mock.fetch("/v1/accounts", {
    method: "OPTIONS",
    headers: {
      Origin: SWAGGER_ORIGIN,
      "Access-Control-Request-Method": "GET",
    },
  });
  expect(mockPreflight.status()).toBe(204);
  expect(mockPreflight.headers()["access-control-allow-origin"]).toBe(SWAGGER_ORIGIN);

  const explorer = await anonymousApi.get(`${SWAGGER_BASE_URL}/`);
  expect(explorer.status()).toBe(200);
  expect(await explorer.text()).toContain("Swagger UI");
  expect((await anonymousApi.get(`${SWAGGER_BASE_URL}/specs/vantaca-demo-openapi.yaml`)).status()).toBe(200);
  expect((await anonymousApi.get(`${SWAGGER_BASE_URL}/specs/northwind-mock-openapi.yaml`)).status()).toBe(200);
});

test("SQL account and transaction reads are masked, exact, and refresh asynchronously", async () => {
  await expectJson(await synchronize(), 200);
  const linkedAccounts = await accounts();
  expect(linkedAccounts).toHaveLength(3);
  for (const account of linkedAccounts) {
    expect(account.balance).toMatch(/^-?\d+\.\d{2}$/);
    expect(account.last_four).toMatch(/^\d{4}$/);
    expect(account).not.toHaveProperty("account_number");
    expect(account).not.toHaveProperty("routing_number");
  }

  const before = await transactions();
  expect(before.transactions.length).toBeGreaterThan(0);
  expect(before.invalidation).toBe("bounded version polling");
  expect(before.transactions[0].amount).toMatch(/^-?\d+\.\d{2}$/);

  const refresh = await expectJson(
    await api.get("/api/accounts/acc_1029/transactions?refresh=true"),
    200,
  );
  expect(refresh.refresh_started).toBe(true);
  const settled = await waitForTransactions("acc_1029", (payload) => !payload.refreshing);
  expect(settled.version).toBe(before.version);

  const missing = await expectJson(
    await api.get("/api/accounts/not-an-account/transactions?refresh=false"),
    404,
  );
  expect(missing.error).toBe("account_not_found");
  const invalid = await expectJson(
    await api.get("/api/accounts/acc_1029/transactions?refresh=true&scenario=unsupported"),
    400,
  );
  expect(invalid.error).toBe("invalid_scenario");
});

for (const scenario of ["429", "500", "503", "latency"] as const) {
  test(`safe-read ${scenario} failure is bounded and preserves the SQL snapshot`, async () => {
    await expectJson(await synchronize(), 200);
    const before = await accounts();
    const started = Date.now();
    const failure = await expectJson(await synchronize(scenario), 502);
    expect(Date.now() - started).toBeLessThan(20_000);
    expect(failure.error).toBe("northwind_sync_failed");
    expect(failure.message).toContain("last SQL snapshot was preserved");

    const degraded = await accounts();
    expect(degraded).toHaveLength(before.length);
    expect(degraded.every((account) => account.freshness.state === "degraded")).toBe(true);
    expect(degraded.map((account) => account.id)).toEqual(before.map((account) => account.id));

    await expectJson(await synchronize(), 200);
    expect((await accounts()).every((account) => account.freshness.state !== "degraded")).toBe(
      true,
    );
  });
}

test("external Northwind activity commits once before frontend-visible invalidation", async () => {
  await expectJson(await synchronize(), 200);
  const before = await transactions();
  const external = await expectJson(
    await api.post("/api/demo/accounts/acc_1029/external-activity", { data: {} }),
    201,
  );
  expect(external.transaction.description).toBe("EXTERNAL DEMO DEPOSIT");

  const refresh = await expectJson(
    await api.get("/api/accounts/acc_1029/transactions?refresh=true"),
    200,
  );
  expect(refresh.refresh_started).toBe(true);
  const changed = await waitForTransactions(
    "acc_1029",
    (payload) => payload.version > before.version,
  );
  expect(changed.version).toBe(before.version + 1);
  expect(
    changed.transactions.filter((item: JsonObject) => item.id === external.transaction.id),
  ).toHaveLength(1);

  await expectJson(
    await api.get("/api/accounts/acc_1029/transactions?refresh=true"),
    200,
  );
  const rechecked = await waitForTransactions("acc_1029", (payload) => !payload.refreshing);
  expect(rechecked.version).toBe(changed.version);
});

test("external activity rejects closed or unknown accounts before calling Northwind", async () => {
  const closed = await expectJson(
    await api.post("/api/demo/accounts/acc_3097/external-activity", { data: {} }),
    422,
  );
  expect(closed.error).toBe("external_activity_unavailable");
  expect(closed.message).toContain("open account");

  const missing = await expectJson(
    await api.post("/api/demo/accounts/not-an-account/external-activity", { data: {} }),
    404,
  );
  expect(missing.error).toBe("account_not_found");
});

test("transfer validation rejects unsafe or malformed requests", async () => {
  const valid = transferBody("validation");
  const cases: Array<[string, JsonObject]> = [
    ["unsafe request identity", { ...valid, request_id: "short" }],
    ["same source and destination", { ...valid, to_account_id: valid.from_account_id }],
    ["fractional precision", { ...valid, amount: "1.001" }],
    ["unsupported currency", { ...valid, currency: "EUR" }],
    ["unsupported scenario", { ...valid, scenario: "retry-money" }],
  ];

  for (const [label, body] of cases) {
    const response = await expectJson(await api.post("/api/transfers", { data: body }), 422);
    expect(response.error, label).toBe("transfer_validation_failed");
  }
});

test("normal transfer is submitted once, deduplicated, posted by duplicate webhook, then returned", async () => {
  const beforeCount = await partnerTransferCount();
  const body = transferBody("normal", "", "23.45");
  const first = await expectJson(await api.post("/api/transfers", { data: body }), 202);
  const afterFirst = await partnerTransferCount();
  const duplicate = await expectJson(await api.post("/api/transfers", { data: body }), 202);
  const afterDuplicate = await partnerTransferCount();

  expect(first.status).toBe("PENDING");
  expect(first.amount).toBe("23.45");
  expect(duplicate.id).toBe(first.id);
  expect(afterFirst).toBe(beforeCount + 1);
  expect(afterDuplicate).toBe(afterFirst);

  const posted = await expectJson(
    await api.post(`/api/demo/transfers/${first.id}/advance`, {
      data: { status: "POSTED", deliveries: 2 },
    }),
    200,
  );
  expect(posted.status).toBe("POSTED");

  const duplicateWebhook = await expectJson(
    await anonymousApi.post("/api/webhooks/northwind", {
      data: {
        event: "transfer.updated",
        transfer_id: first.partner_transfer_id,
        status: "POSTED",
      },
    }),
    200,
  );
  expect(duplicateWebhook).toEqual({ status: "duplicate_acknowledged", trusted: false });

  const returned = await expectJson(
    await api.post(`/api/demo/transfers/${first.id}/advance`, {
      data: { status: "RETURNED", deliveries: 1 },
    }),
    200,
  );
  expect(returned.status).toBe("RETURNED");
  const listed = await expectJson(await api.get("/api/transfers"), 200);
  expect(
    listed.transfers.some(
      (transfer: JsonObject) => transfer.id === first.id && transfer.status === "RETURNED",
    ),
  ).toBe(true);
});

test("a pending transfer can reach a definitive FAILED state", async () => {
  const submitted = await expectJson(
    await api.post("/api/transfers", { data: transferBody("failed", "", "9.87") }),
    202,
  );
  expect(submitted.status).toBe("PENDING");
  const failed = await expectJson(
    await api.post(`/api/demo/transfers/${submitted.id}/advance`, {
      data: { status: "FAILED", deliveries: 1 },
    }),
    200,
  );
  expect(failed.status).toBe("FAILED");
});

for (const scenario of ["post-commit-timeout", "503", "500"] as const) {
  test(`monetary ${scenario} remains durable UNKNOWN and is never resubmitted`, async () => {
    const beforeCount = await partnerTransferCount();
    const body = transferBody(`unknown-${scenario}`, scenario, "11.11");
    const first = await expectJson(await api.post("/api/transfers", { data: body }), 202);
    const afterFirst = await partnerTransferCount();
    const duplicate = await expectJson(await api.post("/api/transfers", { data: body }), 202);
    const afterDuplicate = await partnerTransferCount();

    expect(first.status).toBe("UNKNOWN");
    expect(first.error_category).toBe("ambiguous_outcome");
    expect(duplicate.id).toBe(first.id);
    expect(afterDuplicate).toBe(afterFirst);
    if (scenario === "post-commit-timeout") {
      expect(afterFirst).toBe(beforeCount + 1);
    } else {
      expect(afterFirst).toBe(beforeCount);
    }

    const unsafeAdvance = await expectJson(
      await api.post(`/api/demo/transfers/${first.id}/advance`, {
        data: { status: "POSTED", deliveries: 1 },
      }),
      422,
    );
    expect(unsafeAdvance.error).toBe("invalid_transition");
  });
}

test("webhook schema and internal reconciliation authorization are enforced", async () => {
  const invalidWebhook = await expectJson(
    await anonymousApi.post("/api/webhooks/northwind", {
      data: { event: "wrong.event", transfer_id: "transfer_demo", status: "POSTED" },
    }),
    400,
  );
  expect(invalidWebhook.error).toBe("invalid_webhook");

  const unauthorized = await expectJson(
    await anonymousApi.post("/api/internal/reconcile"),
    401,
  );
  expect(unauthorized.error).toBe("unauthorized");
  const authorized = await anonymousApi.post("/api/internal/reconcile", {
    headers: { "X-Demo-Admin-Key": DEMO_ADMIN_KEY },
  });
  expect([200, 502]).toContain(authorized.status());
  const result = await authorized.json();
  expect(["complete", "partial"]).toContain(result.status);
});
