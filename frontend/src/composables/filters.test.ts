import { describe, expect, it } from "vitest";

import { applyFilters } from "@/composables/filters";
import type { ImageSummary } from "@/models/image";

function makeSummary(params: {
  key: string;
  group: string;
  type: string;
  date: string;
  dyn: Record<string, string>;
}): ImageSummary {
  return {
    bucket: "bucket",
    key: params.key,
    name: params.key,
    group: params.group,
    type: params.type,
    geonames: null,
    productInfo: null,
    dynamicFilters: params.dyn,
    cachedObject: { cacheKey: `${params.key}.png`, lastModified: new Date(params.date) },
    size: { width: 1, height: 1 },
    _hasBeenUpdated: false,
    _lastModified: new Date(params.date),
  } as ImageSummary;
}

describe("applyFilters", () => {
  it("filters by dynamic filters, group/type selection and search", () => {
    const a = makeSummary({
      key: "alpha-key",
      group: "g1",
      type: "t1",
      date: "2025-01-01T00:00:00.000Z",
      dyn: { mode: "prod" },
    });
    const b = makeSummary({
      key: "beta-key",
      group: "g1",
      type: "t2",
      date: "2025-01-03T00:00:00.000Z",
      dyn: { mode: "dev" },
    });
    const c = makeSummary({
      key: "gamma-key",
      group: "g2",
      type: "t1",
      date: "2025-01-02T00:00:00.000Z",
      dyn: { mode: "prod" },
    });

    const out = applyFilters(
      [a, b, c],
      { mode: { prod: true, dev: false } },
      { g1: ["t1"], g2: ["t1"] },
      "alpha"
    );
    expect(out.map((x) => x.key)).toEqual(["alpha-key"]);
  });

  it("returns newest-first ordering after filtering", () => {
    const older = makeSummary({
      key: "older",
      group: "g",
      type: "t",
      date: "2025-01-01T00:00:00.000Z",
      dyn: { mode: "prod" },
    });
    const newer = makeSummary({
      key: "newer",
      group: "g",
      type: "t",
      date: "2025-01-04T00:00:00.000Z",
      dyn: { mode: "prod" },
    });

    const out = applyFilters([older, newer], { mode: { prod: true } }, { g: ["t"] }, "");
    expect(out.map((x) => x.key)).toEqual(["newer", "older"]);
  });
});
