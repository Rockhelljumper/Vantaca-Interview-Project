import { expect, test, type Locator, type Page } from "@playwright/test";

const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:18080";
const SWAGGER_BASE_URL = process.env.SWAGGER_BASE_URL ?? "http://localhost:18090";
const tenantHeaders = {
  "Content-Type": "application/json",
  "X-Demo-Tenant": "tenant_demo",
};

test.describe.configure({ mode: "serial" });

async function openDashboard(page: Page): Promise<void> {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "Linked account snapshots" })).toBeVisible();
  await expect(page.getByTestId("account-acc_1029")).toBeVisible();
}

async function expectTooltip(button: Locator, text: string, page: Page): Promise<void> {
  const tooltipID = await button.getAttribute("aria-describedby");
  expect(tooltipID).toBeTruthy();
  if (await button.isDisabled()) {
    await button.locator("..").focus();
  } else {
    await button.focus();
  }
  const tooltip = page.locator(`[id="${tooltipID}"]`);
  await expect(tooltip).toContainText("What this tests");
  await expect(tooltip).toContainText(text);
  await expect(tooltip).toHaveCSS("opacity", "1");
}

async function submitTransfer(page: Page, amount: string, scenario = ""): Promise<Locator> {
  await page.getByLabel("From").selectOption("acc_1029");
  await page.getByLabel("To").selectOption("acc_2042");
  await page.getByLabel("Amount (USD)").fill(amount);
  await page.getByLabel("Submission behavior").selectOption(scenario);
  await page.getByRole("button", { name: "Submit demo transfer" }).click();
  const newest = page.getByTestId("transfer-card").first();
  await expect(newest).toBeVisible();
  return newest;
}

test.beforeEach(async ({ request }) => {
  const response = await request.post(`${API_BASE_URL}/api/demo/sync`, {
    headers: tenantHeaders,
    data: { scenario: "" },
  });
  expect(response.ok(), await response.text()).toBe(true);
});

test("branded architecture, production boundaries, base tooltips, and mobile layout render", async ({
  page,
}) => {
  await openDashboard(page);
  await expect(page.getByAltText("Vantaca")).toHaveAttribute(
    "src",
    "https://www.vantaca.com/hubfs/Asset%204.svg",
  );
  await expect(
    page.getByRole("heading", { name: /Connected financial experiences/ }),
  ).toBeVisible();
  await expect(page.getByRole("heading", { name: "Recent transactions" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Read-path failure lab" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Guarded ACH transfer" })).toBeVisible();
  await expect(page.getByText("Still blocks production")).toBeVisible();
  await expect(page.getByText(/Synthetic interview environment/)).toBeVisible();
  await expect(page.getByRole("link", { name: /API explorer/ })).toHaveAttribute(
    "href",
    SWAGGER_BASE_URL,
  );

  await expectTooltip(
    page.getByRole("button", { name: "Simulate external deposit" }),
    "eventual consistency",
    page,
  );
  await expectTooltip(
    page.getByRole("button", { name: "Run account + transaction sync" }),
    "successful bounded synchronization",
    page,
  );
  await expectTooltip(
    page.getByRole("button", { name: "Submit demo transfer" }),
    "durable intent",
    page,
  );

  await page.setViewportSize({ width: 390, height: 844 });
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= document.documentElement.clientWidth,
    ),
  ).toBe(true);
  await expect(page.getByAltText("Vantaca")).toBeVisible();
});

test("Swagger UI renders both the core and Northwind mock contracts", async ({ page }) => {
  await page.goto(SWAGGER_BASE_URL);
  await expect(page.locator("#swagger-ui")).toBeVisible();
  await expect(
    page.getByRole("heading", { name: /Vantaca Northwind Integration Demo API/ }),
  ).toBeVisible({ timeout: 15_000 });
  const contractSelector = page.locator(".download-url-wrapper select");
  await expect(contractSelector.locator("option")).toHaveCount(2);
  await contractSelector.selectOption({ label: "Northwind Mock API" });
  await expect(
    page.getByRole("heading", { name: /Northwind Connect Deterministic Mock API/ }),
  ).toBeVisible({ timeout: 15_000 });
});

test("account selection and external activity refresh the SQL-backed transaction view", async ({
  page,
}) => {
  await openDashboard(page);
  const savings = page.getByTestId("account-acc_2042");
  await savings.click();
  await expect(savings).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByRole("table", { name: "Recent transactions" })).toContainText(
    "INTEREST CREDIT",
  );

  const closed = page.getByTestId("account-acc_3097");
  await closed.click();
  await expect(closed).toHaveAttribute("aria-pressed", "true");
  const unavailable = page.getByRole("button", { name: "External activity unavailable" });
  await expect(unavailable).toBeDisabled();
  await expectTooltip(unavailable, "requires an open account", page);

  const checking = page.getByTestId("account-acc_1029");
  await checking.click();
  await expect(checking).toHaveAttribute("aria-pressed", "true");
  await page.getByRole("button", { name: "Simulate external deposit" }).click();
  await expect(page.getByRole("status")).toContainText("Northwind changed outside Vantaca");
  await expect(page.getByRole("status")).toContainText("database committed first", {
    timeout: 15_000,
  });
  await expect(page.getByRole("table", { name: "Recent transactions" })).toContainText(
    "EXTERNAL DEMO DEPOSIT",
  );
});

