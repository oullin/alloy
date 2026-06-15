import type {
  CalendarFormatKey,
  FormatOptions,
  HumanDiffOptions,
  TimeZoneNameStyle,
} from "../types";
import {
  getFormatter,
  getZonedParts,
  millisecondsPerDay,
  millisecondsPerHour,
  millisecondsPerMinute,
  millisecondsPerSecond,
  millisecondsPerWeek,
  monthNames,
  normalizeTimeZone,
  pad,
  readPart,
  weekdayNames,
} from "../calendar";

export const calendarFormatDefaults = (): Record<
  CalendarFormatKey,
  string
> => ({
  lastDay: "[Yesterday at] HH:mm",
  lastWeek: "[Last] dddd [at] HH:mm",
  nextDay: "[Tomorrow at] HH:mm",
  nextWeek: "dddd [at] HH:mm",
  sameDay: "[Today at] HH:mm",
  sameElse: "YYYY-MM-DD",
});

export const isoFormatDefaults = (): Record<string, string> => ({
  atom: "YYYY-MM-DDTHH:mm:ssZZ",
  cookie: "ddd, DD-MMM-YYYY HH:mm:ss [GMT]",
  date: "YYYY-MM-DD",
  dateTime: "YYYY-MM-DD HH:mm:ss",
  iso8601: "YYYY-MM-DDTHH:mm:ssZZ",
  time: "HH:mm:ss",
});

export const timeZoneName = (
  date: Date,
  timeZone: string,
  style: TimeZoneNameStyle,
  locale: string,
): string => {
  const parts = getFormatter(timeZone, style, locale).formatToParts(date);

  return readPart(parts, "timeZoneName") || timeZone;
};

export const formatOffset = (
  offsetMinutes: number,
  separator: ":" | "",
): string => {
  const sign = offsetMinutes >= 0 ? "+" : "-";
  const absolute = Math.abs(offsetMinutes);
  const hours = Math.trunc(absolute / 60);
  const minutes = absolute % 60;

  return `${sign}${pad(hours)}${separator}${pad(minutes)}`;
};

export const ordinal = (value: number): string => {
  const remainder = value % 100;

  if (remainder >= 11 && remainder <= 13) {
    return `${value}th`;
  }

  switch (value % 10) {
    case 1:
      return `${value}st`;
    case 2:
      return `${value}nd`;
    case 3:
      return `${value}rd`;
    default:
      return `${value}th`;
  }
};

export const bestRelativeUnit = (
  milliseconds: number,
): NonNullable<HumanDiffOptions["unit"]> => {
  const absolute = Math.abs(milliseconds);

  if (absolute < millisecondsPerMinute) {
    return "second";
  }
  if (absolute < millisecondsPerHour) {
    return "minute";
  }
  if (absolute < millisecondsPerDay) {
    return "hour";
  }
  if (absolute < millisecondsPerWeek) {
    return "day";
  }
  if (absolute < millisecondsPerDay * 30) {
    return "week";
  }
  if (absolute < millisecondsPerDay * 365) {
    return "month";
  }

  return "year";
};

export const unitDivisor = (
  unit: NonNullable<HumanDiffOptions["unit"]>,
): number => {
  switch (unit) {
    case "second":
      return millisecondsPerSecond;
    case "minute":
      return millisecondsPerMinute;
    case "hour":
      return millisecondsPerHour;
    case "day":
      return millisecondsPerDay;
    case "week":
      return millisecondsPerWeek;
    case "month":
      return millisecondsPerDay * 30;
    case "year":
      return millisecondsPerDay * 365;
  }
};

export class TempoFormatter {
  format(
    input: {
      readonly value: Date;
      readonly currentLocale: string;
      readonly zone: string;
      readonly timestamp: number;
      readonly timestampMs: number;
      readonly offsetFor: (timeZone: string) => number;
    },
    pattern: string,
    options?: FormatOptions,
  ): string {
    const locale = options?.locale ?? input.currentLocale;
    const timeZone = normalizeTimeZone(options?.timeZone ?? input.zone);
    const parts = getZonedParts(input.value, timeZone);
    const offset = input.offsetFor(timeZone);
    const monthsShort = monthNames(locale, "short");
    const monthsLong = monthNames(locale, "long");
    const weekdaysShort = weekdayNames(locale, "short");
    const weekdaysLong = weekdayNames(locale, "long");
    const hour12 = parts.hour % 12 === 0 ? 12 : parts.hour % 12;
    const replacements: Record<string, string> = {
      A: parts.hour < 12 ? "AM" : "PM",
      a: parts.hour < 12 ? "am" : "pm",
      D: String(parts.day),
      DD: pad(parts.day),
      Do: ordinal(parts.day),
      d: String(parts.weekday),
      ddd: weekdaysShort[parts.weekday] ?? "",
      dddd: weekdaysLong[parts.weekday] ?? "",
      H: String(parts.hour),
      HH: pad(parts.hour),
      h: String(hour12),
      hh: pad(hour12),
      M: String(parts.month),
      MM: pad(parts.month),
      MMM: monthsShort[parts.month - 1] ?? "",
      MMMM: monthsLong[parts.month - 1] ?? "",
      m: String(parts.minute),
      mm: pad(parts.minute),
      SSS: pad(parts.millisecond, 3),
      s: String(parts.second),
      ss: pad(parts.second),
      X: String(input.timestamp),
      x: String(input.timestampMs),
      Y: String(parts.year),
      YY: pad(parts.year % 100),
      YYYY: pad(parts.year, 4),
      Z: formatOffset(offset, ":"),
      ZZ: formatOffset(offset, ""),
    };

    return pattern.replace(
      /\[[^\]]*]|YYYY|MMMM|dddd|MMM|ddd|SSS|Do|YY|ZZ|MM|DD|HH|hh|mm|ss|Z|X|x|Y|M|D|H|h|m|s|A|a|d/g,
      (token) =>
        token.startsWith("[") && token.endsWith("]")
          ? token.slice(1, -1)
          : (replacements[token] ?? token),
    );
  }
}
