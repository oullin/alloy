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
  createFromDate,
  createFromTime,
  createMidnightDate,
  createSafe,
  fromJSON,
  hasFormat,
  maximum,
  max,
  minimum,
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

    const convertedMutable = immutable.toMutable();
    expect(convertedMutable).toBeInstanceOf(TempoMutable);
    expect(convertedMutable.addDays(1).toISOString()).toBe(
      "2024-03-01T00:00:00.000Z",
    );
    expect(immutable.toISOString()).toBe("2024-02-29T00:00:00.000Z");

    const convertedImmutable = convertedMutable.toImmutable();
    expect(convertedImmutable).toBeInstanceOf(TempoImmutable);
    convertedMutable.addDays(1);
    expect(convertedImmutable.toISOString()).toBe("2024-03-01T00:00:00.000Z");
    expect(convertedMutable.toISOString()).toBe("2024-03-02T00:00:00.000Z");
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
    expect(tokyo.createFromDate(2025, 1, 2).toISOString()).toBe(
      "2025-01-01T15:00:00.000Z",
    );
    expect(tokyo.createMidnightDate(2025, 1, 2).toISOString()).toBe(
      "2025-01-01T15:00:00.000Z",
    );
    expect(
      tokyo.createFromTime(9, 30, 15, 250).toTimeString("millisecond"),
    ).toBe("09:30:15.250");
    expect(
      createFromDate(2025, 1, 2, { timeZone: "Asia/Tokyo" }).toISOString(),
    ).toBe("2025-01-01T15:00:00.000Z");
    expect(
      createMidnightDate(2025, 1, 2, {
        timeZone: "Asia/Tokyo",
      }).toISOString(),
    ).toBe("2025-01-01T15:00:00.000Z");
    expect(
      createFromTime(9, 30, 15, 250, { timeZone: "UTC" }).toTimeString(
        "millisecond",
      ),
    ).toBe("09:30:15.250");
  });

  it("applies global settings to real date behavior", () => {
    const original = Tempo.settings();

    try {
      Tempo.setLocale("fr-FR");
      expect(Tempo.parse("2024-05-15").monthName()).toBe("mai");
      expect(Tempo.parse("2024-05-15").locale("en-US").monthName()).toBe("May");

      Tempo.setWeekendDays([5, 6]);
      expect(Tempo.parse("2024-05-17").isWeekend()).toBe(true);
      expect(Tempo.parse("2024-05-19").isWeekday()).toBe(true);

      Tempo.setMidDayAt(13);
      expect(Tempo.parse("2024-05-15T13:00:00Z").isMidday()).toBe(true);
      expect(Tempo.parse("2024-05-15").midday().hour).toBe(13);

      Tempo.useMonthsOverflow(false);
      expect(Tempo.parse("2024-01-31").addMonths(1).toDateString()).toBe(
        "2024-02-29",
      );
      Tempo.useYearsOverflow(false);
      expect(Tempo.parse("2024-02-29").addYears(1).toDateString()).toBe(
        "2025-02-28",
      );

      Tempo.useStrictMode(false);
      expect(Tempo.isStrictModeEnabled()).toBe(false);
      Tempo.setHumanDiffOptions({ locale: "en-US", numeric: "auto" });
      expect(Tempo.getHumanDiffOptions().numeric).toBe("auto");

      Tempo.setTestNowAndTimezone("2025-01-01T00:00:00Z", "Asia/Tokyo");
      expect(Tempo.hasTestNow()).toBe(true);
      expect(Tempo.now().timeZone).toBe("Asia/Tokyo");
      expect(Tempo.now().toDateTimeString()).toBe("2025-01-01 09:00:00");
      expect(Tempo.getTestNow()?.toISOString()).toBe(
        "2025-01-01T00:00:00.000Z",
      );
    } finally {
      Tempo.settings(original);
    }
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

  it("rejects normalized component construction through safe constructors", () => {
    expect(Tempo.create({ day: 31, month: 2, year: 2024 }).toDateString()).toBe(
      "2024-03-02",
    );
    expect(() => Tempo.createSafe({ day: 31, month: 2, year: 2024 })).toThrow(
      "Invalid Tempo local date/time components",
    );
    expect(
      Tempo.createSafe({
        day: 29,
        month: 2,
        timeZone: "Asia/Tokyo",
        year: 2024,
      }).toISOString(),
    ).toBe("2024-02-28T15:00:00.000Z");
    expect(
      createSafe({
        day: 29,
        month: 2,
        timeZone: "UTC",
        year: 2024,
      }).toDateString(),
    ).toBe("2024-02-29");
    expect(() =>
      TempoFactory.create({ timeZone: "Asia/Tokyo" }).createSafe({
        day: 31,
        month: 2,
        year: 2024,
      }),
    ).toThrow("Invalid Tempo local date/time components");
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
    expect(base.is("2024-05-15T23:59:00Z", "day")).toBe(true);
    expect(base.isCurrentUnit("day", "2024-05-15T23:59:00Z")).toBe(true);
    expect(base.isBetween(earlier, "2024-05-16T00:00:00Z")).toBe(true);
    expect(base.isBetween(base, "2024-05-16T00:00:00Z")).toBe(true);
    expect(base.between(earlier, "2024-05-16T00:00:00Z")).toBe(true);
    expect(base.between(base, "2024-05-16T00:00:00Z", false)).toBe(false);
    expect(base.betweenIncluded("2024-05-16T00:00:00Z", base)).toBe(true);
    expect(base.betweenExcluded(base, "2024-05-16T00:00:00Z")).toBe(false);
    expect(base.equalTo(base.clone())).toBe(true);
    expect(base.eq(base.clone())).toBe(true);
    expect(base.notEqualTo(earlier)).toBe(true);
    expect(base.ne(earlier)).toBe(true);
    expect(base.greaterThan(earlier)).toBe(true);
    expect(base.gt(earlier)).toBe(true);
    expect(base.greaterThanOrEqualTo(base.clone())).toBe(true);
    expect(base.gte(base.clone())).toBe(true);
    expect(earlier.lessThan(base)).toBe(true);
    expect(earlier.lt(base)).toBe(true);
    expect(earlier.lessThanOrEqualTo(earlier.clone())).toBe(true);
    expect(earlier.lte(earlier.clone())).toBe(true);
    expect(base.diffInHours(earlier)).toBe(26);
    expect(base.diffInMinutes(earlier)).toBe(1594);
    expect(base.diffAsDuration(earlier).toISOString()).toBe("P1DT2H34M45.600S");
    expect(base.diffAsDuration(earlier, { absolute: true }).toISOString()).toBe(
      "P1DT2H34M45.600S",
    );
    expect(base.diffAsDateInterval(earlier).toISOString()).toBe(
      "P1DT2H34M45.600S",
    );
    expect(base.diffAsTempoInterval(earlier).toISOString()).toBe(
      "P1DT2H34M45.600S",
    );
    expect(base.diffInMicroseconds(earlier)).toBe(95685600000);
    expect(base.diffFiltered((item) => item.isWeekday(), earlier)).toBe(1);
    expect(base.getPreciseTimestamp()).toBe(1715769285600000);
    expect(base.getPreciseTimestamp(3)).toBe(1715769285600);
    expect(base.calendar(earlier)).toBe("Tomorrow at 10:34");
    expect(base.calendar("2024-05-20T00:00:00Z")).toBe(
      "Last Wednesday at 10:34",
    );
    expect(base.floor("hour").toISOString()).toBe("2024-05-15T10:00:00.000Z");
    expect(base.ceil("hour").toISOString()).toBe("2024-05-15T11:00:00.000Z");
    expect(base.round("hour").toISOString()).toBe("2024-05-15T11:00:00.000Z");
    expect(base.floorUnit("hour").toISOString()).toBe(
      "2024-05-15T10:00:00.000Z",
    );
    expect(base.ceilUnit("hour").toISOString()).toBe(
      "2024-05-15T11:00:00.000Z",
    );
    expect(base.roundUnit("hour").toISOString()).toBe(
      "2024-05-15T11:00:00.000Z",
    );
    expect(base.floorWeek().toDateString()).toBe("2024-05-13");
    expect(base.ceilWeek().toDateString()).toBe("2024-05-20");
    expect(base.roundWeek().toDateString()).toBe("2024-05-13");
    expect(base.format("YYYY-MM-DD HH:mm:ss.SSS ZZ [Q]M")).toBe(
      "2024-05-15 10:34:45.600 +0000 Q5",
    );
    expect(base.rawFormat("YYYY-MM-DD")).toBe("2024-05-15");
    expect(base.isoFormat("YYYY-MM-DD")).toBe("2024-05-15");
    expect(base.translatedFormat("YYYY-MM-DD")).toBe("2024-05-15");
    expect(base.format("dddd, MMMM Do YYYY", { locale: "en-US" })).toBe(
      "Wednesday, May 15th 2024",
    );
    expect(base.ordinal()).toBe("15th");
    expect(base.ordinal("month")).toBe("5th");
    expect(base.meridiem()).toBe("AM");
    expect(base.addHours(2).meridiem(true)).toBe("pm");
    expect(base.week()).toBe(base.isoWeek);
    expect(base.weekYear()).toBe(base.isoWeekYear);
    expect(base.weeksInYear()).toBe(base.weeksInISOYear);
    expect(base.getDaysFromStartOfWeek()).toBe(2);
    expect(base.setDaysFromStartOfWeek(4).toDateString()).toBe("2024-05-17");
    expect(Tempo.parse("2015-06-01T00:00:00Z").isLongIsoYear()).toBe(true);
    expect(base.monthName()).toBe("May");
    expect(base.shortMonthName()).toBe("May");
    expect(base.dayName()).toBe("Wednesday");
    expect(base.shortDayName()).toBe("Wed");
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

    const inverted = new TempoInterval(
      "2024-01-03T12:00:00Z",
      "2024-01-01T00:00:00Z",
    );
    expect(inverted.isInverted).toBe(true);
    expect(inverted.hours).toBe(-60);
    expect(inverted.invert().hours).toBe(60);
    expect(inverted.absolute().start.toISOString()).toBe(
      "2024-01-01T00:00:00.000Z",
    );
    expect(inverted.absolute().end.toISOString()).toBe(
      "2024-01-03T12:00:00.000Z",
    );

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
    expect(
      period
        .filter((item) => item.day !== 3)
        .map((item) => item.toDateString()),
    ).toEqual(["2024-01-01", "2024-01-05"]);
    expect(period.map((item, index) => `${index}:${item.day}`)).toEqual([
      "0:1",
      "1:3",
      "2:5",
    ]);
    expect(period.every({ days: 1 }).count()).toBe(5);
    expect(period.toDuration().toISOString()).toBe("P4D");

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
    expect(base.daysInYear()).toBe(366);
    expect(base.isLongYear()).toBe(false);
    expect(Tempo.parse("2020-12-31T00:00:00Z").isLongYear()).toBe(true);
    expect(base.daysInMonth()).toBe(29);
    expect(base.isLastOfMonth()).toBe(true);
    expect(base.subDays(1).isLastOfMonth()).toBe(false);
    expect(saturday.isWeekend()).toBe(true);
    expect(base.isWeekday()).toBe(true);
    expect(base.isDayOfWeek("thursday")).toBe(true);
    expect(base.isDayOfWeek("friday")).toBe(false);
    expect(base.isToday("2024-02-29T12:00:00Z")).toBe(true);
    expect(base.isTomorrow("2024-02-28T12:00:00Z")).toBe(true);
    expect(base.isYesterday("2024-03-01T12:00:00Z")).toBe(true);
    expect(base.isPast("2024-03-01T12:00:00Z")).toBe(true);
    expect(base.isFuture("2024-02-28T12:00:00Z")).toBe(true);
    expect(base.isNowOrPast("2024-02-29T00:00:00Z")).toBe(true);
    expect(base.isNowOrFuture("2024-02-29T00:00:00Z")).toBe(true);
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
    expect(monday.nextOrSame("monday").toDateTimeString()).toBe(
      "2024-01-01 12:30:00",
    );
    expect(monday.previousOrSame("monday").toDateTimeString()).toBe(
      "2024-01-01 12:30:00",
    );
    expect(monday.nextOrSame("friday").toDateTimeString()).toBe(
      "2024-01-05 12:30:00",
    );
    expect(monday.previousOrSame("friday").toDateTimeString()).toBe(
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
    expect(may.nthOfQuarter(2, "monday")?.toDateString()).toBe("2024-04-08");
    expect(may.nthOfQuarter(-1, "friday")?.toDateString()).toBe("2024-06-28");
    expect(may.nthOfQuarter(14, "monday")).toBeNull();
    expect(may.firstOfYear("monday").toDateString()).toBe("2024-01-01");
    expect(may.lastOfYear("tuesday").toDateString()).toBe("2024-12-31");
    expect(may.nthOfYear(20, "monday")?.toDateString()).toBe("2024-05-13");
    expect(may.nthOfYear(-1, "tuesday")?.toDateString()).toBe("2024-12-31");
    expect(may.nthOfYear(54, "monday")).toBeNull();
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
    expect(parsed.toMap().get("hours")).toBe(4);
    expect(parsed.toArray()).toEqual([1, 0, 2, 0, 3, 4, 5, 6, 7]);
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
    expect(parsed.isPositive()).toBe(true);
    expect(TempoDuration.parse("-P1D").isNegative()).toBe(true);
    expect(TempoDuration.parse("PT0S").isPositive()).toBe(false);
    expect(TempoDuration.parse("PT0S").isNegative()).toBe(false);
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
    expect(
      Tempo.parse("2024-01-01")
        .toPeriod("2024-01-03")
        .toArray()
        .map((item) => item.toDateString()),
    ).toEqual(["2024-01-01", "2024-01-02", "2024-01-03"]);
    expect(Tempo.parse("2024-01-01").until("2024-01-02").count()).toBe(2);
    expect(Tempo.parse("2024-01-01").range("2024-01-02").count()).toBe(2);
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
    expect(wednesday.diffInUnit("day", friday)).toBe(5);
    expect(
      wednesday.diffInDaysFiltered((item) => item.isMonday(), friday),
    ).toBe(1);
    expect(
      wednesday.diffInHoursFiltered((item) => item.hour === 12, friday),
    ).toBe(5);
    expect(Tempo.parse("2024-10-01T00:00:00Z").diffInQuarters(friday)).toBe(1);
    expect(friday.diffInQuarters("2024-10-01T00:00:00Z")).toBe(-1);
    expect(friday.intervalUntil("2024-11-17T10:00:00Z").quarters).toBe(2);

    expect(friday.isSameSecond("2024-05-17T10:00:00.999Z")).toBe(true);
    expect(friday.isSameMinute("2024-05-17T10:00:59Z")).toBe(true);
    expect(friday.isSameHour("2024-05-17T10:59:59Z")).toBe(true);
    expect(friday.isSameDay("2024-05-17T23:59:59Z")).toBe(true);
    expect(friday.isSameWeek("2024-05-13T00:00:00Z")).toBe(true);
    expect(friday.isSameMonth("2024-05-01T00:00:00Z")).toBe(true);
    expect(friday.isSameQuarter("2024-04-01T00:00:00Z")).toBe(true);
    expect(friday.isSameYear("2024-12-31T23:59:59Z")).toBe(true);
    expect(friday.isSameAs("YYYY-MM-DD", "2024-05-17T23:59:59Z")).toBe(true);
    expect(friday.isSameUnit("day", "2024-05-17T23:59:59Z")).toBe(true);
    expect(friday.isSameAs("YYYY-MM-DD HH:mm", "2024-05-17T23:59:59Z")).toBe(
      false,
    );
    expect(friday.isBirthday("1990-05-17T00:00:00Z")).toBe(true);
    expect(friday.isBirthday("1990-05-18T00:00:00Z")).toBe(false);
    expect(friday.setTime(0, 0, 42, 0).secondsSinceMidnight()).toBe(42);
    expect(friday.setTime(23, 59, 17, 0).secondsUntilEndOfDay()).toBe(42);
    expect(friday.midDay().toTimeString()).toBe("12:00:00");
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
    expect(utc.getOffsetString()).toBe("+00:00");
    expect(utc.utcOffset()).toBe(0);
    expect(utc.timezoneName("shortOffset")).toBe("GMT+0");

    expect(winter.isUtc()).toBe(false);
    expect(winter.offsetMinutes).toBe(-300);
    expect(winter.offsetString()).toBe("-05:00");
    expect(winter.getOffsetString("")).toBe("-0500");
    expect(winter.utcOffset()).toBe(-300);
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
      minimum("2024-05-15T00:00:00Z", "2024-05-10T00:00:00Z").toISOString(),
    ).toBe("2024-05-10T00:00:00.000Z");
    expect(
      maximum("2024-05-15T00:00:00Z", "2024-05-20T00:00:00Z").toISOString(),
    ).toBe("2024-05-20T00:00:00.000Z");
    expect(base.minimum("2024-05-10T00:00:00Z").toISOString()).toBe(
      "2024-05-10T00:00:00.000Z",
    );
    expect(base.maximum("2024-05-20T00:00:00Z").toISOString()).toBe(
      "2024-05-20T00:00:00.000Z",
    );
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
    expect(Tempo.parse("2024-05-15T00:00:00Z").isStartOfDay()).toBe(true);
    expect(Tempo.parse("2024-05-15T00:00:01Z").isStartOf("day")).toBe(false);
    expect(Tempo.parse("2024-05-15T23:59:59.999Z").isEndOf("day")).toBe(true);
    expect(Tempo.parse("2024-05-15T23:59:59.999Z").isEndOfDay()).toBe(true);
    expect(Tempo.parse("2024-05-15T23:59:59.998Z").isEndOf("day")).toBe(false);
    expect(base.startOfMillisecond().toISOString()).toBe(
      "2024-05-15T12:00:00.000Z",
    );
    expect(base.endOfMillisecond().toISOString()).toBe(
      "2024-05-15T12:00:00.000Z",
    );
    expect(base.isStartOfMillisecond()).toBe(true);
    expect(base.isEndOfMillisecond()).toBe(true);
    expect(
      Tempo.parse("2024-05-15T12:34:56.789Z").startOfSecond().toISOString(),
    ).toBe("2024-05-15T12:34:56.000Z");
    expect(
      Tempo.parse("2024-05-15T12:34:56.789Z").endOfSecond().toISOString(),
    ).toBe("2024-05-15T12:34:56.999Z");
    expect(Tempo.parse("2024-05-15T12:34:00Z").isStartOfMinute()).toBe(true);
    expect(Tempo.parse("2024-05-15T12:34:59.999Z").isEndOfMinute()).toBe(true);
    expect(Tempo.parse("2024-05-15T12:00:00Z").isStartOfHour()).toBe(true);
    expect(Tempo.parse("2024-05-15T12:59:59.999Z").isEndOfHour()).toBe(true);
    expect(Tempo.parse("2024-05-13T00:00:00Z").isStartOfWeek()).toBe(true);
    expect(Tempo.parse("2024-05-19T23:59:59.999Z").isEndOfWeek()).toBe(true);
    expect(Tempo.parse("2024-05-01T00:00:00Z").isStartOfMonth()).toBe(true);
    expect(Tempo.parse("2024-05-31T23:59:59.999Z").isEndOfMonth()).toBe(true);
    expect(Tempo.parse("2024-04-01T00:00:00Z").isStartOfQuarter()).toBe(true);
    expect(Tempo.parse("2024-06-30T23:59:59.999Z").isEndOfQuarter()).toBe(true);
    expect(Tempo.parse("2024-01-01T00:00:00Z").isStartOfYear()).toBe(true);
    expect(Tempo.parse("2024-12-31T23:59:59.999Z").isEndOfYear()).toBe(true);
    expect(base.startOfDecade().toDateTimeString()).toBe("2020-01-01 00:00:00");
    expect(base.endOfDecade().toDateTimeString()).toBe("2029-12-31 23:59:59");
    expect(Tempo.parse("2020-01-01T00:00:00Z").isStartOfDecade()).toBe(true);
    expect(Tempo.parse("2020-01-01T00:00:00Z").isStartOfUnit("decade")).toBe(
      true,
    );
    expect(Tempo.parse("2029-12-31T23:59:59.999Z").isEndOfDecade()).toBe(true);
    expect(Tempo.parse("2029-12-31T23:59:59.999Z").isEndOfUnit("decade")).toBe(
      true,
    );
    expect(base.startOfCentury().toDateTimeString()).toBe(
      "2001-01-01 00:00:00",
    );
    expect(base.endOfCentury().toDateTimeString()).toBe("2100-12-31 23:59:59");
    expect(Tempo.parse("2001-01-01T00:00:00Z").isStartOfCentury()).toBe(true);
    expect(Tempo.parse("2100-12-31T23:59:59.999Z").isEndOfCentury()).toBe(true);
    expect(base.startOfMillennium().toDateTimeString()).toBe(
      "2001-01-01 00:00:00",
    );
    expect(base.endOfMillennium().toDateTimeString()).toBe(
      "3000-12-31 23:59:59",
    );
    expect(Tempo.parse("2001-01-01T00:00:00Z").isStartOfMillennium()).toBe(
      true,
    );
    expect(Tempo.parse("3000-12-31T23:59:59.999Z").isEndOfMillennium()).toBe(
      true,
    );
    expect(Tempo.parse(-8640000000000000).isStartOfTime()).toBe(true);
    expect(Tempo.parse(8640000000000000).isEndOfTime()).toBe(true);
  });

  it("serializes named date formats and maps components", () => {
    const tempo = Tempo.parse("2024-05-15T12:34:56.789Z", {
      timeZone: "Asia/Tokyo",
    });

    expect(tempo.toDateTimeLocalString()).toBe("2024-05-15T21:34:56");
    expect(tempo.toDateTimeLocalString("millisecond")).toBe(
      "2024-05-15T21:34:56.789",
    );
    expect(tempo.toFormattedDateString()).toBe("May 15, 2024");
    expect(tempo.toFormattedDayDateString()).toBe("Wed, May 15, 2024");
    expect(tempo.toDayDateTimeString()).toBe("Wed, May 15, 2024 9:34 PM");
    expect(tempo.toTimeString("millisecond")).toBe("21:34:56.789");
    expect(tempo.toIso8601String()).toBe("2024-05-15T21:34:56+09:00");
    expect(tempo.toIso8601ZuluString()).toBe("2024-05-15T12:34:56Z");
    expect(tempo.toIso8601ZuluString("millisecond")).toBe(
      "2024-05-15T12:34:56.789Z",
    );
    expect(tempo.toRfc3339String("millisecond")).toBe(
      "2024-05-15T21:34:56.789+09:00",
    );
    expect(tempo.toRfc822String()).toBe("Wed, 15 May 24 21:34:56 +0900");
    expect(tempo.toRfc850String()).toBe("Wednesday, 15-May-24 21:34:56 +0900");
    expect(tempo.toRfc1036String()).toBe("Wed, 15 May 24 21:34:56 +0900");
    expect(tempo.toRfc1123String()).toBe("Wed, 15 May 2024 21:34:56 +0900");
    expect(tempo.toRfc2822String()).toBe("Wed, 15 May 2024 21:34:56 +0900");
    expect(tempo.toW3cString()).toBe("2024-05-15T21:34:56+09:00");
    expect(tempo.toAtomString()).toBe("2024-05-15T21:34:56+09:00");
    expect(tempo.toRssString()).toBe("Wed, 15 May 2024 21:34:56 +0900");
    expect(tempo.toRfc7231String()).toBe("Wed, 15 May 2024 12:34:56 GMT");
    expect(tempo.toCookieString()).toBe("Wed, 15-May-2024 12:34:56 GMT");
    expect(tempo.toUnixString()).toBe("1715776496");
    expect(tempo.unix()).toBe(1715776496);
    expect(tempo.getTimestampMs()).toBe(1715776496789);
    expect(tempo.toDateTime().toISOString()).toBe("2024-05-15T12:34:56.789Z");
    expect(tempo.toDateTimeImmutable().toISOString()).toBe(
      "2024-05-15T12:34:56.789Z",
    );
    expect(tempo.jsonSerialize()).toBe("2024-05-15T12:34:56.789Z");
    expect(tempo.serialize()).toBe("2024-05-15T12:34:56.789Z");
    expect(JSON.stringify(tempo)).toBe('"2024-05-15T12:34:56.789Z"');
    expect(JSON.stringify(TempoDuration.parse("P1DT2H"))).toBe('"P1DT2H"');
    expect(fromJSON('"2024-05-15T12:34:56.789Z"').toISOString()).toBe(
      "2024-05-15T12:34:56.789Z",
    );
    expect(
      TempoImmutable.fromJSON('"2024-05-15T12:34:56.789Z"').toISOString(),
    ).toBe("2024-05-15T12:34:56.789Z");
    expect(
      TempoMutable.fromJSON('"2024-05-15T12:34:56.789Z"').toISOString(),
    ).toBe("2024-05-15T12:34:56.789Z");
    expect(TempoDuration.fromJSON('"P1DT2H"').toISOString()).toBe("P1DT2H");
    expect(() => Tempo.fromJSON("123")).toThrow("Tempo JSON must be a string");
    expect(Tempo.make(null)).toBeNull();
    expect(Tempo.make("2024-05-15")?.toDateString()).toBe("2024-05-15");
    expect(Tempo.parseFromLocale("2024-05-15", "en-US").toDateString()).toBe(
      "2024-05-15",
    );
    expect(Tempo.getDays().slice(0, 2)).toEqual(["Sunday", "Monday"]);
    expect(Tempo.hasFormatWithModifiers("2024/05/15", "YYYY/MM/DD")).toBe(true);
    expect(Tempo.hasFormatWithModifiers(null, "YYYY/MM/DD")).toBe(false);

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
    expect(tempo.setTime(0, 0, 0, 0).isMidnight()).toBe(true);
    expect(tempo.midday().toISOString()).toBe("2024-05-15T12:00:00.000Z");
    expect(tempo.midday().isMidday()).toBe(true);
    expect(tempo.setHour(0).hour).toBe(0);
    expect(tempo.setMinute(0).minute).toBe(0);
    expect(tempo.setSecond(0).second).toBe(0);
    expect(tempo.setMillisecond(0).millisecond).toBe(0);
    expect(tempo.setDate(2025, 1, 2).toDateString()).toBe("2025-01-02");
    expect(tempo.setUnit("day", 2).toDateString()).toBe("2024-05-02");
    expect(tempo.setDateTime(2025, 1, 2, 3, 4, 5, 6).toISOString()).toBe(
      "2025-01-02T03:04:05.006Z",
    );
    expect(
      tempo.setDateFrom("2025-01-02T03:04:05.006Z").toDateTimeString(),
    ).toBe("2025-01-02 12:34:56");
    expect(
      tempo.setTimeFrom("2025-01-02T03:04:05.006Z").toDateTimeString(),
    ).toBe("2024-05-15 03:04:05");
    expect(
      tempo.setDateTimeFrom("2025-01-02T03:04:05.006Z").toISOString(),
    ).toBe("2025-01-02T03:04:05.006Z");
    expect(tempo.setTimeFromTimeString("03:04:05.006").toISOString()).toBe(
      "2024-05-15T03:04:05.006Z",
    );
    expect(tempo.setTimestamp(0).toISOString()).toBe(
      "1970-01-01T00:00:00.000Z",
    );
    expect(tempo.timestampTo(0).toISOString()).toBe("1970-01-01T00:00:00.000Z");
    expect(tempo.setISODate(2024, 20, 3).toDateString()).toBe("2024-05-15");
    expect(tempo.weekday()).toBe(3);
    expect(tempo.weekday(5).toDateString()).toBe("2024-05-17");
    expect(tempo.setWeekday("monday").toDateString()).toBe("2024-05-13");
    expect(tempo.setDayOfYear(60).toDateString()).toBe("2024-02-29");
    expect(tempo.modify("+2 days").toDateString()).toBe("2024-05-17");
    expect(tempo.modify("previous day").toDateString()).toBe("2024-05-14");
    expect(tempo.change("next week").toDateString()).toBe("2024-05-22");
    expect(tempo.subtract(2, "days").toDateString()).toBe("2024-05-13");
    expect(tempo.addUnit("day", 2).toDateString()).toBe("2024-05-17");
    expect(tempo.addRealUnit("day", 2).toDateString()).toBe("2024-05-17");
    expect(tempo.addUTCUnit("day", 2).toDateString()).toBe("2024-05-17");
    expect(tempo.rawAdd(2, "day").toDateString()).toBe("2024-05-17");
    expect(tempo.addUnitNoOverflow("day", 20, "month").toDateString()).toBe(
      "2024-05-31",
    );
    expect(tempo.subUnit("day", 2).toDateString()).toBe("2024-05-13");
    expect(tempo.subRealUnit("day", 2).toDateString()).toBe("2024-05-13");
    expect(tempo.subUTCUnit("day", 2).toDateString()).toBe("2024-05-13");
    expect(tempo.rawSub(2, "day").toDateString()).toBe("2024-05-13");
    expect(tempo.subUnitNoOverflow("day", 20, "month").toDateString()).toBe(
      "2024-05-01",
    );
    expect(tempo.setUnitNoOverflow("day", 40, "month").toDateString()).toBe(
      "2024-05-31",
    );
    expect(Tempo.parse("2024-05-17T12:00:00Z").nextWeekendDay().dayOfWeek).toBe(
      6,
    );
    expect(
      Tempo.parse("2024-05-20T12:00:00Z").previousWeekendDay().dayOfWeek,
    ).toBe(0);
    expect(tempo.tz("Asia/Tokyo").timeZone).toBe("Asia/Tokyo");
    expect(tempo.shiftTimezone("Asia/Tokyo").toDateTimeString()).toBe(
      "2024-05-15 12:34:56",
    );
    expect(
      Tempo.createFromTimeString("03:04:05.006").toTimeString("millisecond"),
    ).toBe("03:04:05.006");
    expect(
      Tempo.createFromFormat("2024/05/15", "YYYY/MM/DD").toDateString(),
    ).toBe("2024-05-15");
    expect(
      Tempo.rawCreateFromFormat("2024/05/15", "YYYY/MM/DD").toDateString(),
    ).toBe("2024-05-15");
    expect(Tempo.rawParse("2024-05-15").toDateString()).toBe("2024-05-15");
    expect(Tempo.canBeCreatedFromFormat("2024/05/15", "YYYY/MM/DD")).toBe(true);
    expect(Tempo.createFromTimestamp(0).toISOString()).toBe(
      "1970-01-01T00:00:00.000Z",
    );
    expect(Tempo.createFromTimestampMs(1).toISOString()).toBe(
      "1970-01-01T00:00:00.001Z",
    );
    expect(Tempo.fromTimestampUTC(0).toISOString()).toBe(
      "1970-01-01T00:00:00.000Z",
    );
    expect(Tempo.fromTimestampMsUTC(1).toISOString()).toBe(
      "1970-01-01T00:00:00.001Z",
    );
    expect(Tempo.createFromTimestampUTC(0).toISOString()).toBe(
      "1970-01-01T00:00:00.000Z",
    );
    expect(Tempo.createFromTimestampMsUTC(1).toISOString()).toBe(
      "1970-01-01T00:00:00.001Z",
    );
    expect(tempo.addDays(2).from(tempo)).toBe("in 2 days");
    expect(tempo.addDays(2).since(tempo)).toBe("in 2 days");
    expect(tempo.to(tempo.addDays(2))).toBe("in 2 days");
    expect(tempo.addDays(2).timespan(tempo)).toBe("in 2 days");
    expect(tempo.isImmutable()).toBe(true);
    expect(tempo.avoidMutation().toISOString()).toBe(
      "2024-05-15T12:34:56.789Z",
    );
    expect(tempo.cast().toISOString()).toBe("2024-05-15T12:34:56.789Z");
    expect(tempo.tempoize(tempo.addDays(1)).toDateString()).toBe("2024-05-16");
    expect(tempo.nowWithSameTz().timeZone).toBe("UTC");
    expect(TempoImmutable.parse(tempo).isImmutable()).toBe(true);
    expect(tempo.isMutable()).toBe(false);
    expect(TempoImmutable.parse(tempo).isMutable()).toBe(false);
    const mutable = TempoMutable.parse(tempo);
    expect(mutable.isMutable()).toBe(true);
    expect(mutable.avoidMutation().toISOString()).toBe(
      "2024-05-15T12:34:56.789Z",
    );
    expect(mutable.cast().toISOString()).toBe("2024-05-15T12:34:56.789Z");
    expect(mutable.tempoize(tempo.addDays(1)).toDateString()).toBe(
      "2024-05-16",
    );
    expect(mutable.nowWithSameTz().timeZone).toBe("UTC");
  });
});