test("every safe-read scenario preserves data and exposes degraded freshness", async ({ page }) => {
  await openDashboard(page);
  const scenarioSelect = page.getByLabel("Northwind behavior");
  const runButton = page.getByRole("button", { name: "Run account + transaction sync" });
  const cases = [
    ["429", "Retry-After handling"],
    ["500", "transient partner failure"],
    ["503", "partner unavailability"],
    ["latency", "client timeout enforcement"],
  ] as const;

  await expect(scenarioSelect.locator("option")).toHaveCount(5);
  for (const [scenario, tooltipText] of cases) {
    await scenarioSelect.selectOption(scenario);
    await expectTooltip(runButton, tooltipText, page);
    await runButton.click();
    await expect(
      page.getByRole("alert").filter({ hasText: "Northwind refresh failed" }),
    ).toContainText("last SQL snapshot was preserved", { timeout: 20_000 });
    await expect(page.getByTestId("account-acc_1029")).toContainText("Refresh degraded");

    await scenarioSelect.selectOption("");
    await runButton.click();
    await expect(page.getByRole("status")).toContainText("synchronization completed");
    await expect(page.getByTestId("account-acc_1029")).toContainText("Fresh snapshot");
  }
});

test("normal transfers demonstrate duplicate webhook, return, and definitive failure paths", async ({
  page,
}) => {
  await openDashboard(page);
  let newest = await submitTransfer(page, "31.25");
  await expect(newest).toHaveAttribute("data-transfer-status", "PENDING");
  await expect(newest).toContainText("$31.25");

  const postButton = newest.getByRole("button", { name: "Post + duplicate webhook" });
  await expectTooltip(postButton, "webhook idempotency", page);
  await postButton.click();
  newest = page.getByTestId("transfer-card").first();
  await expect(newest).toHaveAttribute("data-transfer-status", "POSTED");
  await expect(page.getByRole("status")).toContainText("Two identical webhooks were delivered");

  const returnButton = newest.getByRole("button", { name: "Demonstrate late return" });
  await expectTooltip(returnButton, "late POSTED → RETURNED", page);
  await returnButton.click();
  await expect(page.getByTestId("transfer-card").first()).toHaveAttribute(
    "data-transfer-status",
    "RETURNED",
  );

  newest = await submitTransfer(page, "32.50");
  const failButton = newest.getByRole("button", { name: "Fail" });
  await expectTooltip(failButton, "definitive PENDING → FAILED", page);
  await failButton.click();
  await expect(page.getByTestId("transfer-card").first()).toHaveAttribute(
    "data-transfer-status",
    "FAILED",
  );
});

test("all ambiguous monetary outcomes stay UNKNOWN in the UI", async ({ page }) => {
  await openDashboard(page);
  const submitButton = page.getByRole("button", { name: "Submit demo transfer" });
  const cases = [
    ["post-commit-timeout", "partner commit with a lost response", "41.01"],
    ["503", "monetary 503", "41.02"],
    ["500", "monetary 500", "41.03"],
  ] as const;

  for (const [scenario, tooltipText, amount] of cases) {
    await page.getByLabel("Submission behavior").selectOption(scenario);
    await expectTooltip(submitButton, tooltipText, page);
    const newest = await submitTransfer(page, amount, scenario);
    await expect(newest).toHaveAttribute("data-transfer-status", "UNKNOWN", {
      timeout: 20_000,
    });
    await expect(newest).toContainText("Do not submit it again");
    await expect(page.getByRole("status")).toContainText("outcome is unknown");
  }
});

test("transfer form surfaces backend safety validation", async ({ page }) => {
  await openDashboard(page);
  await page.getByLabel("From").selectOption("acc_1029");
  await page.getByLabel("To").selectOption("acc_1029");
  await page.getByLabel("Amount (USD)").fill("10.00");
  await page.getByRole("button", { name: "Submit demo transfer" }).click();
  await expect(
    page.getByRole("alert").filter({ hasText: "source and destination accounts" }),
  ).toContainText("source and destination accounts must differ");

  await page.getByLabel("To").selectOption("acc_2042");
  await page.getByLabel("Amount (USD)").fill("1.001");
  await page.getByRole("button", { name: "Submit demo transfer" }).click();
  await expect(page.getByRole("alert").filter({ hasText: "at most two decimals" })).toContainText(
    "at most two decimals",
  );
});
