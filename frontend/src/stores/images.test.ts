import { describe, expect, it } from "vitest";

import type { ImageSummary } from "@/models/image";
import { findSummaryIndex } from "@/stores/images";

function makeSummary(bucket: string, key: string): ImageSummary {
  return {
    bucket,
    key,
    name: key,
    group: "g",
    type: "t",
    geonames: null,
    productInfo: null,
    dynamicFilters: {},
    cachedObject: { cacheKey: `${key}.png`, lastModified: new Date("2025-01-01T00:00:00.000Z") },
    size: { width: 1, height: 1 },
    _hasBeenUpdated: false,
    _lastModified: new Date("2025-01-01T00:00:00.000Z"),
  } as ImageSummary;
}

describe("findSummaryIndex", () => {
  it("returns index for matching bucket/key", () => {
    const arr = [makeSummary("a", "1"), makeSummary("b", "2"), makeSummary("c", "3")];
    expect(findSummaryIndex("b", "2", arr)).toBe(1);
  });

  it("returns -1 when there is no match", () => {
    const arr = [makeSummary("a", "1")];
    expect(findSummaryIndex("missing", "nope", arr)).toBe(-1);
  });
});
