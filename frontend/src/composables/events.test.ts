import { describe, expect, it } from "vitest";

import { parseEventData } from "@/composables/events";

describe("parseEventData", () => {
  it("parses objectTime as Date", () => {
    const event = parseEventData(
      JSON.stringify({
        eventType: "ObjectCreated",
        objectType: "preview",
        imageBucket: "b",
        imageKey: "k",
        objectTime: "2025-01-01T00:00:00.000Z",
      })
    );

    expect(event.objectTime).toBeInstanceOf(Date);
    expect(event.objectTime.toISOString()).toBe("2025-01-01T00:00:00.000Z");
  });
});
