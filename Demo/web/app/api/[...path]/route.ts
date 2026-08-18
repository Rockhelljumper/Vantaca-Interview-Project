import type { NextRequest } from "next/server";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{ path: string[] }>;
};

async function proxy(request: NextRequest, context: RouteContext) {
  const { path } = await context.params;
  const apiBase = process.env.API_INTERNAL_URL ?? "http://localhost:8080";
  const destination = new URL(`/api/${path.join("/")}`, apiBase);
  destination.search = request.nextUrl.search;

  const headers = new Headers();
  for (const name of ["content-type", "x-demo-tenant", "x-correlation-id"]) {
    const value = request.headers.get(name);
    if (value) {
      headers.set(name, value);
    }
  }

  const hasBody = request.method !== "GET" && request.method !== "HEAD";
  const response = await fetch(destination, {
    method: request.method,
    headers,
    body: hasBody ? await request.arrayBuffer() : undefined,
    cache: "no-store",
    signal: AbortSignal.timeout(30_000),
  });

  const responseHeaders = new Headers();
  responseHeaders.set(
    "content-type",
    response.headers.get("content-type") ?? "application/json",
  );
  responseHeaders.set("cache-control", "no-store");
  const correlationID = response.headers.get("x-correlation-id");
  if (correlationID) {
    responseHeaders.set("x-correlation-id", correlationID);
  }

  return new Response(response.body, {
    status: response.status,
    headers: responseHeaders,
  });
}

export const GET = proxy;
export const POST = proxy;
