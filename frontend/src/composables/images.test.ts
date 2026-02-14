import type { ImageSummary } from "@/models/image";
import type { StaticInfo } from "@/models/static_info";
import { describe, expect, it } from "vitest";

import {
  base,
  compareSummaries,
  formatDate,
  formatGeonames,
  limitDisplayedImages,
  summaryKey,
} from "@/composables/images";

function makeSummary(key: string, modified: string): ImageSummary {
  return {
    bucket: "bucket-a",
    key,
    name: key,
    group: "g1",
    type: "t1",
    geonames: null,
    productInfo: null,
    dynamicFilters: {},
    cachedObject: { cacheKey: `${key}.png`, lastModified: new Date(modified) },
    size: { width: 10, height: 10 },
    _hasBeenUpdated: false,
    _lastModified: new Date(modified),
  } as ImageSummary;
}

describe("images composable helpers", () => {
  it("compareSummaries sorts by latest first", () => {
    const oldS = makeSummary("old", "2025-01-01T00:00:00.000Z");
    const newS = makeSummary("new", "2025-01-02T00:00:00.000Z");
    const items = [oldS, newS].sort(compareSummaries);
    expect(items.map((s) => s.key)).toEqual(["new", "old"]);
  });

  it("summaryKey combines bucket and key", () => {
    expect(summaryKey(makeSummary("img/a.png", "2025-01-01T00:00:00.000Z"))).toBe(
      "bucket-a_img/a.png"
    );
  });

  it("limitDisplayedImages uses static limit", () => {
    const items = [
      makeSummary("1", "2025-01-01T00:00:00.000Z"),
      makeSummary("2", "2025-01-02T00:00:00.000Z"),
      makeSummary("3", "2025-01-03T00:00:00.000Z"),
    ];
    const staticInfo = { maxImagesDisplayCount: 2 } as StaticInfo;
    expect(limitDisplayedImages(items, staticInfo).map((s) => s.key)).toEqual(["1", "2"]);
  });

  it("base trims path and strips prefix up to @", () => {
    expect(base("a/b/c@final.txt")).toBe("final.txt");
    expect(base("a/b/final.txt")).toBe("final.txt");
  });

  it("formatDate uses ISO-like output with space separator", () => {
    expect(formatDate(new Date("2025-01-01T12:34:56.789Z"))).toBe("2025-01-01 12:34:56.789Z");
  });

  it("formatGeonames handles null and nested location names", () => {
    expect(formatGeonames(null)).toBe("No geonames found");
    expect(
      formatGeonames({
        objects: [{ name: "Country", states: [{ name: "State", counties: [{ name: "County" }] }] }],
      } as never)
    ).toBe("Country / State / County");
  });
});
