const DEMO_TENANT = "tenant_demo";

export class ApiError extends Error {
  status: number;
  code: string;
  correlationID?: string;

  constructor(status: number, code: string, message: string, correlationID?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.correlationID = correlationID;
  }
}

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("X-Demo-Tenant", DEMO_TENANT);
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(path, {
    ...init,
    headers,
    cache: "no-store",
  });
  const payload = (await response.json()) as {
    error?: string;
    message?: string;
    correlation_id?: string;
  };
  if (!response.ok) {
    throw new ApiError(
      response.status,
      payload.error ?? "request_failed",
      payload.message ?? "The request failed.",
      payload.correlation_id,
    );
  }
  return payload as T;
}
