import { createPinia, setActivePinia } from "pinia";
import { describe, expect, it } from "vitest";

import { parseEventData } from "@/composables/events";
import { compareSummaries } from "@/composables/images";
import type { ImageSummary } from "@/models/image";
import { findSummaryIndex, useImageStore } from "@/stores/images";

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
    lastUpdateOverall: new Date("2025-01-01T00:00:00.000Z"),
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

describe("image update ordering", () => {
  it("uses the preview summary timestamp and advances it for live file updates", () => {
    setActivePinia(createPinia());
    const store = useImageStore();
    const first = makeSummary("bucket", "first");
    first.lastUpdateOverall = new Date("2025-01-02T00:00:00.000Z");
    store.populateSummaries([first]);

    const second = makeSummary("bucket", "second");
    second.lastUpdateOverall = new Date("2025-01-03T00:00:00.000Z");
    store.handleEvent(
      parseEventData(
        JSON.stringify({
          eventType: "ObjectCreated",
          objectType: "preview",
          imageBucket: second.bucket,
          imageKey: second.key,
          objectTime: second.cachedObject.lastModified,
          object: second,
        })
      ),
      undefined
    );

    const sortedKeys = () => [...store.allSummaries].sort(compareSummaries).map((s) => s.key);
    expect(sortedKeys()).toEqual(["second", "first"]);

    for (const [objectType, date] of [
      ["target", "2025-01-04T00:00:00.000Z"],
      ["dynamic_input", "2025-01-01T00:00:00.000Z"],
    ] as const) {
      store.handleEvent(
        {
          eventType: "ObjectCreated",
          objectType,
          imageBucket: first.bucket,
          imageKey: first.key,
          objectTime: new Date(date),
          error: undefined,
        },
        undefined
      );
      expect(sortedKeys()).toEqual(["first", "second"]);
    }
  });
});
