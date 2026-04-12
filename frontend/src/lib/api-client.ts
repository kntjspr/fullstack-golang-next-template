const TOKEN_STORAGE_KEY = "token";

function getToken(): string | null {
  if (typeof globalThis.localStorage === "undefined") {
    return null;
  }

  return globalThis.localStorage.getItem(TOKEN_STORAGE_KEY);
}

function clearToken(): void {
  if (typeof globalThis.localStorage === "undefined") {
    return;
  }

  globalThis.localStorage.removeItem(TOKEN_STORAGE_KEY);
}

export class ApiError extends Error {
  status: number;
  requestId?: string;

  constructor(status: number, message: string, requestId?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.requestId = requestId;
  }
}

function buildRequestURL(path: string): string {
  if (path.startsWith("http://") || path.startsWith("https://")) {
    return path;
  }

  const baseURL = process.env.NEXT_PUBLIC_API_URL ?? "";
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  if (!baseURL) {
    return normalizedPath;
  }

  return new URL(normalizedPath, baseURL).toString();
}

async function parseErrorMessage(response: Response): Promise<string> {
  const defaultMessage = `Request failed with status ${response.status}`;
  const contentType = response.headers.get("content-type") ?? "";

  if (!contentType.includes("application/json")) {
    return defaultMessage;
  }

  try {
    const body = (await response.json()) as { message?: unknown };
    if (typeof body.message === "string" && body.message.length > 0) {
      return body.message;
    }
  } catch {
    return defaultMessage;
  }

  return defaultMessage;
}

async function parseSuccessBody<T>(response: Response): Promise<T> {
  if (response.status === 204) {
    return undefined as T;
  }

  const contentType = response.headers.get("content-type") ?? "";
  if (contentType.includes("application/json")) {
    return (await response.json()) as T;
  }

  return (await response.text()) as T;
}

export class ApiClient {
  async get<T>(path: string): Promise<T> {
    return this.request<T>("GET", path);
  }

  async post<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>("POST", path, body);
  }

  async put<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>("PUT", path, body);
  }

  async del<T>(path: string): Promise<T> {
    return this.request<T>("DELETE", path);
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const headers: Record<string, string> = {};
    const token = getToken();

    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    if (body !== undefined) {
      headers["Content-Type"] = "application/json";
    }

    let response: Response;
    try {
      response = await fetch(buildRequestURL(path), {
        method,
        credentials: "include",
        headers,
        body: body === undefined ? undefined : JSON.stringify(body),
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : "Unknown fetch error";
      throw new ApiError(0, `Network error: ${message}`);
    }

    const requestId = response.headers.get("X-Request-ID") ?? undefined;

    if (response.status === 401) {
      clearToken();
    }

    if (!response.ok) {
      const message = await parseErrorMessage(response);
      throw new ApiError(response.status, message, requestId);
    }

    return parseSuccessBody<T>(response);
  }
}

export const apiClient = new ApiClient();
