import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";

import { ApiClient, ApiError } from "@/lib/api-client";
import server from "@/mocks/server";

class LocalStorageMock implements Storage {
  private store = new Map<string, string>();

  get length(): number {
    return this.store.size;
  }

  clear(): void {
    this.store.clear();
  }

  getItem(key: string): string | null {
    return this.store.get(key) ?? null;
  }

  key(index: number): string | null {
    return Array.from(this.store.keys())[index] ?? null;
  }

  removeItem(key: string): void {
    this.store.delete(key);
  }

  setItem(key: string, value: string): void {
    this.store.set(key, value);
  }
}

describe("ApiClient", () => {
  beforeAll(() => {
    Object.defineProperty(globalThis, "localStorage", {
      value: new LocalStorageMock(),
      configurable: true,
      writable: true,
    });

    server.listen({ onUnhandledRequest: "error" });
  });

  afterEach(() => {
    server.resetHandlers();
    localStorage.clear();
  });

  afterAll(() => {
    server.close();
  });

  beforeEach(() => {
    process.env.NEXT_PUBLIC_API_URL = "http://localhost:3000";
  });

  it("TestGet_Success", async () => {
    const client = new ApiClient();

    server.use(
      http.get("http://localhost:3000/healthz", () =>
        HttpResponse.json(
          {
            status: "ok",
          },
          { status: 200 },
        ),
      ),
    );

    const result = await client.get<{ status: string }>("/healthz");

    expect(result).toEqual({ status: "ok" });
  });

  it("TestGet_401_ClearsToken", async () => {
    const client = new ApiClient();
    localStorage.setItem("token", "example-token");

    server.use(
      http.get("http://localhost:3000/healthz", () =>
        HttpResponse.json(
          { status: "error", message: "unauthorized" },
          {
            status: 401,
          },
        ),
      ),
    );

    await expect(client.get("/healthz")).rejects.toBeInstanceOf(ApiError);
    expect(localStorage.getItem("token")).toBeNull();
  });

  it("TestGet_NetworkError", async () => {
    const client = new ApiClient();

    server.use(
      http.get("http://localhost:3000/healthz", () => {
        return HttpResponse.error();
      }),
    );

    await expect(client.get("/healthz")).rejects.toMatchObject({
      message: expect.stringMatching(/network|fetch/i),
    });
  });

  it("TestGet_RequestIdPropagated", async () => {
    const client = new ApiClient();

    server.use(
      http.get("http://localhost:3000/healthz", () =>
        HttpResponse.json(
          { status: "error", message: "internal server error" },
          {
            status: 500,
            headers: {
              "X-Request-ID": "req-123",
            },
          },
        ),
      ),
    );

    await expect(client.get("/healthz")).rejects.toMatchObject({
      status: 500,
      requestId: "req-123",
    });
  });

  it("TestGet_UsesErrorFieldFromBackend", async () => {
    const client = new ApiClient();

    server.use(
      http.get("http://localhost:3000/healthz", () =>
        HttpResponse.json(
          { error: "invalid token" },
          {
            status: 401,
          },
        ),
      ),
    );

    await expect(client.get("/healthz")).rejects.toMatchObject({
      status: 401,
      message: "invalid token",
    });
  });

  it("TestPost_SendsJSON", async () => {
    const client = new ApiClient();
    const payload = { email: "hello@example.com" };
    let capturedContentType: string | null = null;
    let capturedBody: unknown = null;

    server.use(
      http.post("http://localhost:3000/echo", async ({ request }) => {
        capturedContentType = request.headers.get("content-type");
        capturedBody = await request.json();
        return HttpResponse.json({ ok: true });
      }),
    );

    await client.post("/echo", payload);

    expect(capturedContentType).toContain("application/json");
    expect(capturedBody).toEqual(payload);
  });
});
