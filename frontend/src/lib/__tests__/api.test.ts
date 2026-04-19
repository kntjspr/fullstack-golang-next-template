import { handlers } from "@/mocks/handlers";

describe("API mocks", () => {
  it("registers required backend routes", () => {
    const routeHeaders = handlers.map((handler) => handler.info.header).sort();

    expect(routeHeaders).toEqual([
      "GET /hc/status",
      "GET /healthz",
      "GET /openapi.yaml",
      "GET /swagger/:path",
      "GET /swagger/spec",
      "GET /swagger/ui",
      "GET /users/me",
      "POST /auth/login",
      "POST /auth/logout",
      "POST /auth/refresh",
    ]);
  });

  it("registers users/me mock route", () => {
    const routeHeaders = handlers.map((handler) => handler.info.header);
    expect(routeHeaders).toContain("GET /users/me");
  });

  it("uses only GET and POST methods for API handlers", () => {
    const methods = handlers.map((handler) => handler.info.method);
    const uniqueMethods = Array.from(new Set(methods)).sort();

    expect(uniqueMethods).toEqual(["GET", "POST"]);
  });
});
