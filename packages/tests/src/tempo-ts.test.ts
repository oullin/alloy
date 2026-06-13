import coreFixture from "@tempo/spec/fixtures/core.json" with { type: "json" };
import { Tempo, TempoImmutable } from "@tempo/tempo";
import { describe, expect, it } from "vitest";

describe("Tempo TypeScript shared Carbon fixtures", () => {
  it("uses pinned Carbon metadata", () => {
    expect(coreFixture.metadata).toMatchObject({
      source: "carbon",
      carbonVersion: "3.11.4",
      timezone: "UTC",
    });
  });

  it.each(coreFixture.cases)("$name", (item) => {
    const parsed = TempoImmutable.parse(item.input);

    expect(parsed.toISOString()).toBe(item.expectedIso);
    expect(parsed.toDateString()).toBe(item.expectedDate);

    if (item.expectedAddDaysIso !== undefined) {
      expect(parsed.addDays(item.addDays ?? 0).toISOString()).toBe(
        item.expectedAddDaysIso,
      );
    }
  });

  it("keeps mutable and immutable behavior distinct", () => {
    const mutable = Tempo.parse("2024-02-29T00:00:00+00:00");
    const immutable = TempoImmutable.parse("2024-02-29T00:00:00+00:00");

    expect(mutable.addDays(1)).toBe(mutable);
    expect(mutable.toISOString()).toBe("2024-03-01T00:00:00.000Z");

    const next = immutable.addDays(1);
    expect(next).not.toBe(immutable);
    expect(immutable.toISOString()).toBe("2024-02-29T00:00:00.000Z");
    expect(next.toISOString()).toBe("2024-03-01T00:00:00.000Z");
  });
});
