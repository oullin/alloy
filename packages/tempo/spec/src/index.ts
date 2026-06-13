import { resolve } from "node:path";
import { fileURLToPath } from "node:url";

export const carbonVersion = "3.11.4";

export const specRoot = resolve(fileURLToPath(new URL("..", import.meta.url)));
export const fixturesRoot = resolve(specRoot, "fixtures");

export type TempoCase = {
  name: string;
  input: string;
  expectedIso: string;
  expectedDate: string;
  addDays?: number;
  expectedAddDaysIso?: string;
};

export type TempoFixture = {
  metadata: {
    source: "carbon";
    carbonVersion: string;
    timezone: string;
    generatedAt: string;
  };
  cases: TempoCase[];
};
