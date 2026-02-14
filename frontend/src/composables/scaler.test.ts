import { describe, expect, it, vi } from "vitest";

import { Scaler } from "@/composables/scaler";

describe("Scaler", () => {
  it("initializes with computed value and emits updates", () => {
    const input = document.createElement("input");
    input.type = "range";
    input.min = "10";
    input.max = "30";

    const onUpdate = vi.fn();
    const scaler = new Scaler(input, 50, 16);
    scaler.onUpdateScale = onUpdate;

    input.value = "24";
    input.dispatchEvent(new Event("input"));

    expect(scaler.currentValue()).toBe(24);
    expect(onUpdate).toHaveBeenCalledWith("13px", 24);

    scaler.dispose();
  });

  it("resets value on auxclick", () => {
    const input = document.createElement("input");
    input.type = "range";
    input.min = "0";
    input.max = "20";

    const scaler = new Scaler(input, 25, 16);
    input.value = "20";
    input.dispatchEvent(new Event("auxclick"));

    expect(scaler.currentValue()).toBe(5);

    scaler.dispose();
  });
});
