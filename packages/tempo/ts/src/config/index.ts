import type { HumanDiffOptions, TempoSerializer } from "../types";
import { defaultTimeZone } from "../calendar";

export const defaultHumanDiffOptions: HumanDiffOptions = {
  locale: "en-US",
  numeric: "always",
  style: "long",
};

export const tempoConfig = {
  fallbackLocale: "en-US",
  humanDiffOptions: { ...defaultHumanDiffOptions },
  locale: "en-US",
  midDayAt: 12,
  monthsOverflow: true,
  strictMode: true,
  testNow: null as Date | null,
  timeZone: defaultTimeZone,
  toStringFormat: null as string | null,
  weekendDays: [0, 6],
  yearsOverflow: true,
};

export const tempoState: {
  serializer: TempoSerializer | null;
  lastError: unknown;
} = {
  serializer: null,
  lastError: null,
};

export class TempoClock {
  now(): Date {
    return configuredNow();
  }
}

export class TempoSettingsStore {
  snapshot() {
    return tempoConfig;
  }
}

export const configuredNow = (): Date =>
  tempoConfig.testNow === null
    ? new Date()
    : new Date(tempoConfig.testNow.getTime());
