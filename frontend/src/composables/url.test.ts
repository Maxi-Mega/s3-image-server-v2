import { describe, expect, it } from "vitest";

import { resolveBackendURL } from "@/composables/url";

describe("resolveBackendURL", () => {
  it("joins base URL and path", () => {
    expect(resolveBackendURL("/api/info")).toMatch(/\/api\/info$/);
  });
});
