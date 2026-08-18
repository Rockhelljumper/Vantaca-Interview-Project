"use client";

import {
  type ButtonHTMLAttributes,
  type FormEvent,
  type ReactNode,
  useCallback,
  useEffect,
  useId,
  useMemo,
  useState,
} from "react";
import { api, ApiError } from "@/lib/api";
import { formatMoney, formatTimestamp, transferTone } from "@/lib/format";

type Freshness = {
  state: "current" | "stale" | "degraded";
  fetched_at: string;
  checked_at: string;
  last_error: string | null;
  policy_label: string;
};

type Account = {
  id: string;
  display_name: string;
  type: string;
  last_four: string;
  balance: string;
  currency: string;
  status: "open" | "closed";
  version: number;
  freshness: Freshness;
};

type Transaction = {
  id: string;
  amount: string;
  currency: string;
  description: string;
  merchant_category_code: string | null;
  posted_at: string;
};

type Transfer = {
  id: string;
  request_id: string;
  from_account_id: string;
  to_account_id: string;
  from_display: string;
  to_display: string;
  amount: string;
  currency: string;
  status: "INTENT_RECORDED" | "PENDING" | "POSTED" | "FAILED" | "RETURNED" | "UNKNOWN";
  partner_transfer_id: string | null;
  error_category: string | null;
  message: string;
  created_at: string;
  updated_at: string;
};

type TransactionPayload = {
  account: Account;
  transactions: Transaction[];
  version: number;
  refresh_started: boolean;
  refreshing: boolean;
  invalidation: string;
};

type DemoInfo = {
  mode: string;
  read_model: string;
  unsigned_webhook_mode: boolean;
  production_blockers: string[];
  demo_assumptions: string[];
};

type Notice = {
  kind: "info" | "success" | "warning" | "error";
  text: string;
};

const readScenarios = [
  {
    value: "",
    label: "Normal Northwind response",
    description: "a successful bounded synchronization and current freshness state",
  },
  {
    value: "429",
    label: "429 — retry budget exhausted",
    description: "Retry-After handling, bounded safe-read retries, and stale-data preservation",
  },
  {
    value: "500",
    label: "500 — partner error",
    description: "transient partner failure classification without losing the SQL snapshot",
  },
  {
    value: "503",
    label: "503 — maintenance outage",
    description: "partner unavailability, degraded freshness, and last-known-good data",
  },
  {
    value: "latency",
    label: "Latency — client timeout",
    description: "client timeout enforcement and bounded recovery behavior",
  },
];

const transferScenarios = [
  {
    value: "",
    label: "Normal acceptance",
    description: "durable intent followed by exactly one accepted partner submission",
  },
  {
    value: "post-commit-timeout",
    label: "Post-commit timeout → UNKNOWN",
    description: "a partner commit with a lost response, durable UNKNOWN state, and no resubmission",
  },
  {
    value: "503",
    label: "503 → outcome treated as UNKNOWN",
    description: "conservative handling when a monetary 503 cannot prove that no transfer was created",
  },
  {
    value: "500",
    label: "500 → outcome treated as UNKNOWN",
    description: "conservative handling when a monetary 500 leaves the partner outcome uncertain",
  },
];

