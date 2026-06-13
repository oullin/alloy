import {
  Tempo,
  TempoFactory,
  TempoImmutable,
  TempoMutable,
} from "@tempo/tempo";
import { describe, expect, it } from "vitest";

const dateCases = [
  {
    name: "parse utc start of day",
    input: "2024-02-29T00:00:00+00:00",
    expectedIso: "2024-02-29T00:00:00.000Z",
    expectedDate: "2024-02-29",
    addDays: 1,
    expectedAddDaysIso: "2024-03-01T00:00:00.000Z",
  },
  {
    name: "parse utc end of year",
    input: "2024-12-31T23:30:00+00:00",
    expectedIso: "2024-12-31T23:30:00.000Z",
    expectedDate: "2024-12-31",
    addDays: 1,
    expectedAddDaysIso: "2025-01-01T23:30:00.000Z",
  },
] as const;

describe("Tempo TypeScript behavior", () => {
  it.each(dateCases)("$name", (item) => {
    const parsed = TempoImmutable.parse(item.input);

    expect(parsed.toISOString()).toBe(item.expectedIso);
    expect(parsed.toDateString()).toBe(item.expectedDate);

    if (item.expectedAddDaysIso !== undefined) {
      expect(parsed.addDays(item.addDays ?? 0).toISOString()).toBe(
        item.expectedAddDaysIso,
      );
    }
  });

  it("uses immutable Tempo by default and exposes explicit mutable behavior", () => {
    const tempo = Tempo.parse("2024-02-29T00:00:00+00:00");
    const mutable = TempoMutable.parse("2024-02-29T00:00:00+00:00");
    const immutable = TempoImmutable.parse("2024-02-29T00:00:00+00:00");

    const tempoNext = tempo.addDays(1);
    expect(tempoNext).not.toBe(tempo);
    expect(tempo.toISOString()).toBe("2024-02-29T00:00:00.000Z");
    expect(tempoNext.toISOString()).toBe("2024-03-01T00:00:00.000Z");

    expect(mutable.addDays(1)).toBe(mutable);
    expect(mutable.toISOString()).toBe("2024-03-01T00:00:00.000Z");

    const next = immutable.addDays(1);
    expect(next).not.toBe(immutable);
    expect(immutable.toISOString()).toBe("2024-02-29T00:00:00.000Z");
    expect(next.toISOString()).toBe("2024-03-01T00:00:00.000Z");
  });

  it("accepts both second and millisecond numeric timestamps", () => {
    expect(TempoImmutable.fromTimestamp(1_704_067_200).toISOString()).toBe(
      "2024-01-01T00:00:00.000Z",
    );

    expect(TempoImmutable.parse(1_704_067_200_000).toISOString()).toBe(
      "2024-01-01T00:00:00.000Z",
    );
  });

  it("scopes test-now behavior to factories", () => {
    const fixed = TempoFactory.withTestNow("2025-01-01T00:00:00+00:00");

    expect(fixed.now().toISOString()).toBe("2025-01-01T00:00:00.000Z");
    expect(fixed.immutableNow().toISOString()).toBe("2025-01-01T00:00:00.000Z");
    expect(fixed.mutableNow().toISOString()).toBe("2025-01-01T00:00:00.000Z");
  });
});
