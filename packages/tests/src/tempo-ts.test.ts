import {
  Tempo,
  TempoDuration,
  TempoFactory,
  TempoImmutable,
  TempoInterval,
  TempoMutable,
  TempoPeriod,
  average,
  canParse,
  hasFormat,
  max,
  min,
  today,
  tomorrow,
  tryFromFormat,
  tryParse,
  yesterday,
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
    expect(fixed.today().toISOString()).toBe("2025-01-01T00:00:00.000Z");
    expect(fixed.tomorrow().toISOString()).toBe("2025-01-02T00:00:00.000Z");
    expect(fixed.yesterday().toISOString()).toBe("2024-12-31T00:00:00.000Z");
    expect(fixed.immutableNow().toISOString()).toBe("2025-01-01T00:00:00.000Z");
    expect(fixed.mutableNow().toISOString()).toBe("2025-01-01T00:00:00.000Z");
    expect(Tempo.today({ timeZone: "UTC" }).hour).toBe(0);
    expect(Tempo.tomorrow({ timeZone: "UTC" }).diffInDays(Tempo.today())).toBe(
      1,
    );
    expect(Tempo.yesterday({ timeZone: "UTC" }).diffInDays(Tempo.today())).toBe(
      -1,
    );
    expect(today({ timeZone: "UTC" }).hour).toBe(0);
    expect(tomorrow({ timeZone: "UTC" }).diffInDays(today())).toBe(1);
    expect(yesterday({ timeZone: "UTC" }).diffInDays(today())).toBe(-1);

    const tokyo = TempoFactory.create({ timeZone: "Asia/Tokyo" });

    expect(tokyo.tryParse("2025-01-01 09:00")?.toISOString()).toBe(
      "2025-01-01T00:00:00.000Z",
    );
    expect(tokyo.canParse("not a date")).toBe(false);
    expect(tokyo.tryParse("not a date")).toBeNull();
    expect(
      tokyo
        .tryFromFormat("2025-01-01 09:30", "YYYY-MM-DD HH:mm")
        ?.toISOString(),
    ).toBe("2025-01-01T00:30:00.000Z");
    expect(tokyo.hasFormat("2025-01-01 09:30", "YYYY-MM-DD HH:mm")).toBe(true);
    expect(tokyo.hasFormat("2025/01/01", "YYYY-MM-DD")).toBe(false);
    expect(
      tokyo.fromObject({ day: 1, hour: 9, month: 1, year: 2025 }).toISOString(),
    ).toBe("2025-01-01T00:00:00.000Z");
    expect(tokyo.fromTimestampMs(1_735_689_600_000).toDateTimeString()).toBe(
      "2025-01-01 09:00:00",
    );
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
    expect(base.startOfWeek().toDateString()).toBe("2024-01-29");
    expect(base.endOfWeek({ weekStartsOn: 0 }).toDateString()).toBe(
      "2024-02-03",
    );
    expect(base.endOf("month").toISOString()).toBe("2024-01-31T23:59:59.999Z");
    expect(base.startOf("quarter").toISOString()).toBe(
      "2024-01-01T00:00:00.000Z",
    );
    expect(base.startOfQuarter().toISOString()).toBe(
      "2024-01-01T00:00:00.000Z",
    );
    expect(base.endOfQuarter().toISOString()).toBe("2024-03-31T23:59:59.999Z");
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

    const calendarInterval = new TempoInterval(
      "2023-01-01T00:00:00Z",
      "2024-03-01T00:00:00Z",
    );
    expect(calendarInterval.weeks).toBe(60);
    expect(calendarInterval.months).toBe(14);
    expect(calendarInterval.years).toBe(1);

    const period = new TempoPeriod("2024-01-01", "2024-01-05", {
      step: { days: 2 },
    });

    expect(period.toArray().map((item) => item.toDateString())).toEqual([
      "2024-01-01",
      "2024-01-03",
      "2024-01-05",
    ]);
    expect(period.count()).toBe(3);
    expect(period.first()?.toDateString()).toBe("2024-01-01");
    expect(period.last()?.toDateString()).toBe("2024-01-05");
    expect(period.contains("2024-01-04")).toBe(true);
    expect(period.contains("2024-01-06")).toBe(false);
    expect(period.isEmpty()).toBe(false);

    const openPeriod = new TempoPeriod("2024-01-01", "2024-01-05", {
      includeEnd: false,
      step: { days: 2 },
    });
    expect(openPeriod.toArray().map((item) => item.toDateString())).toEqual([
      "2024-01-01",
      "2024-01-03",
    ]);
    expect(openPeriod.contains("2024-01-05")).toBe(false);

    const reversePeriod = new TempoPeriod("2024-01-05", "2024-01-01", {
      step: { days: -2 },
    });
    expect(reversePeriod.toArray().map((item) => item.toDateString())).toEqual([
      "2024-01-05",
      "2024-01-03",
      "2024-01-01",
    ]);
    expect(() =>
      new TempoPeriod("2024-01-05", "2024-01-01", {
        step: { days: 1 },
      }).toArray(),
    ).toThrow("advance toward the end");
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
      Tempo.fromFormat(
        "Wednesday, May 15th 2024 10:34 PM",
        "dddd, MMMM Do YYYY hh:mm A",
      ).toISOString(),
    ).toBe("2024-05-15T22:34:00.000Z");

    expect(
      Tempo.fromFormat("Wed, May 15 2024", "ddd, MMM D YYYY").toDateString(),
    ).toBe("2024-05-15");

    expect(
      TempoFactory.create({ timeZone: "Asia/Tokyo" })
        .fromFormat("2024-01-01 09:00", "YYYY-MM-DD HH:mm")
        .toISOString(),
    ).toBe("2024-01-01T00:00:00.000Z");

    expect(Tempo.canParse("2024-05-15T10:34:45Z")).toBe(true);
    expect(Tempo.canParse("not a date")).toBe(false);
    expect(canParse("2024-05-15")).toBe(true);
    expect(tryParse("not a date")).toBeNull();
    expect(TempoImmutable.tryParse("2024-05-15")).toBeInstanceOf(
      TempoImmutable,
    );
    expect(Tempo.tryParse("2024-05-15")).toBeInstanceOf(Tempo);
    expect(TempoMutable.tryParse("2024-05-15")).toBeInstanceOf(TempoMutable);

    expect(Tempo.hasFormat("2024/05/15 10:34", "YYYY/MM/DD HH:mm")).toBe(true);
    expect(Tempo.hasFormat("2024-05-15", "YYYY/MM/DD")).toBe(false);
    expect(hasFormat("2024-05-15", "YYYY-MM-DD")).toBe(true);
    expect(
      tryFromFormat("2024/05/15 10:34", "YYYY/MM/DD HH:mm")?.toISOString(),
    ).toBe("2024-05-15T10:34:00.000Z");
    expect(Tempo.tryFromFormat("2024-05-15", "YYYY/MM/DD")).toBeNull();
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
    expect(may.nthOfMonth(3, "monday")?.toDateString()).toBe("2024-05-20");
    expect(may.nthOfMonth(-1, "monday")?.toDateString()).toBe("2024-05-27");
    expect(may.nthOfMonth(5, "monday")).toBeNull();
    expect(may.firstOfQuarter("monday").toDateString()).toBe("2024-04-01");
    expect(may.lastOfQuarter("friday").toDateString()).toBe("2024-06-28");
    expect(may.firstOfYear("monday").toDateString()).toBe("2024-01-01");
    expect(may.lastOfYear("tuesday").toDateString()).toBe("2024-12-31");
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

  it("clamps ranges, averages instants, and detects unit boundaries", () => {
    const base = Tempo.parse("2024-05-15T12:00:00Z");

    expect(
      Tempo.parse("2024-05-01T00:00:00Z")
        .clamp("2024-05-10T00:00:00Z", "2024-05-20T00:00:00Z")
        .toISOString(),
    ).toBe("2024-05-10T00:00:00.000Z");
    expect(
      base.clamp("2024-05-10T00:00:00Z", "2024-05-20T00:00:00Z").toISOString(),
    ).toBe("2024-05-15T12:00:00.000Z");
    expect(
      Tempo.parse("2024-05-30T00:00:00Z")
        .clamp("2024-05-10T00:00:00Z", "2024-05-20T00:00:00Z")
        .toISOString(),
    ).toBe("2024-05-20T00:00:00.000Z");

    expect(base.average("2024-05-17T12:00:00Z").toISOString()).toBe(
      "2024-05-16T12:00:00.000Z",
    );
    expect(
      Tempo.average(
        "2024-05-15T00:00:00Z",
        "2024-05-17T00:00:00Z",
      ).toISOString(),
    ).toBe("2024-05-16T00:00:00.000Z");
    expect(
      average("2024-05-15T00:00:00Z", "2024-05-17T00:00:00Z").toISOString(),
    ).toBe("2024-05-16T00:00:00.000Z");
    expect(
      min("2024-05-15T00:00:00Z", "2024-05-10T00:00:00Z").toISOString(),
    ).toBe("2024-05-10T00:00:00.000Z");
    expect(
      max("2024-05-15T00:00:00Z", "2024-05-20T00:00:00Z").toISOString(),
    ).toBe("2024-05-20T00:00:00.000Z");
    expect(
      base
        .closest(
          "2024-05-10T00:00:00Z",
          "2024-05-16T00:00:00Z",
          "2024-05-20T00:00:00Z",
        )
        .toISOString(),
    ).toBe("2024-05-16T00:00:00.000Z");
    expect(
      base
        .farthest("2024-05-10T00:00:00Z", "2024-05-22T00:00:00Z")
        .toISOString(),
    ).toBe("2024-05-22T00:00:00.000Z");

    expect(Tempo.parse("2024-05-15T00:00:00Z").isStartOf("day")).toBe(true);
    expect(Tempo.parse("2024-05-15T00:00:01Z").isStartOf("day")).toBe(false);
    expect(Tempo.parse("2024-05-15T23:59:59.999Z").isEndOf("day")).toBe(true);
    expect(Tempo.parse("2024-05-15T23:59:59.998Z").isEndOf("day")).toBe(false);
  });

  it("serializes named date formats and maps components", () => {
    const tempo = Tempo.parse("2024-05-15T12:34:56.789Z", {
      timeZone: "Asia/Tokyo",
    });

    expect(tempo.toDateTimeLocalString()).toBe("2024-05-15T21:34:56");
    expect(tempo.toDateTimeLocalString("millisecond")).toBe(
      "2024-05-15T21:34:56.789",
    );
    expect(tempo.toTimeString("millisecond")).toBe("21:34:56.789");
    expect(tempo.toIso8601String()).toBe("2024-05-15T21:34:56+09:00");
    expect(tempo.toRfc3339String("millisecond")).toBe(
      "2024-05-15T21:34:56.789+09:00",
    );
    expect(tempo.toAtomString()).toBe("2024-05-15T21:34:56+09:00");
    expect(tempo.toRssString()).toBe("Wed, 15 May 2024 21:34:56 +0900");
    expect(tempo.toRfc7231String()).toBe("Wed, 15 May 2024 12:34:56 GMT");
    expect(tempo.toCookieString()).toBe("Wed, 15-May-2024 12:34:56 GMT");
    expect(tempo.toUnixString()).toBe("1715776496");

    expect(tempo.toMap().get("timeZone")).toBe("Asia/Tokyo");
    expect(tempo.toMap().get("hour")).toBe(21);
  });

  it("sets explicit components including zero values", () => {
    const tempo = Tempo.parse("2024-05-15T12:34:56.789Z", {
      timeZone: "UTC",
    });

    expect(tempo.setTime(0, 0, 0, 0).toISOString()).toBe(
      "2024-05-15T00:00:00.000Z",
    );
    expect(tempo.setHour(0).hour).toBe(0);
    expect(tempo.setMinute(0).minute).toBe(0);
    expect(tempo.setSecond(0).second).toBe(0);
    expect(tempo.setMillisecond(0).millisecond).toBe(0);
    expect(tempo.setDate(2025, 1, 2).toDateString()).toBe("2025-01-02");
  });
});
