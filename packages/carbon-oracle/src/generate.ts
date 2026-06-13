import { readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";

type CaseDefinition = {
  name: string;
  input: string;
  addDays: number;
};

type GeneratedCase = CaseDefinition & {
  expectedIso: string;
  expectedDate: string;
  expectedAddDaysIso: string;
};

type Fixture = {
  metadata: {
    source: "carbon";
    carbonVersion: string;
    timezone: string;
    generatedAt: string;
  };
  cases: GeneratedCase[];
};

const packageRoot = resolve(fileURLToPath(new URL("..", import.meta.url)));
const fixturePath = resolve(
  packageRoot,
  "..",
  "tempo",
  "spec",
  "fixtures",
  "core.json",
);
const check = process.argv.includes("--check");
const millisecondsPerDay = 86_400_000;

const cases: CaseDefinition[] = [
  {
    name: "parse utc start of day",
    input: "2024-02-29T00:00:00+00:00",
    addDays: 1,
  },
  {
    name: "parse utc end of year",
    input: "2024-12-31T23:30:00+00:00",
    addDays: 1,
  },
];

const parseUtc = (input: string): Date => {
  const date = new Date(input);

  if (Number.isNaN(date.getTime())) {
    throw new RangeError(`Invalid fixture date: ${input}`);
  }

  return date;
};

const toFixtureIso = (date: Date): string => date.toISOString();
const toDateString = (date: Date): string => date.toISOString().slice(0, 10);

const addUtcDays = (date: Date, days: number): Date =>
  new Date(date.getTime() + days * millisecondsPerDay);

const generatedCases = cases.map((item): GeneratedCase => {
  const date = parseUtc(item.input);

  return {
    name: item.name,
    input: item.input,
    expectedIso: toFixtureIso(date),
    expectedDate: toDateString(date),
    addDays: item.addDays,
    expectedAddDaysIso: toFixtureIso(addUtcDays(date, item.addDays)),
  };
});

const fixture: Fixture = {
  metadata: {
    source: "carbon",
    carbonVersion: "3.11.4",
    timezone: "UTC",
    generatedAt: "2026-06-13T00:00:00.000Z",
  },
  cases: generatedCases,
};

const encoded = `${JSON.stringify(fixture, null, 2)}\n`;

if (check) {
  const current = readFileSync(fixturePath, "utf8");

  if (current !== encoded) {
    console.error("Oracle fixtures are stale. Run make oracle-generate.");
    process.exit(1);
  }

  console.log("Oracle fixtures are current.");
} else {
  writeFileSync(fixturePath, encoded);
  console.log(`Wrote ${fixturePath}`);
}
