import { http, HttpResponse } from "msw";

const validationErrorExample = {
  status: "error",
  message: "validation failed",
};

const statusExample = {
  status: "ok",
};

const healthzExample = {
  status: "ok",
  checked_at: "2026-01-01T00:00:00Z",
  components: {
    app: { status: "up" },
    postgres: { status: "disabled" },
    redis: { status: "disabled" },
  },
};

const openAPIYAMLExample = "openapi: 3.0.3";
const swaggerHTMLExample = "<html><body>Swagger UI</body></html>";
const redocHTMLExample =
  "<html><body><redoc spec-url=\"/swagger/spec\"></redoc></body></html>";
const loginResponseExample = {
  token: "mock-jwt-token",
  expires_at: "2026-01-01T01:00:00Z",
};
const userProfileExample = {
  id: "user-123",
  email: "user@example.com",
  role: "user",
  created_at: "2026-01-01T00:00:00Z",
};

function isObjectBody(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function validationErrorResponse(): Response {
  return HttpResponse.json(validationErrorExample, { status: 422 });
}

async function validateOptionalBody(request: Request): Promise<Response | null> {
  const rawBody = await request.text();
  if (rawBody.trim().length === 0) {
    return null;
  }

  try {
    const parsed = JSON.parse(rawBody);
    if (isObjectBody(parsed)) {
      return null;
    }
  } catch {
    return validationErrorResponse();
  }

  return validationErrorResponse();
}

export const handlers = [
  http.get("/hc/status", async ({ request }) => {
    const validationError = await validateOptionalBody(request);
    if (validationError) {
      return validationError;
    }

    return HttpResponse.json(statusExample, { status: 200 });
  }),

  http.get("/healthz", async ({ request }) => {
    const validationError = await validateOptionalBody(request);
    if (validationError) {
      return validationError;
    }

    return HttpResponse.json(healthzExample, { status: 200 });
  }),

  http.get("/openapi.yaml", async ({ request }) => {
    const validationError = await validateOptionalBody(request);
    if (validationError) {
      return validationError;
    }

    return new HttpResponse(openAPIYAMLExample, {
      status: 200,
      headers: {
        "Content-Type": "application/yaml",
      },
    });
  }),

  http.get("/swagger/spec", async ({ request }) => {
    const validationError = await validateOptionalBody(request);
    if (validationError) {
      return validationError;
    }

    return new HttpResponse(openAPIYAMLExample, {
      status: 200,
      headers: {
        "Content-Type": "application/yaml",
      },
    });
  }),

  http.get("/swagger/ui", async ({ request }) => {
    const validationError = await validateOptionalBody(request);
    if (validationError) {
      return validationError;
    }

    return new HttpResponse(redocHTMLExample, {
      status: 200,
      headers: {
        "Content-Type": "text/html",
      },
    });
  }),

  http.get("/swagger/:path", async ({ request }) => {
    const validationError = await validateOptionalBody(request);
    if (validationError) {
      return validationError;
    }

    return new HttpResponse(swaggerHTMLExample, {
      status: 200,
      headers: {
        "Content-Type": "text/html",
      },
    });
  }),

  http.post("/auth/login", async () => {
    return HttpResponse.json(loginResponseExample, {
      status: 200,
      headers: {
        "Set-Cookie": "auth_token=mock-jwt-token; Path=/; HttpOnly; SameSite=Lax",
      },
    });
  }),

  http.post("/auth/logout", async () => {
    return HttpResponse.json(
      { message: "logged out" },
      {
        status: 200,
      },
    );
  }),

  http.get("/users/me", async () => {
    return HttpResponse.json(userProfileExample, {
      status: 200,
    });
  }),
];