export function Dashboard() {
  const [info, setInfo] = useState<DemoInfo | null>(null);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [selectedAccountID, setSelectedAccountID] = useState("");
  const [transactionData, setTransactionData] = useState<TransactionPayload | null>(null);
  const [transfers, setTransfers] = useState<Transfer[]>([]);
  const [loading, setLoading] = useState(true);
  const [transactionsLoading, setTransactionsLoading] = useState(false);
  const [actionPending, setActionPending] = useState("");
  const [notice, setNotice] = useState<Notice | null>(null);
  const [readScenario, setReadScenario] = useState("");
  const [transferScenario, setTransferScenario] = useState("");
  const [fromAccountID, setFromAccountID] = useState("");
  const [toAccountID, setToAccountID] = useState("");
  const [amount, setAmount] = useState("250.00");

  const openAccounts = useMemo(
    () => accounts.filter((account) => account.status === "open"),
    [accounts],
  );
  const selectedAccount = useMemo(
    () => accounts.find((account) => account.id === selectedAccountID),
    [accounts, selectedAccountID],
  );

  const showError = useCallback((error: unknown) => {
    if (error instanceof ApiError) {
      const suffix = error.correlationID ? ` Reference ${error.correlationID}.` : "";
      setNotice({ kind: "error", text: `${error.message}${suffix}` });
      return;
    }
    setNotice({ kind: "error", text: "The demo action failed unexpectedly." });
  }, []);

  const loadAccounts = useCallback(async () => {
    const payload = await api<{ accounts: Account[] }>("/api/accounts");
    setAccounts(payload.accounts);
    setSelectedAccountID((current) => current || payload.accounts[0]?.id || "");
    setFromAccountID(
      (current) =>
        current || payload.accounts.find((item) => item.status === "open")?.id || "",
    );
    setToAccountID((current) => {
      if (current) return current;
      const open = payload.accounts.filter((item) => item.status === "open");
      return open[1]?.id ?? open[0]?.id ?? "";
    });
  }, []);

  const loadTransfers = useCallback(async () => {
    const payload = await api<{ transfers: Transfer[] }>("/api/transfers");
    setTransfers(payload.transfers);
  }, []);

  const pollForTransactionVersion = useCallback(
    async (accountID: string, initialVersion: number) => {
      for (let attempt = 0; attempt < 8; attempt += 1) {
        await new Promise((resolve) => window.setTimeout(resolve, 650));
        const next = await api<TransactionPayload>(
          `/api/accounts/${encodeURIComponent(accountID)}/transactions?refresh=false`,
        );
        setTransactionData(next);
        if (next.version > initialVersion) {
          await loadAccounts();
          setNotice({
            kind: "success",
            text:
              "Northwind differed from SQL. The database committed first, then the UI re-fetched version " +
              next.version +
              ".",
          });
          return;
        }
        if (!next.refreshing) {
          if (next.account.freshness.last_error) {
            setNotice({
              kind: "warning",
              text: "Northwind refresh failed. The last trustworthy SQL snapshot remains visible.",
            });
          } else {
            setNotice({
              kind: "info",
              text: "Northwind matched SQL. No frontend invalidation was necessary.",
            });
          }
          return;
        }
      }
      setNotice({
        kind: "warning",
        text: "The bounded refresh window ended. The SQL snapshot remains available while background work is reviewed.",
      });
    },
    [loadAccounts],
  );

  const loadTransactions = useCallback(
    async (accountID: string, refresh: boolean, scenario = "") => {
      if (!accountID) return;
      setTransactionsLoading(true);
      try {
        const query = new URLSearchParams({ refresh: String(refresh) });
        if (scenario) query.set("scenario", scenario);
        const payload = await api<TransactionPayload>(
          `/api/accounts/${encodeURIComponent(accountID)}/transactions?${query.toString()}`,
        );
        setTransactionData(payload);
        if (refresh && payload.refresh_started) {
          void pollForTransactionVersion(accountID, payload.version).catch(showError);
        }
      } catch (error) {
        showError(error);
      } finally {
        setTransactionsLoading(false);
      }
    },
    [pollForTransactionVersion, showError],
  );

  useEffect(() => {
    void (async () => {
      try {
        const [demoInfo] = await Promise.all([
          api<DemoInfo>("/api/demo/info"),
          loadAccounts(),
          loadTransfers(),
        ]);
        setInfo(demoInfo);
      } catch (error) {
        showError(error);
      } finally {
        setLoading(false);
      }
    })();
  }, [loadAccounts, loadTransfers, showError]);

  useEffect(() => {
    if (selectedAccountID) {
      void loadTransactions(selectedAccountID, true);
    }
  }, [selectedAccountID, loadTransactions]);

  async function simulateExternalActivity() {
    if (!selectedAccountID || selectedAccount?.status !== "open") {
      setNotice({
        kind: "warning",
        text: "External activity can only be simulated for an open account.",
      });
      return;
    }
    setActionPending("external");
    setNotice(null);
    try {
      const payload = await api<{ message: string }>(
        `/api/demo/accounts/${encodeURIComponent(selectedAccountID)}/external-activity`,
        { method: "POST", body: "{}" },
      );
      setNotice({ kind: "info", text: payload.message });
      await loadTransactions(selectedAccountID, true);
    } catch (error) {
      showError(error);
    } finally {
      setActionPending("");
    }
  }

  async function runRefreshLab() {
    setActionPending("sync");
    setNotice(null);
    try {
      await api("/api/demo/sync", {
        method: "POST",
        body: JSON.stringify({ scenario: readScenario }),
      });
      await Promise.all([loadAccounts(), loadTransactions(selectedAccountID, false)]);
      setNotice({
        kind: "success",
        text: readScenario
          ? "The selected scenario recovered inside the safe read policy."
          : "Bounded account and transaction synchronization completed.",
      });
    } catch (error) {
      showError(error);
      await Promise.all([loadAccounts(), loadTransactions(selectedAccountID, false)]);
    } finally {
      setActionPending("");
    }
  }

  async function submitTransfer(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setActionPending("transfer");
    setNotice(null);
    try {
      const requestID = `demo-${crypto.randomUUID()}`;
      const transfer = await api<Transfer>("/api/transfers", {
        method: "POST",
        body: JSON.stringify({
          request_id: requestID,
          from_account_id: fromAccountID,
          to_account_id: toAccountID,
          amount,
          currency: "USD",
          scenario: transferScenario,
        }),
      });
      await loadTransfers();
      setNotice({
        kind: transfer.status === "UNKNOWN" ? "warning" : "success",
        text: transfer.message,
      });
    } catch (error) {
      showError(error);
    } finally {
      setActionPending("");
    }
  }

  async function advanceTransfer(
    transfer: Transfer,
    status: "POSTED" | "FAILED" | "RETURNED",
  ) {
    setActionPending(`advance-${transfer.id}`);
    setNotice(null);
    try {
      const deliveries = status === "POSTED" ? 2 : 1;
      const updated = await api<Transfer>(
        `/api/demo/transfers/${encodeURIComponent(transfer.id)}/advance`,
        {
          method: "POST",
          body: JSON.stringify({ status, deliveries }),
        },
      );
      await loadTransfers();
      setNotice({
        kind: status === "RETURNED" || status === "FAILED" ? "warning" : "success",
        text:
          status === "POSTED"
            ? `${updated.message} Two identical webhooks were delivered to prove deduplication.`
            : updated.message,
      });
    } catch (error) {
      showError(error);
    } finally {
      setActionPending("");
    }
  }

  if (loading) {
    return (
      <main className="grid min-h-screen place-items-center px-6">
        <div className="rounded-3xl border border-[#dce5ea] bg-white px-10 py-9 text-center shadow-[0_22px_60px_rgba(21,60,79,0.12)]">
          <div className="mx-auto mb-5 size-11 animate-spin rounded-full border-4 border-[#d7e8ef] border-t-[#01679b]" />
          <p className="font-bold text-[#153c4f]">Loading the SQL-backed demo…</p>
          <p className="mt-1 text-xs text-[#6b7f8a]">Connecting Vantaca, SQL Server, and Northwind</p>
        </div>
      </main>
    );
  }

  return (
    <main className="mx-auto min-h-screen w-full min-w-0 max-w-[1480px] px-4 py-5 sm:px-7 lg:px-10 lg:py-8">
      <header className="mb-8 overflow-hidden rounded-[1.75rem] border border-[#dce5ea] bg-white shadow-[0_26px_80px_rgba(21,60,79,0.13)]">
        <div className="flex flex-wrap items-center justify-between gap-3 px-4 py-4 sm:gap-4 sm:px-9">
          <div className="flex min-w-0 items-center gap-5">
            <img
              src="https://www.vantaca.com/hubfs/Asset%204.svg"
              alt="Vantaca"
              width="180"
              height="65"
              className="h-9 w-auto shrink-0 sm:h-11"
            />
            <span className="hidden h-8 w-px bg-[#dce5ea] sm:block" aria-hidden="true" />
            <div className="hidden sm:block">
              <p className="text-[11px] font-extrabold uppercase tracking-[0.16em] text-[#01679b]">
                Northwind Connect
              </p>
              <p className="text-sm font-semibold text-[#536b78]">Integration architecture lab</p>
            </div>
          </div>
          <div className="flex min-w-0 items-center gap-2">
            <a
              href="http://localhost:18090"
              target="_blank"
              rel="noreferrer"
              className="hidden rounded-full border border-[#bcd3df] px-3 py-1.5 text-xs font-bold text-[#01679b] transition hover:border-[#01679b] hover:bg-[#eef7fb] md:inline-flex"
            >
              API explorer ↗
            </a>
            <span className="inline-flex items-center gap-2 rounded-full bg-[#eff7eb] px-3 py-1.5 text-xs font-bold text-[#396f2b]">
              <span className="size-2 rounded-full bg-[#63b249]" aria-hidden="true" />
              <span className="sm:hidden">Local</span>
              <span className="hidden sm:inline">Local environment</span>
            </span>
            <span className="hidden rounded-full bg-[#f3f3fa] px-3 py-1.5 text-xs font-semibold text-[#536b78] md:inline-flex">
              Synthetic data only
            </span>
          </div>
        </div>

        <div className="relative overflow-hidden bg-[#153c4f] px-6 py-9 text-white sm:px-9 lg:py-11">
          <div className="absolute -right-20 -top-32 size-80 rounded-full border-[52px] border-[#01679b]/35" aria-hidden="true" />
          <div className="absolute -bottom-24 right-[28%] size-52 rounded-full bg-[#63b249]/12 blur-2xl" aria-hidden="true" />
          <div className="relative grid gap-9 lg:grid-cols-[1.35fr_0.65fr] lg:items-end">
            <div>
              <div className="mb-5 flex flex-wrap items-center gap-3">
                <span className="rounded-full bg-[#63b249] px-3 py-1 text-[11px] font-extrabold uppercase tracking-[0.16em] text-white">
                  Interview demo
                </span>
                <span className="text-sm font-medium text-[#c9dbe4]">
                  Local containers · No real money
                </span>
              </div>
              <h1 className="max-w-4xl text-3xl font-extrabold leading-[1.08] tracking-[-0.035em] sm:text-5xl">
                Connected financial experiences,
                <span className="mt-1 block text-[#8fca78]">handled carefully.</span>
              </h1>
              <p className="mt-5 max-w-3xl text-[15px] font-medium leading-7 text-[#d8e5eb]">
                Northwind remains authoritative while Vantaca serves a timestamped SQL snapshot,
                reconciles external activity, and exposes ambiguous transfer outcomes instead of
                silently retrying money movement.
              </p>
            </div>
            <div className="grid gap-2.5 text-sm">
              {[
                ["Experience", "Next.js 16"],
                ["Application", "Go modular service"],
                ["Read model", info?.read_model ?? "SQL Server 2022"],
                ["Partner", "Deterministic Go mock"],
              ].map(([label, value]) => (
                <div
                  key={label}
                  className="grid min-w-0 grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)] items-center gap-3 rounded-xl border border-white/12 bg-white/[0.06] px-4 py-3 backdrop-blur-sm"
                >
                  <span className="text-[#b7ced9]">{label}</span>
                  <strong className="min-w-0 break-words text-right text-white">{value}</strong>
                </div>
              ))}
            </div>
          </div>
        </div>
      </header>

      <div aria-live="polite" aria-atomic="true">
        {notice ? <NoticeBanner notice={notice} onDismiss={() => setNotice(null)} /> : null}
      </div>

      <section aria-labelledby="accounts-heading" className="mb-7">
        <div className="mb-4 flex flex-wrap items-end justify-between gap-3">
          <div>
            <p className="text-[11px] font-extrabold uppercase tracking-[0.18em] text-[#01679b]">
              Workflow 01
            </p>
            <h2 id="accounts-heading" className="mt-1 text-2xl font-extrabold tracking-tight text-[#153c4f]">
              Linked account snapshots
            </h2>
          </div>
          <p className="max-w-xl text-sm font-medium leading-6 text-[#607784]">
            Balances are fetched snapshots—not “live” funds or transfer eligibility.
          </p>
        </div>

        <div className="grid gap-4 md:grid-cols-3">
          {accounts.map((account) => {
            const selected = account.id === selectedAccountID;
            return (
              <button
                key={account.id}
                type="button"
                data-testid={`account-${account.id}`}
                onClick={() => setSelectedAccountID(account.id)}
                className={`group rounded-2xl border p-5 text-left transition duration-200 ${
                  selected
                    ? "border-[#01679b] bg-[#01679b] text-white shadow-[0_16px_34px_rgba(1,103,155,0.2)]"
                    : "border-[#dce5ea] bg-white hover:-translate-y-0.5 hover:border-[#7cabc1] hover:shadow-[0_14px_30px_rgba(21,60,79,0.1)]"
                }`}
                aria-pressed={selected}
              >
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <p
                      className={`text-xs font-bold uppercase tracking-[0.16em] ${
                        selected ? "text-[#d6edf7]" : "text-[#708590]"
                      }`}
                    >
                      {account.status}
                    </p>
                    <h3 className="mt-1 text-lg font-extrabold">{account.display_name}</h3>
                  </div>
                  <FreshnessBadge state={account.freshness.state} selected={selected} />
                </div>
                <p className="mt-7 text-3xl font-black tracking-[-0.035em]">
                  {formatMoney(account.balance, account.currency)}
                </p>
                <p className={`mt-2 text-xs font-medium ${selected ? "text-[#d6edf7]/80" : "text-[#708590]"}`}>
                  Fetched {formatTimestamp(account.freshness.fetched_at)} · version {account.version}
                </p>
              </button>
            );
          })}
        </div>
      </section>

      <div className="grid min-w-0 gap-7 xl:grid-cols-[minmax(0,1.25fr)_minmax(0,0.75fr)]">
        <section
          aria-labelledby="transactions-heading"
          className="min-w-0 rounded-[1.75rem] border border-[#dce5ea] bg-white p-5 shadow-[0_12px_36px_rgba(21,60,79,0.07)] sm:p-7"
        >
          <div className="mb-5 flex flex-wrap items-start justify-between gap-4">
            <div>
              <p className="text-[11px] font-extrabold uppercase tracking-[0.18em] text-[#01679b]">
                Workflow 02
              </p>
              <h2 id="transactions-heading" className="mt-1 text-2xl font-extrabold tracking-tight text-[#153c4f]">
                Recent transactions
              </h2>
              <p className="mt-1 text-sm font-medium text-[#607784]">
                SQL responds first; a bounded worker compares Northwind in the background.
              </p>
            </div>
            <TestActionButton
              type="button"
              onClick={simulateExternalActivity}
              disabled={!selectedAccountID || selectedAccount?.status !== "open" || actionPending !== ""}
              tooltip={
                selectedAccount?.status === "closed"
                  ? "This test requires an open account. Northwind does not permit new external activity on a closed account."
                  : "Tests eventual consistency when activity happens outside Vantaca: SQL responds first, Northwind differs, the database commits, and the UI re-fetches only after that commit."
              }
              className="rounded-xl border border-[#b8d5a9] bg-[#eff7eb] px-4 py-2.5 text-sm font-bold text-[#396f2b] transition hover:border-[#63b249] hover:bg-[#e4f2dd] disabled:cursor-not-allowed disabled:opacity-50"
            >
              {actionPending === "external"
                ? "Changing Northwind…"
                : selectedAccount?.status === "closed"
                  ? "External activity unavailable"
                  : "Simulate external deposit"}
            </TestActionButton>
          </div>

          {transactionData ? (
            <>
              <div className="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-[#e5ebef] bg-[#f7f9fa] px-4 py-3 text-xs text-[#607784]">
                <span>
                  SQL version <strong className="text-slate-900">{transactionData.version}</strong>
                  {" · fetched "}
                  {formatTimestamp(transactionData.account.freshness.fetched_at)}
                </span>
                <span className="inline-flex items-center gap-2 font-bold text-[#01679b]">
                  <span className={`size-2 rounded-full ${transactionData.refreshing || transactionsLoading ? "animate-pulse bg-[#63b249]" : "bg-[#01679b]"}`} aria-hidden="true" />
                  {transactionData.refreshing || transactionsLoading
                    ? "Comparing Northwind…"
                    : "Snapshot ready"}
                </span>
              </div>
              <div className="overflow-x-auto">
                <table
                  aria-label="Recent transactions"
                  className="w-full min-w-[620px] border-separate border-spacing-0"
                >
                  <thead>
                    <tr className="text-left text-xs font-black uppercase tracking-[0.12em] text-slate-500">
                      <th className="border-b border-slate-200 px-3 py-3">Description</th>
                      <th className="border-b border-slate-200 px-3 py-3">Posted</th>
                      <th className="border-b border-slate-200 px-3 py-3">MCC</th>
                      <th className="border-b border-slate-200 px-3 py-3 text-right">Amount</th>
                    </tr>
                  </thead>
                  <tbody>
                    {transactionData.transactions.map((transaction) => (
                      <tr key={transaction.id} className="text-sm">
                        <td className="border-b border-slate-100 px-3 py-4 font-semibold text-slate-800">
                          {transaction.description}
                          <span className="mt-0.5 block font-mono text-[11px] font-normal text-slate-400">
                            {transaction.id}
                          </span>
                        </td>
                        <td className="border-b border-slate-100 px-3 py-4 text-slate-600">
                          {formatTimestamp(transaction.posted_at)}
                        </td>
                        <td className="border-b border-slate-100 px-3 py-4 text-slate-500">
                          {transaction.merchant_category_code ?? "—"}
                        </td>
                        <td
                          className={`border-b border-slate-100 px-3 py-4 text-right font-black ${
                            transaction.amount.startsWith("-")
                              ? "text-slate-900"
                              : "text-emerald-700"
                          }`}
                        >
                          {formatMoney(transaction.amount, transaction.currency)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          ) : (
            <p className="rounded-2xl bg-slate-50 p-6 text-sm text-slate-500">
              Select an account to load its SQL snapshot.
            </p>
          )}
        </section>

        <section
          aria-labelledby="failure-heading"
          className="min-w-0 rounded-[1.75rem] border border-[#cddfe8] bg-[#f5fafc] p-5 shadow-[0_12px_36px_rgba(21,60,79,0.06)] sm:p-7"
        >
          <p className="text-[11px] font-extrabold uppercase tracking-[0.18em] text-[#01679b]">Workflow 06</p>
          <h2 id="failure-heading" className="mt-1 text-2xl font-extrabold tracking-tight text-[#153c4f]">
            Read-path failure lab
          </h2>
          <p className="mt-2 text-sm font-medium leading-6 text-[#607784]">
            Safe reads have bounded retries. Failures preserve the last SQL snapshot and record
            degraded freshness.
          </p>
          <label htmlFor="read-scenario" className="mt-6 block text-sm font-bold text-[#153c4f]">
            Northwind behavior
          </label>
          <select
            id="read-scenario"
            value={readScenario}
            onChange={(event) => setReadScenario(event.target.value)}
            className="mt-2 w-full rounded-xl border border-[#bfd0d9] bg-white px-3 py-3 text-sm shadow-sm transition hover:border-[#7cabc1]"
          >
            {readScenarios.map((scenario) => (
              <option key={scenario.value} value={scenario.value}>
                {scenario.label}
              </option>
            ))}
          </select>
          <TestActionButton
            type="button"
            onClick={runRefreshLab}
            disabled={actionPending !== ""}
            tooltip={`Tests ${readScenarios.find((scenario) => scenario.value === readScenario)?.description ?? "the selected safe-read behavior"}. The application must preserve the last trustworthy SQL snapshot when Northwind is unavailable.`}
            wrapperClassName="mt-4 w-full"
            className="w-full rounded-xl bg-[#01679b] px-4 py-3 font-bold text-white shadow-[0_10px_20px_rgba(1,103,155,0.18)] transition hover:bg-[#005982] disabled:cursor-not-allowed disabled:opacity-50"
          >
            {actionPending === "sync" ? "Running bounded policy…" : "Run account + transaction sync"}
          </TestActionButton>
          <div className="mt-6 rounded-xl border border-[#dce5ea] bg-white p-4 text-sm leading-6 text-[#536b78]">
            <strong>Production question:</strong> Product still owns acceptable staleness and
            wording; Northwind still owes quotas and source-as-of semantics.
          </div>
        </section>
      </div>

      <section
        aria-labelledby="transfer-heading"
        className="mt-7 rounded-[1.75rem] border border-[#dce5ea] bg-white p-5 shadow-[0_12px_36px_rgba(21,60,79,0.07)] sm:p-7"
      >
        <div className="grid gap-7 xl:grid-cols-[0.7fr_1.3fr]">
          <div>
            <p className="text-[11px] font-extrabold uppercase tracking-[0.18em] text-[#01679b]">
              Workflows 03–05
            </p>
            <h2 id="transfer-heading" className="mt-1 text-2xl font-extrabold tracking-tight text-[#153c4f]">
              Guarded ACH transfer
            </h2>
            <p className="mt-2 text-sm font-medium leading-6 text-[#607784]">
              The demo persists a request identity, makes exactly one submission attempt, and
              never turns a timeout into “failed.”
            </p>

            <form onSubmit={submitTransfer} className="mt-6 space-y-4">
              <AccountSelect
                id="from-account"
                label="From"
                value={fromAccountID}
                accounts={openAccounts}
                onChange={setFromAccountID}
              />
              <AccountSelect
                id="to-account"
                label="To"
                value={toAccountID}
                accounts={openAccounts}
                onChange={setToAccountID}
              />
              <div>
                <label htmlFor="amount" className="block text-sm font-bold text-[#153c4f]">
                  Amount (USD)
                </label>
                <input
                  id="amount"
                  inputMode="decimal"
                  value={amount}
                  onChange={(event) => setAmount(event.target.value)}
                  className="mt-1.5 w-full rounded-xl border border-[#bfd0d9] bg-white px-3 py-3 text-sm shadow-sm transition hover:border-[#7cabc1]"
                  aria-describedby="amount-help"
                />
                <p id="amount-help" className="mt-1 text-xs font-medium text-[#708590]">
                  Exact two-decimal string; backend stores integer cents.
                </p>
              </div>
              <div>
                <label htmlFor="transfer-scenario" className="block text-sm font-bold text-[#153c4f]">
                  Submission behavior
                </label>
                <select
                  id="transfer-scenario"
                  value={transferScenario}
                  onChange={(event) => setTransferScenario(event.target.value)}
                  className="mt-1.5 w-full rounded-xl border border-[#bfd0d9] bg-white px-3 py-3 text-sm shadow-sm transition hover:border-[#7cabc1]"
                >
                  {transferScenarios.map((scenario) => (
                    <option key={scenario.value} value={scenario.value}>
                      {scenario.label}
                    </option>
                  ))}
                </select>
              </div>
              <TestActionButton
                type="submit"
                disabled={actionPending !== "" || !fromAccountID || !toAccountID}
                tooltip={`Tests ${transferScenarios.find((scenario) => scenario.value === transferScenario)?.description ?? "the selected transfer-submission behavior"}. Every path begins with a durable Vantaca request identity and permits only one partner POST.`}
                wrapperClassName="w-full"
                className="w-full rounded-xl bg-[#63b249] px-4 py-3.5 font-extrabold text-white shadow-[0_10px_22px_rgba(99,178,73,0.23)] transition hover:bg-[#559d3f] disabled:cursor-not-allowed disabled:opacity-50"
              >
                {actionPending === "transfer"
                  ? "Recording and submitting once…"
                  : "Submit demo transfer"}
              </TestActionButton>
            </form>
          </div>

          <div>
            <div className="mb-4 flex items-center justify-between gap-3">
              <h3 className="text-lg font-extrabold text-[#153c4f]">Transfer activity</h3>
              <span className="rounded-full bg-[#f3f3fa] px-3 py-1 text-xs font-bold text-[#536b78]">
                {transfers.length} requests
              </span>
            </div>
            <div className="space-y-3">
              {transfers.length === 0 ? (
                <div className="rounded-2xl border border-dashed border-[#bfd0d9] bg-[#f9fbfc] p-8 text-center text-sm font-medium text-[#708590]">
                  Submit a normal or ambiguous transfer to populate the lifecycle.
                </div>
              ) : (
                transfers.map((transfer) => (
                  <TransferCard
                    key={transfer.id}
                    transfer={transfer}
                    pending={actionPending === `advance-${transfer.id}`}
                    onAdvance={(status) => advanceTransfer(transfer, status)}
                  />
                ))
              )}
            </div>
          </div>
        </div>
      </section>

      <section className="mt-7 grid gap-4 lg:grid-cols-2">
        <article className="rounded-2xl border border-[#dce5ea] bg-white p-6 shadow-[0_10px_28px_rgba(21,60,79,0.05)]">
          <p className="text-[11px] font-extrabold uppercase tracking-[0.16em] text-[#01679b]">
            Demo assumptions
          </p>
          <ul className="mt-4 space-y-3 text-sm font-medium leading-6 text-[#536b78]">
            {info?.demo_assumptions.map((item) => (
              <li key={item} className="flex gap-3">
                <span className="text-[#63b249]">●</span>
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </article>
        <article className="rounded-2xl border border-[#efd1d1] bg-[#fff8f8] p-6 shadow-[0_10px_28px_rgba(21,60,79,0.05)]">
          <p className="text-[11px] font-extrabold uppercase tracking-[0.16em] text-[#a13e3e]">
            Still blocks production
          </p>
          <ul className="mt-4 space-y-3 text-sm font-medium leading-6 text-[#6f3b3b]">
            {info?.production_blockers.map((item) => (
              <li key={item} className="flex gap-3">
                <span className="text-rose-600">◆</span>
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </article>
      </section>

      <footer className="py-9 text-center text-xs font-medium text-[#708590]">
        <span className="font-bold text-[#153c4f]">Vantaca × Northwind integration lab.</span>{" "}
        Synthetic interview environment; passing this demo is not Northwind certification or production approval.
      </footer>
    </main>
  );
}

function TestActionButton({
  tooltip,
  wrapperClassName = "",
  className = "",
  children,
  ...buttonProps
}: ButtonHTMLAttributes<HTMLButtonElement> & {
  tooltip: string;
  wrapperClassName?: string;
  children: ReactNode;
}) {
  const tooltipID = useId();
  const disabled = Boolean(buttonProps.disabled);

  return (
    <span
      className={`group/tooltip relative inline-flex ${wrapperClassName}`}
      tabIndex={disabled ? 0 : undefined}
      aria-disabled={disabled || undefined}
      aria-describedby={disabled ? tooltipID : undefined}
    >
      <button
        {...buttonProps}
        aria-describedby={tooltipID}
        className={`inline-flex items-center justify-center gap-2 ${className}`}
      >
        <span>{children}</span>
        <span
          aria-hidden="true"
          className="grid size-5 shrink-0 place-items-center rounded-full border border-current/25 bg-white/15 text-[11px] font-extrabold leading-none"
        >
          i
        </span>
      </button>
      <span
        id={tooltipID}
        role="tooltip"
        className="pointer-events-none absolute bottom-[calc(100%+0.75rem)] left-0 z-[80] w-80 max-w-[calc(100vw-4.5rem)] rounded-xl bg-[#153c4f] px-4 py-3 text-left text-xs font-medium leading-5 text-white opacity-0 shadow-[0_14px_34px_rgba(21,60,79,0.3)] transition duration-150 group-hover/tooltip:opacity-100 group-focus-within/tooltip:opacity-100 sm:left-1/2 sm:max-w-[calc(100vw-2rem)] sm:-translate-x-1/2"
      >
        <span className="mb-1 block text-[10px] font-extrabold uppercase tracking-[0.16em] text-[#8fca78]">
          What this tests
        </span>
        {tooltip}
        <span
          aria-hidden="true"
          className="absolute left-6 top-full size-3 -translate-x-1/2 -translate-y-1/2 rotate-45 bg-[#153c4f] sm:left-1/2"
        />
      </span>
    </span>
  );
}

function AccountSelect({
  id,
  label,
  value,
  accounts,
  onChange,
}: {
  id: string;
  label: string;
  value: string;
  accounts: Account[];
  onChange: (value: string) => void;
}) {
  return (
    <div>
      <label htmlFor={id} className="block text-sm font-bold text-[#153c4f]">
        {label}
      </label>
      <select
        id={id}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="mt-1.5 w-full rounded-xl border border-[#bfd0d9] bg-white px-3 py-3 text-sm shadow-sm transition hover:border-[#7cabc1]"
      >
        {accounts.map((account) => (
          <option key={account.id} value={account.id}>
            {account.display_name}
          </option>
        ))}
      </select>
    </div>
  );
}

function FreshnessBadge({
  state,
  selected,
}: {
  state: Freshness["state"];
  selected: boolean;
}) {
  const label =
    state === "current"
      ? "Fresh snapshot"
      : state === "stale"
        ? "Stale snapshot"
        : "Refresh degraded";
  const classes = selected
    ? "bg-white/10 text-white"
    : state === "current"
      ? "bg-emerald-50 text-emerald-700"
      : state === "stale"
        ? "bg-amber-50 text-amber-800"
        : "bg-rose-50 text-rose-700";
  return (
    <span className={`rounded-full px-2.5 py-1 text-[11px] font-bold ${classes}`}>
      {label}
    </span>
  );
}

function NoticeBanner({
  notice,
  onDismiss,
}: {
  notice: Notice;
  onDismiss: () => void;
}) {
  const classes = {
    info: "border-sky-200 bg-sky-50 text-sky-950",
    success: "border-emerald-200 bg-emerald-50 text-emerald-950",
    warning: "border-amber-200 bg-amber-50 text-amber-950",
    error: "border-rose-200 bg-rose-50 text-rose-950",
  }[notice.kind];
  return (
    <div
      role={notice.kind === "error" ? "alert" : "status"}
      className={`mb-6 flex items-start justify-between gap-4 rounded-2xl border px-4 py-3 text-sm ${classes}`}
    >
      <p className="leading-6">{notice.text}</p>
      <button
        type="button"
        onClick={onDismiss}
        className="shrink-0 rounded-lg px-2 py-1 font-bold"
        aria-label="Dismiss message"
      >
        ×
      </button>
    </div>
  );
}

function TransferCard({
  transfer,
  pending,
  onAdvance,
}: {
  transfer: Transfer;
  pending: boolean;
  onAdvance: (status: "POSTED" | "FAILED" | "RETURNED") => void;
}) {
  const tone = transferTone(transfer.status);
  const toneClasses = {
    neutral: "bg-[#f3f3fa] text-[#536b78]",
    info: "bg-[#e8f4fa] text-[#01679b]",
    success: "bg-[#eff7eb] text-[#396f2b]",
    danger: "bg-rose-100 text-rose-800",
    warning: "bg-amber-100 text-amber-900",
  }[tone];

  return (
    <article
      data-testid="transfer-card"
      data-transfer-status={transfer.status}
      className="rounded-2xl border border-[#dce5ea] bg-white p-4 shadow-[0_6px_18px_rgba(21,60,79,0.04)]"
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="font-bold text-[#153c4f]">
            {transfer.from_display} → {transfer.to_display}
          </p>
          <p className="mt-1 text-xs font-medium text-[#708590]">
            {formatTimestamp(transfer.created_at)} · request{" "}
            {transfer.request_id.slice(0, 20)}…
          </p>
        </div>
        <div className="text-right">
          <p className="text-lg font-extrabold text-[#153c4f]">
            {formatMoney(transfer.amount, transfer.currency)}
          </p>
          <span
            className={`mt-1 inline-block rounded-full px-2.5 py-1 text-[11px] font-black ${toneClasses}`}
          >
            {transfer.status}
          </span>
        </div>
      </div>
      <p className="mt-3 rounded-xl bg-[#f7f9fa] px-3 py-2 text-xs font-medium leading-5 text-[#607784]">
        {transfer.message}
      </p>
      {transfer.error_category ? (
        <p className="mt-2 font-mono text-[11px] text-amber-800">
          safe error category: {transfer.error_category}
        </p>
      ) : null}
      <div className="mt-3 flex flex-wrap gap-2">
        {transfer.status === "PENDING" ? (
          <>
            <TestActionButton
              type="button"
              disabled={pending}
              onClick={() => onAdvance("POSTED")}
              tooltip="Tests the PENDING → POSTED transition and webhook idempotency. The mock sends the same POSTED webhook twice; the inbox stores one event and reconciliation confirms partner state."
              className="rounded-lg border border-[#b8d5a9] bg-[#eff7eb] px-3 py-2 text-xs font-bold text-[#396f2b] transition hover:bg-[#e4f2dd] disabled:opacity-50"
            >
              Post + duplicate webhook
            </TestActionButton>
            <TestActionButton
              type="button"
              disabled={pending}
              onClick={() => onAdvance("FAILED")}
              tooltip="Tests a definitive PENDING → FAILED transition and verifies that an unsuccessful transfer remains distinct from UNKNOWN, POSTED, and RETURNED."
              className="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-xs font-bold text-rose-800 transition hover:bg-rose-100 disabled:opacity-50"
            >
              Fail
            </TestActionButton>
          </>
        ) : null}
        {transfer.status === "POSTED" ? (
          <TestActionButton
            type="button"
            disabled={pending}
            onClick={() => onAdvance("RETURNED")}
            tooltip="Tests the late POSTED → RETURNED transition, proving that a posted ACH transfer is not treated as an irreversible final success."
            className="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-xs font-bold text-rose-800 transition hover:bg-rose-100 disabled:opacity-50"
          >
            Demonstrate late return
          </TestActionButton>
        ) : null}
      </div>
    </article>
  );
}
