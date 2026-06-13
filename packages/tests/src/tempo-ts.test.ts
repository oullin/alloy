import {
  Tempo,
  TempoDuration,
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

  it("parses known token formats and applies offsets", () => {
    expect(
      Tempo.fromFormat(
        "2024/05/15 10:34:45.600 +0900",
        "YYYY/MM/DD HH:mm:ss.SSS ZZ",
      ).toISOString(),
    ).toBe("2024-05-15T01:34:45.600Z");

    expect(
      Tempo.fromFormat("05-15-24 10:34 PM", "MM-DD-YY hh:mm A").toISOString(),
    ).toBe("2024-05-15T22:34:00.000Z");

    expect(
      TempoFactory.create({ timeZone: "Asia/Tokyo" })
        .fromFormat("2024-01-01 09:00", "YYYY-MM-DD HH:mm")
        .toISOString(),
    ).toBe("2024-01-01T00:00:00.000Z");
  });

  it("exposes calendar predicates and humanized diffs", () => {
    const base = Tempo.parse("2024-02-29T00:00:00Z");
    const saturday = Tempo.parse("2024-03-02T00:00:00Z");

    expect(base.isLeapYear()).toBe(true);
    expect(base.daysInMonth()).toBe(29);
    expect(saturday.isWeekend()).toBe(true);
    expect(base.isWeekday()).toBe(true);
    expect(base.isToday("2024-02-29T12:00:00Z")).toBe(true);
    expect(base.isTomorrow("2024-02-28T12:00:00Z")).toBe(true);
    expect(base.isYesterday("2024-03-01T12:00:00Z")).toBe(true);
    expect(
      base.addDays(2).diffForHumans(base, {
        locale: "en-US",
        numeric: "always",
      }),
    ).toBe("in 2 days");
    expect(
      base.diffForHumans(base.addHours(3), {
        locale: "en-US",
        numeric: "always",
      }),
    ).toBe("3 hours ago");
  });

  it("supports ISO week metadata and weekday navigation", () => {
    const monday = Tempo.parse("2024-01-01T12:30:00Z");
    const week53 = Tempo.parse("2020-12-31T12:30:00Z");
    const friday = Tempo.parse("2024-05-17T12:30:00Z");

    expect(monday.isoWeekday).toBe(1);
    expect(monday.isoWeek).toBe(1);
    expect(monday.isoWeekYear).toBe(2024);
    expect(monday.weeksInISOYear).toBe(52);
    expect(week53.isoWeek).toBe(53);
    expect(week53.isoWeekYear).toBe(2020);
    expect(week53.weeksInISOYear).toBe(53);

    expect(monday.isMonday()).toBe(true);
    expect(monday.next("friday").toDateTimeString()).toBe(
      "2024-01-05 12:30:00",
    );
    expect(monday.previous("friday").toDateTimeString()).toBe(
      "2023-12-29 12:30:00",
    );
    expect(friday.nextWeekday().toDateString()).toBe("2024-05-20");
    expect(monday.previousWeekday().toDateString()).toBe("2023-12-29");
  });

  it("finds weekday positions inside months and computes full-year ages", () => {
    const may = Tempo.parse("2024-05-17T12:30:00Z");
    const birthday = Tempo.parse("2000-06-15T00:00:00Z");

    expect(may.firstOfMonth().toDateString()).toBe("2024-05-01");
    expect(may.firstOfMonth("monday").toDateString()).toBe("2024-05-06");
    expect(may.lastOfMonth().toDateString()).toBe("2024-05-31");
    expect(may.lastOfMonth("friday").toDateString()).toBe("2024-05-31");
    expect(birthday.age("2024-06-14T23:59:59Z")).toBe(23);
    expect(birthday.age("2024-06-15T00:00:00Z")).toBe(24);
  });

  it("parses, normalizes, serializes, and applies durations", () => {
    const parsed = TempoDuration.parse("P1Y2M3DT4H5M6.007S");
    const normalized = TempoDuration.parse("P1Y14M8DT25H61M61.250S");

    expect(parsed.toObject()).toMatchObject({
      days: 3,
      hours: 4,
      milliseconds: 7,
      minutes: 5,
      months: 2,
      seconds: 6,
      years: 1,
    });
    expect(parsed.toISOString()).toBe("P1Y2M3DT4H5M6.007S");
    expect(normalized.toObject()).toMatchObject({
      days: 9,
      hours: 2,
      milliseconds: 250,
      minutes: 2,
      months: 2,
      seconds: 1,
      years: 2,
    });
    expect(normalized.toISOString()).toBe("P2Y2M9DT2H2M1.250S");
    expect(TempoDuration.parse("PT0S").isZero()).toBe(true);
    expect(TempoDuration.parse("P2W").normalized().toISOString()).toBe("P14D");

    expect(
      Tempo.parse("2024-01-31T00:00:00Z").addDuration("P1M").toDateString(),
    ).toBe("2024-03-02");
    expect(
      Tempo.parse("2024-01-31T00:00:00Z")
        .subDuration(new TempoDuration({ days: 2, hours: 3 }))
        .toISOString(),
    ).toBe("2024-01-28T21:00:00.000Z");

    const interval = new TempoInterval(
      "2024-01-01T00:00:00Z",
      "2024-01-03T12:00:00Z",
    );
    expect(interval.toDuration().toISOString()).toBe("P2DT12H");

    const period = new TempoPeriod("2024-01-01", "2024-01-05", {
      step: "P2D",
    });
    expect(period.toArray().map((item) => item.toDateString())).toEqual([
      "2024-01-01",
      "2024-01-03",
      "2024-01-05",
    ]);
  });

  it("adds business days and compares same calendar units", () => {
    const friday = Tempo.parse("2024-05-17T10:00:00Z");
    const wednesday = Tempo.parse("2024-05-22T10:00:00Z");

    expect(friday.addWeekdays(1).toDateTimeString()).toBe(
      "2024-05-20 10:00:00",
    );
    expect(friday.addWeekdays(3).toDateTimeString()).toBe(
      "2024-05-22 10:00:00",
    );
    expect(wednesday.subWeekdays(3).toDateTimeString()).toBe(
      "2024-05-17 10:00:00",
    );
    expect(wednesday.diffInWeekdays(friday)).toBe(3);
    expect(wednesday.diffInWeekendDays(friday)).toBe(2);
    expect(friday.diffInWeekdays(wednesday)).toBe(-3);
    expect(friday.diffInWeekdays(wednesday, { absolute: true })).toBe(3);

    expect(friday.isSameSecond("2024-05-17T10:00:00.999Z")).toBe(true);
    expect(friday.isSameMinute("2024-05-17T10:00:59Z")).toBe(true);
    expect(friday.isSameHour("2024-05-17T10:59:59Z")).toBe(true);
    expect(friday.isSameDay("2024-05-17T23:59:59Z")).toBe(true);
    expect(friday.isSameWeek("2024-05-13T00:00:00Z")).toBe(true);
    expect(friday.isSameMonth("2024-05-01T00:00:00Z")).toBe(true);
    expect(friday.isSameQuarter("2024-04-01T00:00:00Z")).toBe(true);
    expect(friday.isSameYear("2024-12-31T23:59:59Z")).toBe(true);
    expect(friday.isBirthday("1990-05-17T00:00:00Z")).toBe(true);
    expect(friday.isBirthday("1990-05-18T00:00:00Z")).toBe(false);
  });

  it("exposes timezone names, offsets, UTC predicates, and DST state", () => {
    const utc = Tempo.parse("2024-01-01T00:00:00Z");
    const winter = Tempo.parse("2024-01-15T12:00:00Z", {
      timeZone: "America/New_York",
    });
    const summer = Tempo.parse("2024-07-15T12:00:00Z", {
      timeZone: "America/New_York",
    });

    expect(utc.isUtc()).toBe(true);
    expect(utc.offsetString()).toBe("+00:00");
    expect(utc.offsetString("")).toBe("+0000");
    expect(utc.timezoneName("shortOffset")).toBe("GMT+0");

    expect(winter.isUtc()).toBe(false);
    expect(winter.offsetMinutes).toBe(-300);
    expect(winter.offsetString()).toBe("-05:00");
    expect(winter.timezoneName("shortOffset")).toBe("GMT-5");
    expect(winter.isDST()).toBe(false);

    expect(summer.offsetMinutes).toBe(-240);
    expect(summer.offsetString()).toBe("-04:00");
    expect(summer.timezoneName("shortOffset")).toBe("GMT-4");
    expect(summer.isDST()).toBe(true);
  });
});
