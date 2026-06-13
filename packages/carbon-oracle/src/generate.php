<?php

declare(strict_types=1);

$root = dirname(__DIR__, 2);
$fixturePath = $root . '/tempo/spec/fixtures/core.json';
$check = in_array('--check', $argv, true);

$autoload = __DIR__ . '/../vendor/autoload.php';
if (file_exists($autoload)) {
    require_once $autoload;
}

use Carbon\CarbonImmutable;

$cases = [
    [
        'name' => 'parse utc start of day',
        'input' => '2024-02-29T00:00:00+00:00',
        'addDays' => 1,
    ],
    [
        'name' => 'parse utc end of year',
        'input' => '2024-12-31T23:30:00+00:00',
        'addDays' => 1,
    ],
];

$generatedCases = array_map(static function (array $case): array {
    if (class_exists(CarbonImmutable::class)) {
        $date = CarbonImmutable::parse($case['input'])->utc();

        return [
            'name' => $case['name'],
            'input' => $case['input'],
            'expectedIso' => $date->format('Y-m-d\TH:i:s.v\Z'),
            'expectedDate' => $date->toDateString(),
            'addDays' => $case['addDays'],
            'expectedAddDaysIso' => $date->addDays($case['addDays'])->format('Y-m-d\TH:i:s.v\Z'),
        ];
    }

    $date = new DateTimeImmutable($case['input']);
    $date = $date->setTimezone(new DateTimeZone('UTC'));

    return [
        'name' => $case['name'],
        'input' => $case['input'],
        'expectedIso' => $date->format('Y-m-d\TH:i:s.v\Z'),
        'expectedDate' => $date->format('Y-m-d'),
        'addDays' => $case['addDays'],
        'expectedAddDaysIso' => $date->modify('+' . $case['addDays'] . ' day')->format('Y-m-d\TH:i:s.v\Z'),
    ];
}, $cases);

$fixture = [
    'metadata' => [
        'source' => 'carbon',
        'carbonVersion' => '3.11.4',
        'timezone' => 'UTC',
        'generatedAt' => '2026-06-13T00:00:00.000Z',
    ],
    'cases' => $generatedCases,
];

$encoded = json_encode($fixture, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES) . PHP_EOL;
$encoded = preg_replace_callback(
    '/^( +)/m',
    static fn (array $matches): string => str_repeat(' ', intdiv(strlen($matches[1]), 2)),
    $encoded
);

if ($check) {
    $current = file_get_contents($fixturePath);
    if ($current !== $encoded) {
        fwrite(STDERR, "Carbon oracle fixtures are stale. Run make oracle-generate.\n");
        exit(1);
    }

    echo "Carbon oracle fixtures are current.\n";
    exit(0);
}

file_put_contents($fixturePath, $encoded);
echo "Wrote {$fixturePath}\n";
