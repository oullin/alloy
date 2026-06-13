import {
  Tempo,
  TempoFactory,
  TempoImmutable,
  TempoInterval,
  TempoMutable,
  TempoPeriod,
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

  it("creates and renders timezone-aware local components", () => {
    const tokyo = Tempo.create({
      day: 1,
      hour: 9,
      millisecond: 456,
      minute: 30,
      month: 1,
      second: 15,
      timeZone: "Asia/Tokyo",
      year: 2024,
    });

    expect(tokyo.toISOString()).toBe("2024-01-01T00:30:15.456Z");
    expect(tokyo.toDateTimeString()).toBe("2024-01-01 09:30:15");
    expect(tokyo.offsetMinutes).toBe(540);
    expect(tokyo.toArray()).toEqual([2024, 1, 1, 9, 30, 15, 456]);
    expect(tokyo.toObject()).toMatchObject({
      day: 1,
      hour: 9,
      millisecond: 456,
      minute: 30,
      month: 1,
      offsetMinutes: 540,
      second: 15,
      timeZone: "Asia/Tokyo",
      weekday: 1,
      year: 2024,
    });
  });

  it("converts timezones with and without preserving local time", () => {
    const utc = Tempo.parse("2024-01-01T12:00:00Z");
    const tokyo = utc.setTimezone("Asia/Tokyo");
    const preserved = utc.setTimezone("Asia/Tokyo", true);

    expect(tokyo.toDateTimeString()).toBe("2024-01-01 21:00:00");
    expect(tokyo.toISOString()).toBe("2024-01-01T12:00:00.000Z");
    expect(preserved.toDateTimeString()).toBe("2024-01-01 12:00:00");
    expect(preserved.toISOString()).toBe("2024-01-01T03:00:00.000Z");
  });

  it("supports setters, unit arithmetic, boundaries, and overflow modes", () => {
    const base = Tempo.parse("2024-01-31T10:20:30.400Z");

    expect(
      base
        .set({ day: 15, hour: 1, minute: 2, second: 3, millisecond: 4 })
        .toISOString(),
    ).toBe("2024-01-15T01:02:03.004Z");
    expect(base.addMonths(1).toDateString()).toBe("2024-03-02");
    expect(base.addMonthsNoOverflow(1).toDateString()).toBe("2024-02-29");
    expect(
      Tempo.parse("2024-02-29T00:00:00Z").addYearsNoOverflow(1).toDateString(),
    ).toBe("2025-02-28");
    expect(base.startOf("day").toISOString()).toBe("2024-01-31T00:00:00.000Z");
    expect(base.endOf("month").toISOString()).toBe("2024-01-31T23:59:59.999Z");
    expect(base.startOf("quarter").toISOString()).toBe(
      "2024-01-01T00:00:00.000Z",
    );
    expect(base.endOf("year").toISOString()).toBe("2024-12-31T23:59:59.999Z");
  });

  it("compares, diffs, rounds, and formats real instants", () => {
    const base = Tempo.parse("2024-05-15T10:34:45.600Z");
    const earlier = Tempo.parse("2024-05-14T08:00:00Z");

    expect(base.isAfter(earlier)).toBe(true);
    expect(base.isSame("2024-05-15T23:59:00Z", "day")).toBe(true);
    expect(base.isBetween(earlier, "2024-05-16T00:00:00Z")).toBe(true);
    expect(base.diffInHours(earlier)).toBe(26);
    expect(base.diffInMinutes(earlier)).toBe(1594);
    expect(base.floor("hour").toISOString()).toBe("2024-05-15T10:00:00.000Z");
    expect(base.ceil("hour").toISOString()).toBe("2024-05-15T11:00:00.000Z");
    expect(base.round("hour").toISOString()).toBe("2024-05-15T11:00:00.000Z");
    expect(base.format("YYYY-MM-DD HH:mm:ss.SSS ZZ [Q]M")).toBe(
      "2024-05-15 10:34:45.600 +0000 Q5",
    );
    expect(base.format("dddd, MMMM Do YYYY", { locale: "en-US" })).toBe(
      "Wednesday, May 15th 2024",
    );
  });

  it("iterates periods and computes intervals", () => {
    const interval = new TempoInterval(
      "2024-01-01T00:00:00Z",
      "2024-01-03T12:00:00Z",
    );

    expect(interval.days).toBe(2);
    expect(interval.hours).toBe(60);
    expect(interval.contains("2024-01-03T12:00:00Z")).toBe(true);
    expect(interval.contains("2024-01-03T12:00:00Z", "[)")).toBe(false);

    const period = new TempoPeriod("2024-01-01", "2024-01-05", {
      step: { days: 2 },
    });

    expect(period.toArray().map((item) => item.toDateString())).toEqual([
      "2024-01-01",
      "2024-01-03",
      "2024-01-05",
    ]);
  });

  it("mutates mutable instances for the expanded API", () => {
    const mutable = TempoMutable.parse("2024-01-01T00:00:00Z");

    expect(mutable.addHours(5).startOf("day")).toBe(mutable);
    expect(mutable.toISOString()).toBe("2024-01-01T00:00:00.000Z");
    expect(mutable.setTimezone("Asia/Tokyo")).toBe(mutable);
    expect(mutable.toDateTimeString()).toBe("2024-01-01 09:00:00");
  });
});
