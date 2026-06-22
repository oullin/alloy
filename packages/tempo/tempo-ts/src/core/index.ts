import { cloneTempoPolicy, defaultTempoPolicy, policyToSettings, resolveTempoPolicy } from '#config';
import { TempoDuration, durationFromInput } from '#duration';
import { TempoInterval, TempoPeriod } from '#ranges';
import { replaceTranslationTokens, type TempoRuntime } from '#runtime';
import { bestRelativeUnit, calendarFormatDefaults, formatOffset, isoFormatDefaults, ordinal, timeZoneName, TempoFormatter, unitDivisor } from '#formatting';
import { asDate, fromNumericTimestamp, parseFromPattern, TempoParser, zoneFromInput } from '#parsing';

import type {
	BoundaryUnit,
	CalendarFormatKey,
	CalendarFormats,
	ComparisonUnit,
	DiffOptions,
	DurationInput,
	FormatOptions,
	HumanDiffOptions,
	PeriodOptions,
	StartOfWeekOptions,
	TempoComponents,
	TempoInput,
	TempoObject,
	TempoOptions,
	TempoPolicy,
	TempoSettableComponents,
	TempoSettings,
	TempoTranslator,
	TimeStringPrecision,
	TimeUnit,
	TimeZoneNameStyle,
	WeekdayInput,
} from '#types';

import {
	assertFiniteNumber,
	assertSafeZonedComponents,
	dateFromPartsAsUTC,
	dateFromZonedComponents,
	daysInMonth,
	defaultTimeZone,
	fixedUnitMilliseconds,
	getZonedParts,
	isoWeekData,
	millisecondsPerMinute,
	millisecondsPerSecond,
	monthNames,
	normalizeTimeZone,
	normalizeUnit,
	pad,
	resolveWeekday,
	weekdayNames,
	weeksInISOYear,
	type ZonedParts,
} from '#calendar';

type ComparableTempo = {
	readonly timestampMs: number;
	startOf: (unit: BoundaryUnit) => { readonly timestampMs: number };
};

class TempoComparison {
	comparableValue(source: ComparableTempo, unit: ComparisonUnit): number {
		return unit === 'millisecond' ? source.timestampMs : source.startOf(unit).timestampMs;
	}
}

const tempoComparison = new TempoComparison();
const tempoParser = new TempoParser();
const averageMilliseconds = (left: number, right: number): number => Math.trunc(left / 2) + Math.trunc(right / 2) + Math.trunc(((left % 2) + (right % 2)) / 2);
const millisecondDistance = (left: number, right: number): number => Math.abs(left - right);

const optionsFromPolicy = (policy: TempoPolicy, overrides: TempoOptions = {}): TempoOptions => ({
	fallbackLocale: policy.fallbackLocale,
	humanDiffOptions: policy.humanDiffOptions,
	locale: policy.locale,
	midDayAt: policy.midDayAt,
	monthsOverflow: policy.monthsOverflow,
	runtime: policy.runtime,
	serializer: policy.serializer,
	strictMode: policy.strictMode,
	testNow: policy.testNow,
	timeZone: policy.timeZone,
	toStringFormat: policy.toStringFormat,
	translator: policy.translator,
	weekendDays: policy.weekendDays,
	yearsOverflow: policy.yearsOverflow,
	...overrides,
});

export class TempoImmutable {
	protected value: Date;
	protected currentLocale: string;
	protected policy: TempoPolicy;
	protected runtime: TempoRuntime;
	protected zone: string;

	constructor(input: TempoInput = new Date(), options?: TempoOptions) {
		const basePolicy = input instanceof TempoImmutable ? input.policySnapshot() : defaultTempoPolicy();
		const policy = resolveTempoPolicy(options, basePolicy);

		this.value = tempoParser.asDate(input, optionsFromPolicy(policy, options));
		this.runtime = tempoParser.runtimeFromOptions(input, optionsFromPolicy(policy, options));
		this.currentLocale = options?.locale ?? policy.locale ?? this.runtime.locale;
		this.zone = tempoParser.zoneFromInput(input, optionsFromPolicy(policy, options));
		this.policy = cloneTempoPolicy({
			...policy,
			locale: this.currentLocale,
			runtime: this.runtime,
			timeZone: this.zone,
		});
	}

	static now(options?: TempoOptions): TempoImmutable {
		const policy = resolveTempoPolicy(options);

		return new TempoImmutable(policy.testNow ?? new Date(), optionsFromPolicy(policy));
	}

	static today(options?: TempoOptions): TempoImmutable {
		return TempoImmutable.now(options).startOfDay();
	}

	static tomorrow(options?: TempoOptions): TempoImmutable {
		return TempoImmutable.today(options).addDays(1);
	}

	static yesterday(options?: TempoOptions): TempoImmutable {
		return TempoImmutable.today(options).subDays(1);
	}

	static parse(input: TempoInput, options?: TempoOptions): TempoImmutable {
		return new TempoImmutable(input, options);
	}

	static fromJSON(input: string, options?: TempoOptions): TempoImmutable {
		const value = JSON.parse(input) as unknown;

		if (typeof value !== 'string') {
			throw new RangeError('Tempo JSON must be a string');
		}

		return TempoImmutable.parse(value, options);
	}

	static fromSerialized(input: string, options?: TempoOptions): TempoImmutable {
		return TempoImmutable.fromJSON(input, options);
	}

	static make(input: TempoInput | null | undefined, options?: TempoOptions): TempoImmutable | null {
		return input === null || input === undefined ? null : TempoImmutable.parse(input, options);
	}

	static parseFromLocale(input: string, _locale?: string, options?: TempoOptions): TempoImmutable {
		return TempoImmutable.parse(input, options);
	}

	static days(locale = 'en-US'): string[] {
		return [...weekdayNames(locale, 'long')];
	}

	static availableLocales(): string[] {
		const policy = defaultTempoPolicy();

		return [...new Set(Intl.DateTimeFormat.supportedLocalesOf([policy.locale, policy.fallbackLocale, 'en-US']))];
	}

	static availableLocalesInfo(): Record<string, { readonly name: string }> {
		const displayNames = typeof Intl.DisplayNames === 'function' ? new Intl.DisplayNames([defaultTempoPolicy().locale], { type: 'language' }) : null;

		return Object.fromEntries(TempoImmutable.availableLocales().map((locale) => [locale, { name: displayNames?.of(locale) ?? locale }]));
	}

	static calendarFormats(): Record<CalendarFormatKey, string> {
		return calendarFormatDefaults();
	}

	static formatsToIsoReplacements(): Record<string, string> {
		return {
			A: 'A',
			D: 'D',
			H: 'H',
			M: 'M',
			S: 'S',
			Y: 'Y',
			d: 'd',
			h: 'h',
			m: 'm',
			s: 's',
		};
	}

	static isoFormats(): Record<string, string> {
		return isoFormatDefaults();
	}

	static isoUnits(): BoundaryUnit[] {
		return ['millisecond', 'second', 'minute', 'hour', 'day', 'week', 'month', 'quarter', 'year'];
	}

	static timeFormatByPrecision(precision: TimeStringPrecision = 'second'): string {
		return precision === 'millisecond' ? 'HH:mm:ss.SSS' : 'HH:mm:ss';
	}

	static weekStartsAt(): number {
		return 1;
	}

	static weekEndsAt(): number {
		return 0;
	}

	static localeHasDiffSyntax(locale: string): boolean {
		return Intl.RelativeTimeFormat.supportedLocalesOf([locale]).length > 0;
	}

	static localeHasDiffOneDayWords(locale: string): boolean {
		return TempoImmutable.localeHasDiffSyntax(locale);
	}

	static localeHasDiffTwoDayWords(locale: string): boolean {
		return TempoImmutable.localeHasDiffSyntax(locale);
	}

	static localeHasPeriodSyntax(locale: string): boolean {
		return Intl.ListFormat.supportedLocalesOf([locale]).length > 0;
	}

	static localeHasShortUnits(locale: string): boolean {
		return Intl.RelativeTimeFormat.supportedLocalesOf([locale]).length > 0;
	}

	static sleep(seconds: number): Promise<void> {
		assertFiniteNumber(seconds, 'Seconds');

		return new Promise((resolve) => {
			setTimeout(resolve, Math.max(0, seconds * 1000));
		});
	}

	static hasRelativeKeywords(input: string): boolean {
		return /\b(now|today|tomorrow|yesterday|next|last|ago)\b|^[+-]/i.test(input.trim());
	}

	static isModifiableUnit(unit: TimeUnit): boolean {
		try {
			normalizeUnit(unit);

			return true;
		} catch {
			return false;
		}
	}

	static singularUnit(unit: TimeUnit): BoundaryUnit {
		return normalizeUnit(unit) as BoundaryUnit;
	}

	static pluralUnit(unit: TimeUnit): string {
		return `${normalizeUnit(unit)}s`;
	}

	static fromFormat(input: string, pattern: string, options?: TempoOptions): TempoImmutable {
		return new TempoImmutable(parseFromPattern(input, pattern, options), options);
	}

	static createNormalized(components: TempoComponents, options?: TempoOptions): TempoImmutable {
		const policy = resolveTempoPolicy(options);
		const timeZone = normalizeTimeZone(components.timeZone ?? policy.timeZone);

		return new TempoImmutable(dateFromZonedComponents(components), {
			...optionsFromPolicy(policy, options),
			strictMode: false,
			timeZone,
		});
	}

	static create(components: TempoComponents, options?: TempoOptions): TempoImmutable {
		const policy = resolveTempoPolicy(options);
		const timeZone = normalizeTimeZone(components.timeZone ?? policy.timeZone);
		const date = dateFromZonedComponents(components, timeZone);

		assertSafeZonedComponents(components, date, timeZone);

		return new TempoImmutable(date, optionsFromPolicy(policy, { timeZone }));
	}

	static createStrict(components: TempoComponents): TempoImmutable {
		return TempoImmutable.create(components);
	}

	static instance(input: TempoInput, options?: TempoOptions): TempoImmutable {
		return TempoImmutable.parse(input, options);
	}

	static fromDate(year: number, month = 1, day = 1, options?: TempoOptions): TempoImmutable {
		return TempoImmutable.create({ day, month, timeZone: options?.timeZone, year }, options);
	}

	static fromTime(hour = 0, minute = 0, second = 0, millisecond = 0, options?: TempoOptions): TempoImmutable {
		return TempoImmutable.today(options).setTime(hour, minute, second, millisecond);
	}

	static fromTimeString(time: string, options?: TempoOptions): TempoImmutable {
		return TempoImmutable.today(options).setTimeFromTimeString(time);
	}

	static fromObject(components: TempoComponents): TempoImmutable {
		return TempoImmutable.create(components);
	}

	static fromTimestamp(timestamp: number, options?: TempoOptions): TempoImmutable {
		return new TempoImmutable(fromNumericTimestamp(timestamp), options);
	}

	static fromTimestampMs(timestamp: number, options?: TempoOptions): TempoImmutable {
		assertFiniteNumber(timestamp, 'Timestamp');

		return new TempoImmutable(new Date(timestamp), options);
	}

	static fromTimestampUTC(timestamp: number): TempoImmutable {
		return TempoImmutable.fromTimestamp(timestamp, {
			timeZone: defaultTimeZone,
		});
	}

	static fromTimestampMsUTC(timestamp: number): TempoImmutable {
		return TempoImmutable.fromTimestampMs(timestamp, {
			timeZone: defaultTimeZone,
		});
	}

	static min(...items: readonly TempoInput[]): TempoImmutable {
		if (items.length === 0) {
			throw new RangeError('Tempo.min requires at least one input');
		}

		return items.map((item) => TempoImmutable.parse(item)).reduce((earliest, item) => (item.isBefore(earliest) ? item : earliest));
	}

	static max(...items: readonly TempoInput[]): TempoImmutable {
		if (items.length === 0) {
			throw new RangeError('Tempo.max requires at least one input');
		}

		return items.map((item) => TempoImmutable.parse(item)).reduce((latest, item) => (item.isAfter(latest) ? item : latest));
	}

	static average(start: TempoInput, end: TempoInput): TempoImmutable {
		const startTempo = TempoImmutable.parse(start);

		const endTempo = TempoImmutable.parse(end, {
			timeZone: startTempo.timeZone,
		});

		return new TempoImmutable(new Date(averageMilliseconds(startTempo.timestampMs, endTempo.timestampMs)), { timeZone: startTempo.timeZone });
	}

	protected make(value: Date, timeZone = this.zone, locale = this.currentLocale, runtime = this.runtime.with({ locale })): this {
		const Constructor = this.constructor as new (input: TempoInput, options?: TempoOptions) => this;

		return new Constructor(value, optionsFromPolicy(this.policy, { locale, runtime, timeZone }));
	}

	get timeZone(): string {
		return this.zone;
	}

	getLocale(): string {
		return this.currentLocale;
	}

	getRuntime(): TempoRuntime {
		return this.runtime;
	}

	withRuntime(runtime: TempoRuntime): this {
		return this.make(this.value, this.zone, runtime.locale, runtime);
	}

	translator(): TempoTranslator {
		return this.runtime.translatorState();
	}

	withTranslator(translator: TempoTranslator): this {
		const runtime = this.runtime.with({
			fallbackLocale: translator.fallbackLocale,
			locale: translator.locale ?? this.currentLocale,
			translator,
		});

		return this.withRuntime(runtime);
	}

	hasTranslator(): boolean {
		return this.runtime.hasTranslator();
	}

	locale(): string;

	locale(locale: string): this;

	locale(locale?: string): string | this {
		return locale === undefined ? this.currentLocale : this.make(this.value, this.zone, locale, this.runtime.with({ locale }));
	}

	getSettings(): TempoSettings {
		return policyToSettings(this.policy);
	}

	policySnapshot(): TempoPolicy {
		return cloneTempoPolicy(this.policy);
	}

	settings(settings?: TempoSettings): TempoSettings | this {
		if (settings === undefined) {
			return this.getSettings();
		}

		return this.make(
			this.value,
			settings.timeZone === undefined ? this.zone : normalizeTimeZone(settings.timeZone),
			settings.locale ?? this.currentLocale,
			this.runtime.with({
				fallbackLocale: settings.fallbackLocale,
				locale: settings.locale ?? this.currentLocale,
			}),
		);
	}

	get timestamp(): number {
		return Math.trunc(this.value.getTime() / millisecondsPerSecond);
	}

	get timestampMs(): number {
		return this.value.getTime();
	}

	get year(): number {
		return this.parts().year;
	}

	get month(): number {
		return this.parts().month;
	}

	get quarter(): number {
		return Math.trunc((this.month - 1) / 3) + 1;
	}

	get day(): number {
		return this.parts().day;
	}

	get dayOfWeek(): number {
		return this.parts().weekday;
	}

	get isoWeekday(): number {
		return this.dayOfWeek === 0 ? 7 : this.dayOfWeek;
	}

	get isoWeek(): number {
		return isoWeekData(this.parts()).week;
	}

	get isoWeekYear(): number {
		return isoWeekData(this.parts()).year;
	}

	get weeksInISOYear(): number {
		return weeksInISOYear(this.isoWeekYear);
	}

	get isoWeeksInYear(): number {
		return this.weeksInISOYear;
	}

	get dayOfYear(): number {
		return this.diffInDays(this.startOf('year')) + 1;
	}

	get hour(): number {
		return this.parts().hour;
	}

	get minute(): number {
		return this.parts().minute;
	}

	get second(): number {
		return this.parts().second;
	}

	get millisecond(): number {
		return this.parts().millisecond;
	}

	private fieldValue(field: keyof TempoObject | BoundaryUnit | 'timestamp' | 'timestampMs'): unknown {
		switch (field) {
			case 'timestamp':
				return this.timestamp;

			case 'timestampMs':
				return this.timestampMs;

			case 'millisecond':

			case 'second':

			case 'minute':

			case 'hour':

			case 'day':

			case 'month':

			case 'year':
				return this[field];

			default:
				return this.toObject()[field as keyof TempoObject];
		}
	}

	paddedUnit(unit: keyof TempoObject | BoundaryUnit, length = 2): string {
		return pad(Number(this.fieldValue(unit)), length);
	}

	get offsetMinutes(): number {
		const parts = this.parts();
		const localAsUTC = dateFromPartsAsUTC(parts);

		return Math.trunc((localAsUTC - this.value.getTime()) / millisecondsPerMinute);
	}

	offsetString(separator: ':' | '' = ':'): string {
		return formatOffset(this.offsetMinutes, separator);
	}

	utcOffset(): number {
		return this.offsetMinutes;
	}

	timezoneName(style: TimeZoneNameStyle = 'short', locale = this.currentLocale): string {
		return timeZoneName(this.value, this.zone, style, locale);
	}

	monthName(locale = this.currentLocale): string {
		return monthNames(locale, 'long')[this.month - 1] ?? '';
	}

	shortMonthName(locale = this.currentLocale): string {
		return monthNames(locale, 'short')[this.month - 1] ?? '';
	}

	dayName(locale = this.currentLocale): string {
		return weekdayNames(locale, 'long')[this.dayOfWeek] ?? '';
	}

	shortDayName(locale = this.currentLocale): string {
		return weekdayNames(locale, 'short')[this.dayOfWeek] ?? '';
	}

	minDayName(locale = this.currentLocale): string {
		return this.shortDayName(locale).slice(0, 2);
	}

	translateNumber(value: number, locale = this.currentLocale): string {
		return new Intl.NumberFormat(locale).format(value);
	}

	translate(message: string, replacements: Record<string, string> = {}): string {
		const translated = this.runtime.translate(message, replacements);

		return typeof translated === 'string' ? translated : this.translateWith(message, replacements);
	}

	translateWith(message: string, replacements: Record<string, string> = {}): string {
		return replaceTranslationTokens(message, replacements);
	}

	translationMessage(key: string): string | number | null {
		return this.runtime.getMessage(key);
	}

	translationMessageWith(key: string, replacements: Record<string, string> = {}): string | number | null {
		const message = this.translationMessage(key);

		return typeof message === 'string' ? this.translateWith(message, replacements) : message;
	}

	translateTimeString(input: string, locale = 'en-US', targetLocale = this.currentLocale): string {
		let output = input;

		const sourceNames = [...monthNames(locale, 'long'), ...monthNames(locale, 'short'), ...weekdayNames(locale, 'long'), ...weekdayNames(locale, 'short')];
		const targetNames = [...monthNames(targetLocale, 'long'), ...monthNames(targetLocale, 'short'), ...weekdayNames(targetLocale, 'long'), ...weekdayNames(targetLocale, 'short')];

		sourceNames.forEach((sourceName, index) => {
			output = output.replaceAll(sourceName, targetNames[index] ?? sourceName);
		});

		return output;
	}

	translateTimeStringTo(input: string, targetLocale: string): string {
		return this.translateTimeString(input, this.currentLocale, targetLocale);
	}

	isUtc(): boolean {
		return this.zone === defaultTimeZone && this.offsetMinutes === 0;
	}

	isLocal(): boolean {
		return this.zone === Intl.DateTimeFormat().resolvedOptions().timeZone;
	}

	isDST(): boolean {
		const parts = this.parts();

		const januaryOffset = this.offsetForDate(
			dateFromZonedComponents({
				day: 1,
				month: 1,
				timeZone: this.zone,
				year: parts.year,
			}),
		);

		const julyOffset = this.offsetForDate(
			dateFromZonedComponents({
				day: 1,
				month: 7,
				timeZone: this.zone,
				year: parts.year,
			}),
		);

		const standardOffset = Math.min(januaryOffset, julyOffset);

		return this.offsetMinutes > standardOffset;
	}

	isLeapYear(): boolean {
		const year = this.year;

		return year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
	}

	daysInYear(): number {
		return this.isLeapYear() ? 366 : 365;
	}

	isLongYear(): boolean {
		return this.weeksInISOYear === 53;
	}

	isLongIsoYear(): boolean {
		return this.isLongYear();
	}

	isLastOfMonth(): boolean {
		return this.day === this.daysInMonth();
	}

	daysInMonth(): number {
		return daysInMonth(this.year, this.month);
	}

	isWeekend(): boolean {
		return this.policy.weekendDays.includes(this.dayOfWeek);
	}

	isSunday(): boolean {
		return this.dayOfWeek === 0;
	}

	isMonday(): boolean {
		return this.dayOfWeek === 1;
	}

	isTuesday(): boolean {
		return this.dayOfWeek === 2;
	}

	isWednesday(): boolean {
		return this.dayOfWeek === 3;
	}

	isThursday(): boolean {
		return this.dayOfWeek === 4;
	}

	isFriday(): boolean {
		return this.dayOfWeek === 5;
	}

	isSaturday(): boolean {
		return this.dayOfWeek === 6;
	}

	isDayOfWeek(weekday: WeekdayInput): boolean {
		return this.dayOfWeek === resolveWeekday(weekday);
	}

	isWeekday(): boolean {
		return !this.isWeekend();
	}

	isPast(reference: TempoInput = new Date()): boolean {
		return this.isBefore(reference);
	}

	isFuture(reference: TempoInput = new Date()): boolean {
		return this.isAfter(reference);
	}

	isNowOrPast(reference: TempoInput = new Date()): boolean {
		return this.isSameOrBefore(reference);
	}

	isNowOrFuture(reference: TempoInput = new Date()): boolean {
		return this.isSameOrAfter(reference);
	}

	isToday(reference: TempoInput = new Date()): boolean {
		return this.isSame(reference, 'day');
	}

	isTomorrow(reference: TempoInput = new Date()): boolean {
		return this.isSame(TempoImmutable.parse(reference, { timeZone: this.zone }).addDays(1), 'day');
	}

	isYesterday(reference: TempoInput = new Date()): boolean {
		return this.isSame(TempoImmutable.parse(reference, { timeZone: this.zone }).subDays(1), 'day');
	}

	isMidnight(): boolean {
		return this.hour === 0 && this.minute === 0 && this.second === 0 && this.millisecond === 0;
	}

	isMidday(): boolean {
		return this.hour === this.policy.midDayAt && this.minute === 0 && this.second === 0 && this.millisecond === 0;
	}

	clone(): this {
		return this.make(this.value);
	}

	copy(): this {
		return this.clone();
	}

	avoidMutation(): this {
		return this.clone();
	}

	cast(): this {
		return this.clone();
	}

	tempoize(input: TempoInput): this {
		const value = TempoImmutable.parse(input, { timeZone: this.zone });

		return this.make(value.toDate(), value.timeZone);
	}

	nowWithSameTz(): this {
		return this.make(new Date(), this.zone);
	}

	modify(modifier: string): this {
		const value = modifier.trim();

		if (value === '') {
			throw new RangeError('Tempo modifier cannot be empty');
		}

		try {
			const parsed = TempoImmutable.parse(value, { timeZone: this.zone });

			return this.make(parsed.toDate(), parsed.timeZone);
		} catch {
			// Continue with relative modifier parsing below.
		}

		const relative = value.match(/^([+-]?\d+(?:\.\d+)?)\s*(milliseconds?|seconds?|minutes?|hours?|days?|weeks?|months?|quarters?|years?|decades?|centuries?|millenniums?|millennia)$/i);

		if (relative !== null) {
			return this.add(Number(relative[1]), relative[2].toLowerCase() as TimeUnit);
		}

		const directional = value.match(/^(next|last|previous)\s+(milliseconds?|seconds?|minutes?|hours?|days?|weeks?|months?|quarters?|years?|decades?|centuries?|millenniums?|millennia)$/i);

		if (directional !== null) {
			const amount = directional[1].toLowerCase() === 'next' ? 1 : -1;

			return this.add(amount, directional[2].toLowerCase() as TimeUnit);
		}

		switch (value.toLowerCase()) {
			case 'now':
				return this.nowWithSameTz();

			case 'today':
				return this.make(new Date(), this.zone).startOfDay();

			case 'tomorrow':
				return this.make(new Date(), this.zone).startOfDay().addDays(1);

			case 'yesterday':
				return this.make(new Date(), this.zone).startOfDay().subDays(1);
		}

		throw new RangeError(`Invalid Tempo modifier: ${modifier}`);
	}

	change(modifier: string): this {
		return this.modify(modifier);
	}

	toImmutable(): TempoImmutable {
		return new TempoImmutable(this.value, optionsFromPolicy(this.policy));
	}

	timezone(timeZone: string): this {
		return this.setTimezone(timeZone);
	}

	tz(timeZone: string): this {
		return this.timezone(timeZone);
	}

	setTimezone(timeZone: string, keepLocalTime = false): this {
		const nextZone = normalizeTimeZone(timeZone);

		if (!keepLocalTime) {
			return this.make(this.value, nextZone);
		}

		return this.make(dateFromZonedComponents({ ...this.toObject(), timeZone: nextZone }), nextZone);
	}

	shiftTimezone(timeZone: string): this {
		return this.setTimezone(timeZone, true);
	}

	utc(): this {
		return this.setTimezone(defaultTimeZone);
	}

	local(): this {
		return this.setTimezone(Intl.DateTimeFormat().resolvedOptions().timeZone);
	}

	set(components: TempoSettableComponents): this {
		const timeZone = normalizeTimeZone(components.timeZone ?? this.zone);

		return this.make(
			dateFromZonedComponents(
				{
					...this.toObject(),
					...components,
					timeZone,
				},
				timeZone,
			),
			timeZone,
		);
	}

	setUnit(unit: TimeUnit, value: number): this {
		switch (normalizeUnit(unit)) {
			case 'year':
				return this.setYear(value);

			case 'month':
				return this.setMonth(value);

			case 'day':
				return this.setDay(value);

			case 'hour':
				return this.setHour(value);

			case 'minute':
				return this.setMinute(value);

			case 'second':
				return this.setSecond(value);

			case 'millisecond':
				return this.setMillisecond(value);

			default:
				throw new RangeError(`Tempo cannot set unit: ${unit}`);
		}
	}

	yearTo(year: number): this {
		return this.set({ year });
	}

	monthTo(month: number): this {
		return this.set({ month });
	}

	dayTo(day: number): this {
		return this.set({ day });
	}

	hourTo(hour: number): this {
		return this.set({ hour });
	}

	minuteTo(minute: number): this {
		return this.set({ minute });
	}

	secondTo(second: number): this {
		return this.set({ second });
	}

	millisecondTo(millisecond: number): this {
		return this.set({ millisecond });
	}

	setYear(year: number): this {
		return this.set({ year });
	}

	setMonth(month: number): this {
		return this.set({ month });
	}

	setDay(day: number): this {
		return this.set({ day });
	}

	setDate(year: number, month: number, day: number): this {
		return this.set({ day, month, year });
	}

	setDateFrom(source: TempoInput): this {
		const date = TempoImmutable.parse(source, { timeZone: this.zone });

		return this.setDate(date.year, date.month, date.day);
	}

	setDateTime(year: number, month: number, day: number, hour: number, minute: number, second = 0, millisecond = 0): this {
		return this.set({ day, hour, millisecond, minute, month, second, year });
	}

	setDateTimeFrom(source: TempoInput): this {
		const date = TempoImmutable.parse(source, { timeZone: this.zone });

		return this.setDateTime(date.year, date.month, date.day, date.hour, date.minute, date.second, date.millisecond);
	}

	setHour(hour: number): this {
		return this.set({ hour });
	}

	setMinute(minute: number): this {
		return this.set({ minute });
	}

	setSecond(second: number): this {
		return this.set({ second });
	}

	setMillisecond(millisecond: number): this {
		return this.set({ millisecond });
	}

	setTime(hour: number, minute = this.minute, second = this.second, millisecond = this.millisecond): this {
		return this.set({ hour, millisecond, minute, second });
	}

	setTimeFrom(source: TempoInput): this {
		const date = TempoImmutable.parse(source, { timeZone: this.zone });

		return this.setTime(date.hour, date.minute, date.second, date.millisecond);
	}

	setTimeFromTimeString(time: string): this {
		const parsed = TempoImmutable.parse(`${this.toDateString()}T${time}`, {
			timeZone: this.zone,
		});

		return this.setTime(parsed.hour, parsed.minute, parsed.second, parsed.millisecond);
	}

	setTimestamp(timestamp: number): this {
		return this.make(fromNumericTimestamp(timestamp), this.zone);
	}

	setISODate(year: number, week: number, day = 1): this {
		const isoYearStart = this.make(new Date(Date.UTC(year, 0, 4)), this.zone).startOfWeek({ weekStartsOn: 1 });

		return isoYearStart
			.addWeeks(week - 1)
			.addDays(day - 1)
			.setTime(this.hour, this.minute, this.second, this.millisecond);
	}

	setISOWeek(week: number, day = this.isoWeekday): this {
		return this.setISODate(this.isoWeekYear, week, day);
	}

	setISOWeekYear(year: number, day = this.isoWeekday): this {
		return this.setISODate(year, this.isoWeek, day);
	}

	setISOWeekday(value: WeekdayInput): this {
		const weekday = resolveWeekday(value);

		return this.setISODate(this.isoWeekYear, this.isoWeek, weekday === 0 ? 7 : weekday);
	}

	weekday(): number;

	weekday(value: WeekdayInput): this;

	weekday(value?: WeekdayInput): number | this {
		if (value === undefined) {
			return this.dayOfWeek;
		}

		return this.addDays(resolveWeekday(value) - this.dayOfWeek);
	}

	setWeekday(value: WeekdayInput): this {
		return this.weekday(value);
	}

	setDayOfYear(day: number): this {
		return this.startOfYear()
			.addDays(day - 1)
			.setTime(this.hour, this.minute, this.second, this.millisecond);
	}

	dayOfYearTo(day: number): this {
		return this.setDayOfYear(day);
	}

	setUnitNoOverflow(valueUnit: TimeUnit, value: number, overflowUnit: BoundaryUnit): this {
		return this.setUnit(valueUnit, value).clamp(this.startOf(overflowUnit), this.endOf(overflowUnit));
	}

	midday(): this {
		return this.setTime(this.policy.midDayAt, 0, 0, 0);
	}

	midDay(): this {
		return this.midday();
	}

	add(value: number, unit: TimeUnit): this {
		assertFiniteNumber(value, 'Amount');

		const fixed = fixedUnitMilliseconds(unit);

		if (fixed !== null) {
			return this.make(new Date(this.value.getTime() + value * fixed));
		}

		switch (normalizeUnit(unit)) {
			case 'month':
				return this.addMonths(value);

			case 'quarter':
				return this.addMonths(value * 3);

			case 'year':
				return this.addYears(value);

			default:
				return this.make(this.value);
		}
	}

	sub(value: number, unit: TimeUnit): this {
		return this.add(-value, unit);
	}

	addNoOverflow(value: number, valueUnit: TimeUnit, overflowUnit: BoundaryUnit): this {
		return this.add(value, valueUnit).clamp(this.startOf(overflowUnit), this.endOf(overflowUnit));
	}

	subNoOverflow(value: number, valueUnit: TimeUnit, overflowUnit: BoundaryUnit): this {
		return this.addNoOverflow(-value, valueUnit, overflowUnit);
	}

	addDuration(duration: DurationInput): this {
		const value = durationFromInput(duration);

		return this.addYears(value.years)
			.addMonths(value.quarters * 3 + value.months)
			.addWeeks(value.weeks)
			.addDays(value.days)
			.addHours(value.hours)
			.addMinutes(value.minutes)
			.addSeconds(value.seconds)
			.addMilliseconds(value.milliseconds);
	}

	subDuration(duration: DurationInput): this {
		return this.addDuration(durationFromInput(duration).negated());
	}

	addMilliseconds(milliseconds: number): this {
		return this.add(milliseconds, 'millisecond');
	}

	subMilliseconds(milliseconds: number): this {
		return this.sub(milliseconds, 'millisecond');
	}

	addSeconds(seconds: number): this {
		return this.add(seconds, 'second');
	}

	subSeconds(seconds: number): this {
		return this.sub(seconds, 'second');
	}

	addMinutes(minutes: number): this {
		return this.add(minutes, 'minute');
	}

	subMinutes(minutes: number): this {
		return this.sub(minutes, 'minute');
	}

	addHours(hours: number): this {
		return this.add(hours, 'hour');
	}

	subHours(hours: number): this {
		return this.sub(hours, 'hour');
	}

	addDays(days: number): this {
		return this.add(days, 'day');
	}

	subDays(days: number): this {
		return this.sub(days, 'day');
	}

	addWeekdays(days: number): this {
		assertFiniteNumber(days, 'Weekdays');

		if (days === 0) {
			return this.clone();
		}

		const direction = days < 0 ? -1 : 1;

		let remaining = Math.abs(Math.trunc(days));
		let current = this.clone();

		while (remaining > 0) {
			current = current.addDays(direction);

			if (current.isWeekday()) {
				remaining -= 1;
			}
		}

		return current;
	}

	subWeekdays(days: number): this {
		return this.addWeekdays(-days);
	}

	addWeeks(weeks: number): this {
		return this.add(weeks, 'week');
	}

	subWeeks(weeks: number): this {
		return this.sub(weeks, 'week');
	}

	addMonths(months: number): this {
		if (!this.policy.monthsOverflow) {
			return this.addMonthsNoOverflow(months);
		}

		assertFiniteNumber(months, 'Months');

		const parts = this.toObject();

		return this.make(
			dateFromZonedComponents(
				{
					...parts,
					month: parts.month + months,
				},
				this.zone,
			),
		);
	}

	addMonthsNoOverflow(months: number): this {
		assertFiniteNumber(months, 'Months');

		const parts = this.toObject();

		const firstOfTarget = dateFromZonedComponents(
			{
				...parts,
				day: 1,
				month: parts.month + months,
			},
			this.zone,
		);

		const target = getZonedParts(firstOfTarget, this.zone);
		const day = Math.min(parts.day, daysInMonth(target.year, target.month));

		return this.make(
			dateFromZonedComponents(
				{
					...parts,
					day,
					month: parts.month + months,
				},
				this.zone,
			),
		);
	}

	subMonths(months: number): this {
		return this.addMonths(-months);
	}

	subMonthsNoOverflow(months: number): this {
		return this.addMonthsNoOverflow(-months);
	}

	addQuarters(quarters: number): this {
		return this.addMonths(quarters * 3);
	}

	subQuarters(quarters: number): this {
		return this.addQuarters(-quarters);
	}

	addYears(years: number): this {
		if (!this.policy.yearsOverflow) {
			return this.addYearsNoOverflow(years);
		}

		assertFiniteNumber(years, 'Years');

		const parts = this.toObject();

		return this.make(
			dateFromZonedComponents(
				{
					...parts,
					year: parts.year + years,
				},
				this.zone,
			),
		);
	}

	addYearsNoOverflow(years: number): this {
		const parts = this.toObject();

		const day = Math.min(parts.day, daysInMonth(parts.year + years, parts.month));

		return this.set({ day, year: parts.year + years });
	}

	age(reference: TempoInput = new Date()): number {
		return TempoImmutable.parse(reference, { timeZone: this.zone }).diffInYears(this);
	}

	subYears(years: number): this {
		return this.addYears(-years);
	}

	subYearsNoOverflow(years: number): this {
		return this.addYearsNoOverflow(-years);
	}

	startOf(unit: BoundaryUnit, options?: StartOfWeekOptions): this {
		const parts = this.toObject();

		switch (unit) {
			case 'millisecond':
				return this.clone();

			case 'second':
				return this.set({ millisecond: 0 });

			case 'minute':
				return this.set({ millisecond: 0, second: 0 });

			case 'hour':
				return this.set({ millisecond: 0, minute: 0, second: 0 });

			case 'day':
				return this.set({ hour: 0, millisecond: 0, minute: 0, second: 0 });

			case 'week': {
				const weekStartsOn = options?.weekStartsOn ?? 1;
				const delta = (parts.weekday - weekStartsOn + 7) % 7;

				return this.startOf('day').subDays(delta);
			}

			case 'month':
				return this.set({
					day: 1,
					hour: 0,
					millisecond: 0,
					minute: 0,
					second: 0,
				});

			case 'quarter':
				return this.set({
					day: 1,
					hour: 0,
					millisecond: 0,
					minute: 0,
					month: (this.quarter - 1) * 3 + 1,
					second: 0,
				});

			case 'year':
				return this.set({
					day: 1,
					hour: 0,
					millisecond: 0,
					minute: 0,
					month: 1,
					second: 0,
				});

			case 'decade':
				return this.setDate(parts.year - (parts.year % 10), 1, 1).startOfDay();

			case 'century':
				return this.setDate(parts.year - ((parts.year - 1) % 100), 1, 1).startOfDay();

			case 'millennium':
				return this.setDate(parts.year - ((parts.year - 1) % 1000), 1, 1).startOfDay();
		}
	}

	endOf(unit: BoundaryUnit, options?: StartOfWeekOptions): this {
		switch (unit) {
			case 'millisecond':
				return this.clone();

			case 'second':
				return this.startOf('second').addSeconds(1).subMilliseconds(1);

			case 'minute':
				return this.startOf('minute').addMinutes(1).subMilliseconds(1);

			case 'hour':
				return this.startOf('hour').addHours(1).subMilliseconds(1);

			case 'day':
				return this.startOf('day').addDays(1).subMilliseconds(1);

			case 'week':
				return this.startOf('week', options).addWeeks(1).subMilliseconds(1);

			case 'month':
				return this.startOf('month').addMonths(1).subMilliseconds(1);

			case 'quarter':
				return this.startOf('quarter').addQuarters(1).subMilliseconds(1);

			case 'year':
				return this.startOf('year').addYears(1).subMilliseconds(1);

			case 'decade':
				return this.startOf('decade').addYears(10).subMilliseconds(1);

			case 'century':
				return this.startOf('century').addYears(100).subMilliseconds(1);

			case 'millennium':
				return this.startOf('millennium').addYears(1000).subMilliseconds(1);
		}
	}

	isStartOf(unit: BoundaryUnit, options?: StartOfWeekOptions): boolean {
		return this.isSame(this.startOf(unit, options));
	}

	isStartOfUnit(unit: BoundaryUnit, options?: StartOfWeekOptions): boolean {
		return this.isStartOf(unit, options);
	}

	isEndOf(unit: BoundaryUnit, options?: StartOfWeekOptions): boolean {
		return this.isSame(this.endOf(unit, options));
	}

	isEndOfUnit(unit: BoundaryUnit, options?: StartOfWeekOptions): boolean {
		return this.isEndOf(unit, options);
	}

	isStartOfTime(): boolean {
		return this.timestampMs <= -8640000000000000;
	}

	isEndOfTime(): boolean {
		return this.timestampMs >= 8640000000000000;
	}

	isCurrentUnit(unit: ComparisonUnit, reference: TempoInput = new Date()): boolean {
		return this.isSame(reference, unit);
	}

	isStartOfMillisecond(): boolean {
		return this.isStartOf('millisecond');
	}

	isEndOfMillisecond(): boolean {
		return this.isEndOf('millisecond');
	}

	isStartOfSecond(): boolean {
		return this.isStartOf('second');
	}

	isEndOfSecond(): boolean {
		return this.isEndOf('second');
	}

	isStartOfMinute(): boolean {
		return this.isStartOf('minute');
	}

	isEndOfMinute(): boolean {
		return this.isEndOf('minute');
	}

	isStartOfHour(): boolean {
		return this.isStartOf('hour');
	}

	isEndOfHour(): boolean {
		return this.isEndOf('hour');
	}

	startOfMillisecond(): this {
		return this.startOf('millisecond');
	}

	endOfMillisecond(): this {
		return this.endOf('millisecond');
	}

	startOfSecond(): this {
		return this.startOf('second');
	}

	endOfSecond(): this {
		return this.endOf('second');
	}

	startOfMinute(): this {
		return this.startOf('minute');
	}

	endOfMinute(): this {
		return this.endOf('minute');
	}

	startOfHour(): this {
		return this.startOf('hour');
	}

	endOfHour(): this {
		return this.endOf('hour');
	}

	isStartOfDay(): boolean {
		return this.isStartOf('day');
	}

	isEndOfDay(): boolean {
		return this.isEndOf('day');
	}

	isStartOfWeek(options?: StartOfWeekOptions): boolean {
		return this.isStartOf('week', options);
	}

	isEndOfWeek(options?: StartOfWeekOptions): boolean {
		return this.isEndOf('week', options);
	}

	isStartOfMonth(): boolean {
		return this.isStartOf('month');
	}

	isEndOfMonth(): boolean {
		return this.isEndOf('month');
	}

	isStartOfQuarter(): boolean {
		return this.isStartOf('quarter');
	}

	isEndOfQuarter(): boolean {
		return this.isEndOf('quarter');
	}

	isStartOfYear(): boolean {
		return this.isStartOf('year');
	}

	isEndOfYear(): boolean {
		return this.isEndOf('year');
	}

	isStartOfDecade(): boolean {
		return this.isStartOf('decade');
	}

	isEndOfDecade(): boolean {
		return this.isEndOf('decade');
	}

	isStartOfCentury(): boolean {
		return this.isStartOf('century');
	}

	isEndOfCentury(): boolean {
		return this.isEndOf('century');
	}

	isStartOfMillennium(): boolean {
		return this.isStartOf('millennium');
	}

	isEndOfMillennium(): boolean {
		return this.isEndOf('millennium');
	}

	startOfDay(): this {
		return this.startOf('day');
	}

	endOfDay(): this {
		return this.endOf('day');
	}

	startOfWeek(options?: StartOfWeekOptions): this {
		return this.startOf('week', options);
	}

	endOfWeek(options?: StartOfWeekOptions): this {
		return this.endOf('week', options);
	}

	startOfMonth(): this {
		return this.startOf('month');
	}

	endOfMonth(): this {
		return this.endOf('month');
	}

	startOfQuarter(): this {
		return this.startOf('quarter');
	}

	endOfQuarter(): this {
		return this.endOf('quarter');
	}

	firstOfMonth(weekday?: WeekdayInput): this {
		const first = this.startOf('month');

		if (weekday === undefined) {
			return first;
		}

		const target = resolveWeekday(weekday);
		const delta = (target - first.dayOfWeek + 7) % 7;

		return first.addDays(delta);
	}

	lastOfMonth(weekday?: WeekdayInput): this {
		const last = this.endOf('month').startOf('day');

		if (weekday === undefined) {
			return last;
		}

		const target = resolveWeekday(weekday);
		const delta = (last.dayOfWeek - target + 7) % 7;

		return last.subDays(delta);
	}

	nthOfMonth(occurrence: number, weekday: WeekdayInput): this | null {
		if (!Number.isInteger(occurrence) || occurrence === 0) {
			throw new RangeError('Tempo nthOfMonth occurrence must be a non-zero integer');
		}

		const currentMonth = this.month;

		const candidate = occurrence > 0 ? this.firstOfMonth(weekday).addWeeks(occurrence - 1) : this.lastOfMonth(weekday).subWeeks(Math.abs(occurrence) - 1);

		return candidate.month === currentMonth ? candidate : null;
	}

	firstOfQuarter(weekday?: WeekdayInput): this {
		const first = this.startOf('quarter');

		if (weekday === undefined) {
			return first;
		}

		const target = resolveWeekday(weekday);
		const delta = (target - first.dayOfWeek + 7) % 7;

		return first.addDays(delta);
	}

	lastOfQuarter(weekday?: WeekdayInput): this {
		const last = this.endOf('quarter').startOf('day');

		if (weekday === undefined) {
			return last;
		}

		const target = resolveWeekday(weekday);
		const delta = (last.dayOfWeek - target + 7) % 7;

		return last.subDays(delta);
	}

	nthOfQuarter(occurrence: number, weekday: WeekdayInput): this | null {
		if (!Number.isInteger(occurrence) || occurrence === 0) {
			throw new RangeError('Tempo nthOfQuarter occurrence must be a non-zero integer');
		}

		const currentQuarter = this.quarter;
		const currentYear = this.year;

		const candidate = occurrence > 0 ? this.firstOfQuarter(weekday).addWeeks(occurrence - 1) : this.lastOfQuarter(weekday).subWeeks(Math.abs(occurrence) - 1);

		return candidate.quarter === currentQuarter && candidate.year === currentYear ? candidate : null;
	}

	firstOfYear(weekday?: WeekdayInput): this {
		const first = this.startOf('year');

		if (weekday === undefined) {
			return first;
		}

		const target = resolveWeekday(weekday);
		const delta = (target - first.dayOfWeek + 7) % 7;

		return first.addDays(delta);
	}

	lastOfYear(weekday?: WeekdayInput): this {
		const last = this.endOf('year').startOf('day');

		if (weekday === undefined) {
			return last;
		}

		const target = resolveWeekday(weekday);
		const delta = (last.dayOfWeek - target + 7) % 7;

		return last.subDays(delta);
	}

	nthOfYear(occurrence: number, weekday: WeekdayInput): this | null {
		if (!Number.isInteger(occurrence) || occurrence === 0) {
			throw new RangeError('Tempo nthOfYear occurrence must be a non-zero integer');
		}

		const currentYear = this.year;

		const candidate = occurrence > 0 ? this.firstOfYear(weekday).addWeeks(occurrence - 1) : this.lastOfYear(weekday).subWeeks(Math.abs(occurrence) - 1);

		return candidate.year === currentYear ? candidate : null;
	}

	startOfYear(): this {
		return this.startOf('year');
	}

	endOfYear(): this {
		return this.endOf('year');
	}

	startOfDecade(): this {
		return this.startOf('decade');
	}

	endOfDecade(): this {
		return this.endOf('decade');
	}

	startOfCentury(): this {
		return this.startOf('century');
	}

	endOfCentury(): this {
		return this.endOf('century');
	}

	startOfMillennium(): this {
		return this.startOf('millennium');
	}

	endOfMillennium(): this {
		return this.endOf('millennium');
	}

	floor(unit: TimeUnit): this {
		const fixed = fixedUnitMilliseconds(unit);

		if (fixed === null) {
			return this.startOf(normalizeUnit(unit) as BoundaryUnit);
		}

		return this.make(new Date(Math.floor(this.timestampMs / fixed) * fixed));
	}

	floorUnit(unit: TimeUnit): this {
		return this.floor(unit);
	}

	floorWeek(options?: StartOfWeekOptions): this {
		return this.startOfWeek(options);
	}

	ceil(unit: TimeUnit): this {
		const floored = this.floor(unit);

		if (floored.isSame(this)) {
			return floored;
		}

		return floored.add(1, unit);
	}

	ceilUnit(unit: TimeUnit): this {
		return this.ceil(unit);
	}

	ceilWeek(options?: StartOfWeekOptions): this {
		const floored = this.floorWeek(options);

		return floored.isSame(this) ? floored : floored.addWeeks(1);
	}

	round(unit: TimeUnit): this {
		const fixed = fixedUnitMilliseconds(unit);

		if (fixed === null) {
			const start = this.startOf(normalizeUnit(unit) as BoundaryUnit);
			const end = this.endOf(normalizeUnit(unit) as BoundaryUnit);

			const midpoint = start.timestampMs + (end.timestampMs - start.timestampMs) / 2;

			return this.timestampMs >= midpoint ? this.ceil(unit) : start;
		}

		return this.make(new Date(Math.round(this.timestampMs / fixed) * fixed));
	}

	roundUnit(unit: TimeUnit): this {
		return this.round(unit);
	}

	roundWeek(options?: StartOfWeekOptions): this {
		const start = this.startOfWeek(options);
		const end = this.endOfWeek(options);

		const midpoint = start.timestampMs + (end.timestampMs - start.timestampMs) / 2;

		return this.timestampMs >= midpoint ? this.ceilWeek(options) : start;
	}

	next(weekday: WeekdayInput): this {
		const target = resolveWeekday(weekday);
		const delta = (target - this.dayOfWeek + 7) % 7 || 7;

		return this.addDays(delta);
	}

	previous(weekday: WeekdayInput): this {
		const target = resolveWeekday(weekday);
		const delta = (this.dayOfWeek - target + 7) % 7 || 7;

		return this.subDays(delta);
	}

	nextOrSame(weekday: WeekdayInput): this {
		return this.dayOfWeek === resolveWeekday(weekday) ? this.clone() : this.next(weekday);
	}

	previousOrSame(weekday: WeekdayInput): this {
		return this.dayOfWeek === resolveWeekday(weekday) ? this.clone() : this.previous(weekday);
	}

	nextWeekday(): this {
		let next = this.addDays(1);

		while (next.isWeekend()) {
			next = next.addDays(1);
		}

		return next;
	}

	previousWeekday(): this {
		let previous = this.subDays(1);

		while (previous.isWeekend()) {
			previous = previous.subDays(1);
		}

		return previous;
	}

	nextWeekendDay(): this {
		let next = this.addDays(1);

		while (next.isWeekday()) {
			next = next.addDays(1);
		}

		return next;
	}

	previousWeekendDay(): this {
		let previous = this.subDays(1);

		while (previous.isWeekday()) {
			previous = previous.subDays(1);
		}

		return previous;
	}

	diff(other: TempoInput, unit: TimeUnit = 'millisecond', options?: DiffOptions): number {
		const otherDate = asDate(other);
		const rawMilliseconds = this.timestampMs - otherDate.getTime();

		const milliseconds = options?.absolute ? Math.abs(rawMilliseconds) : rawMilliseconds;

		const fixed = fixedUnitMilliseconds(unit);

		if (fixed !== null) {
			const value = milliseconds / fixed;

			return options?.float ? value : Math.trunc(value);
		}

		const otherTempo = TempoImmutable.parse(other, { timeZone: this.zone });
		const sign = milliseconds < 0 ? -1 : 1;
		const start = sign < 0 ? this : otherTempo;
		const end = sign < 0 ? otherTempo : this;
		const startParts = start.toObject();
		const endParts = end.toObject();

		let months = (endParts.year - startParts.year) * 12 + (endParts.month - startParts.month);

		if (endParts.day < startParts.day) {
			months -= 1;
		}

		const result = normalizeUnit(unit) === 'year' ? months / 12 : normalizeUnit(unit) === 'quarter' ? months / 3 : months;

		const signed = options?.absolute ? Math.abs(result) : result * sign;

		return options?.float ? signed : Math.trunc(signed);
	}

	diffAsDuration(other: TempoInput, options?: DiffOptions): TempoDuration {
		return new TempoDuration({
			milliseconds: this.diffInMilliseconds(other, options),
		}).normalized();
	}

	diffAsDateInterval(other: TempoInput, options?: DiffOptions): TempoDuration {
		return this.diffAsDuration(other, options);
	}

	diffAsTempoInterval(other: TempoInput, options?: DiffOptions): TempoDuration {
		return this.diffAsDuration(other, options);
	}

	diffInMilliseconds(other: TempoInput, options?: DiffOptions): number {
		return this.diff(other, 'millisecond', options);
	}

	diffInMicroseconds(other: TempoInput, options?: DiffOptions): number {
		return this.diffInMilliseconds(other, options) * 1000;
	}

	diffInSeconds(other: TempoInput, options?: DiffOptions): number {
		return this.diff(other, 'second', options);
	}

	diffInMinutes(other: TempoInput, options?: DiffOptions): number {
		return this.diff(other, 'minute', options);
	}

	diffInHours(other: TempoInput, options?: DiffOptions): number {
		return this.diff(other, 'hour', options);
	}

	diffInDays(other: TempoInput, options?: DiffOptions): number {
		return this.diff(other, 'day', options);
	}

	diffInWeeks(other: TempoInput, options?: DiffOptions): number {
		return this.diff(other, 'week', options);
	}

	diffInWeekdays(other: TempoInput, options?: DiffOptions): number {
		return this.diffFilteredDays(other, (item) => item.isWeekday(), options);
	}

	diffInWeekendDays(other: TempoInput, options?: DiffOptions): number {
		return this.diffFilteredDays(other, (item) => item.isWeekend(), options);
	}

	diffInMonths(other: TempoInput, options?: DiffOptions): number {
		return this.diff(other, 'month', options);
	}

	diffInQuarters(other: TempoInput, options?: DiffOptions): number {
		return this.diff(other, 'quarter', options);
	}

	diffInYears(other: TempoInput, options?: DiffOptions): number {
		return this.diff(other, 'year', options);
	}

	diffInUnit(unit: TimeUnit, other: TempoInput, options?: DiffOptions): number {
		return this.diff(other, unit, options);
	}

	diffInDaysFiltered(predicate: (item: TempoImmutable) => boolean, other: TempoInput, options?: DiffOptions): number {
		return this.diffFilteredDays(other, predicate, options);
	}

	diffFiltered(predicate: (item: TempoImmutable) => boolean, other: TempoInput, options?: DiffOptions): number {
		return this.diffInDaysFiltered(predicate, other, options);
	}

	diffInHoursFiltered(predicate: (item: TempoImmutable) => boolean, other: TempoInput, options?: DiffOptions): number {
		const otherTempo = TempoImmutable.parse(other, { timeZone: this.zone });
		const sign = this.isBefore(otherTempo, 'hour') ? -1 : 1;
		const start = sign < 0 ? this.startOf('hour') : otherTempo.startOf('hour');
		const end = sign < 0 ? otherTempo.startOf('hour') : this.startOf('hour');

		let current = start;
		let count = 0;

		while (current.isBefore(end, 'hour')) {
			current = current.addHours(1);

			if (current.isSameOrBefore(end, 'hour') && predicate(current)) {
				count += 1;
			}
		}

		const result = options?.absolute ? count : count * sign;

		return options?.float ? result : Math.trunc(result);
	}

	secondsSinceMidnight(): number {
		return this.diffInSeconds(this.startOfDay(), { absolute: true });
	}

	secondsUntilEndOfDay(): number {
		return this.diffInSeconds(this.endOfDay(), { absolute: true });
	}

	calendar(referenceTime: TempoInput = new Date(), formats: CalendarFormats = {}): string {
		const reference = TempoImmutable.parse(referenceTime, {
			timeZone: this.zone,
		});

		const diff = this.startOfDay().diffInDays(reference.startOfDay());

		const key: CalendarFormatKey = diff === 0 ? 'sameDay' : diff === 1 ? 'nextDay' : diff > 1 && diff < 7 ? 'nextWeek' : diff === -1 ? 'lastDay' : diff < -1 && diff > -7 ? 'lastWeek' : 'sameElse';

		const defaults = calendarFormatDefaults();

		return this.format(formats[key] ?? defaults[key]);
	}

	diffForHumans(other: TempoInput = new Date(), options?: HumanDiffOptions): string {
		const resolvedOptions = {
			...this.policy.humanDiffOptions,
			...options,
		};

		const rawMilliseconds = this.timestampMs - asDate(other).getTime();
		const unit = resolvedOptions.unit ?? bestRelativeUnit(rawMilliseconds);
		const value = Math.round(rawMilliseconds / unitDivisor(unit));

		const formatter = new Intl.RelativeTimeFormat(resolvedOptions.locale ?? this.currentLocale ?? this.runtime.locale, {
			numeric: resolvedOptions.numeric ?? 'always',
			style: resolvedOptions.style ?? 'long',
		});

		return formatter.format(resolvedOptions.absolute ? Math.abs(value) : value, unit);
	}

	from(other: TempoInput = new Date(), options?: HumanDiffOptions): string {
		return this.diffForHumans(other, options);
	}

	since(other: TempoInput = new Date(), options?: HumanDiffOptions): string {
		return this.from(other, options);
	}

	to(other: TempoInput = new Date(), options?: HumanDiffOptions): string {
		return TempoImmutable.parse(other, { timeZone: this.zone }).diffForHumans(this, options);
	}

	fromNow(options?: HumanDiffOptions): string {
		return this.diffForHumans(new Date(), options);
	}

	toNow(options?: HumanDiffOptions): string {
		return TempoImmutable.parse(new Date(), {
			timeZone: this.zone,
		}).diffForHumans(this, options);
	}

	ago(options?: HumanDiffOptions): string {
		return this.fromNow(options);
	}

	timespan(other: TempoInput = new Date(), options?: HumanDiffOptions): string {
		return this.diffForHumans(other, { ...options, absolute: true });
	}

	isImmutable(): boolean {
		return true;
	}

	isMutable(): boolean {
		return false;
	}

	isBefore(other: TempoInput, unit: ComparisonUnit = 'millisecond'): boolean {
		return this.comparableValue(unit) < TempoImmutable.parse(other, { timeZone: this.zone }).comparableValue(unit);
	}

	isAfter(other: TempoInput, unit: ComparisonUnit = 'millisecond'): boolean {
		return this.comparableValue(unit) > TempoImmutable.parse(other, { timeZone: this.zone }).comparableValue(unit);
	}

	isSame(other: TempoInput, unit: ComparisonUnit = 'millisecond'): boolean {
		return this.comparableValue(unit) === TempoImmutable.parse(other, { timeZone: this.zone }).comparableValue(unit);
	}

	is(other: TempoInput, unit: ComparisonUnit = 'day'): boolean {
		return this.isSame(other, unit);
	}

	equalTo(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.isSame(other, unit);
	}

	eq(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.equalTo(other, unit);
	}

	notEqualTo(other: TempoInput, unit?: ComparisonUnit): boolean {
		return !this.isSame(other, unit);
	}

	ne(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.notEqualTo(other, unit);
	}

	greaterThan(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.isAfter(other, unit);
	}

	gt(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.greaterThan(other, unit);
	}

	greaterThanOrEqualTo(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.isSameOrAfter(other, unit);
	}

	gte(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.greaterThanOrEqualTo(other, unit);
	}

	lessThan(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.isBefore(other, unit);
	}

	lt(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.lessThan(other, unit);
	}

	lessThanOrEqualTo(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.isSameOrBefore(other, unit);
	}

	lte(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.lessThanOrEqualTo(other, unit);
	}

	isSameSecond(other: TempoInput): boolean {
		return this.isSame(other, 'second');
	}

	isSameMinute(other: TempoInput): boolean {
		return this.isSame(other, 'minute');
	}

	isSameHour(other: TempoInput): boolean {
		return this.isSame(other, 'hour');
	}

	isSameDay(other: TempoInput): boolean {
		return this.isSame(other, 'day');
	}

	isSameWeek(other: TempoInput): boolean {
		return this.isSame(other, 'week');
	}

	isSameMonth(other: TempoInput): boolean {
		return this.isSame(other, 'month');
	}

	isSameQuarter(other: TempoInput): boolean {
		return this.isSame(other, 'quarter');
	}

	isSameYear(other: TempoInput): boolean {
		return this.isSame(other, 'year');
	}

	isSameAs(pattern: string, other: TempoInput, options?: FormatOptions): boolean {
		const compare = TempoImmutable.parse(other, { timeZone: this.zone });

		return this.format(pattern, options) === compare.format(pattern, options);
	}

	isSameUnit(unit: ComparisonUnit, other: TempoInput): boolean {
		return this.isSame(other, unit);
	}

	isBirthday(other: TempoInput = new Date()): boolean {
		const compare = TempoImmutable.parse(other, { timeZone: this.zone });

		return this.month === compare.month && this.day === compare.day;
	}

	isSameOrBefore(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.isSame(other, unit) || this.isBefore(other, unit);
	}

	isSameOrAfter(other: TempoInput, unit?: ComparisonUnit): boolean {
		return this.isSame(other, unit) || this.isAfter(other, unit);
	}

	isBetween(start: TempoInput, end: TempoInput, unit: ComparisonUnit = 'millisecond', inclusivity: '()' | '[]' | '[)' | '(]' = '[]'): boolean {
		const startTempo = TempoImmutable.parse(start, { timeZone: this.zone });
		const endTempo = TempoImmutable.parse(end, { timeZone: this.zone });

		const [lower, upper] = startTempo.isAfter(endTempo, unit) ? [endTempo, startTempo] : [startTempo, endTempo];

		const afterStart = inclusivity.startsWith('[') ? this.isSameOrAfter(lower, unit) : this.isAfter(lower, unit);

		const beforeEnd = inclusivity.endsWith(']') ? this.isSameOrBefore(upper, unit) : this.isBefore(upper, unit);

		return afterStart && beforeEnd;
	}

	between(start: TempoInput, end: TempoInput, equal = true, unit: ComparisonUnit = 'millisecond'): boolean {
		return this.isBetween(start, end, unit, equal ? '[]' : '()');
	}

	betweenIncluded(start: TempoInput, end: TempoInput, unit: ComparisonUnit = 'millisecond'): boolean {
		return this.isBetween(start, end, unit, '[]');
	}

	betweenExcluded(start: TempoInput, end: TempoInput, unit: ComparisonUnit = 'millisecond'): boolean {
		return this.isBetween(start, end, unit, '()');
	}

	clamp(min: TempoInput, max: TempoInput): this {
		const minTempo = TempoImmutable.parse(min, { timeZone: this.zone });
		const maxTempo = TempoImmutable.parse(max, { timeZone: this.zone });

		if (minTempo.isAfter(maxTempo)) {
			throw new RangeError('Tempo clamp minimum must be before maximum');
		}

		if (this.isBefore(minTempo)) {
			return this.make(minTempo.toDate(), minTempo.timeZone);
		}

		if (this.isAfter(maxTempo)) {
			return this.make(maxTempo.toDate(), maxTempo.timeZone);
		}

		return this.clone();
	}

	average(other: TempoInput): this {
		const end = TempoImmutable.parse(other, { timeZone: this.zone });

		return this.make(new Date(averageMilliseconds(this.timestampMs, end.timestampMs)));
	}

	closest(...items: readonly TempoInput[]): this {
		if (items.length === 0) {
			throw new RangeError('Tempo.closest requires at least one input');
		}

		const closest = items
			.map((item) => TempoImmutable.parse(item, { timeZone: this.zone }))
			.reduce((best, item) => (millisecondDistance(item.timestampMs, this.timestampMs) < millisecondDistance(best.timestampMs, this.timestampMs) ? item : best));

		return this.make(closest.toDate(), closest.timeZone);
	}

	farthest(...items: readonly TempoInput[]): this {
		if (items.length === 0) {
			throw new RangeError('Tempo.farthest requires at least one input');
		}

		const farthest = items
			.map((item) => TempoImmutable.parse(item, { timeZone: this.zone }))
			.reduce((best, item) => (millisecondDistance(item.timestampMs, this.timestampMs) > millisecondDistance(best.timestampMs, this.timestampMs) ? item : best));

		return this.make(farthest.toDate(), farthest.timeZone);
	}

	min(other: TempoInput): this {
		return this.isBefore(other) ? this : this.make(asDate(other), zoneFromInput(other, undefined));
	}

	max(other: TempoInput): this {
		return this.isAfter(other) ? this : this.make(asDate(other), zoneFromInput(other, undefined));
	}

	format(pattern: string, options?: FormatOptions): string {
		return new TempoFormatter().format(
			{
				currentLocale: this.currentLocale,
				offsetFor: (timeZone) => this.offsetFor(timeZone),
				timestamp: this.timestamp,
				timestampMs: this.timestampMs,
				value: this.value,
				zone: this.zone,
			},
			pattern,
			options,
		);
	}

	ordinal(unit: 'day' | 'month' | 'quarter' | 'year' = 'day'): string {
		const value = unit === 'year' ? this.year : unit === 'quarter' ? this.quarter : unit === 'month' ? this.month : this.day;

		return ordinal(value);
	}

	meridiem(lowercase = false): string {
		const value = this.hour < 12 ? 'AM' : 'PM';

		return lowercase ? value.toLowerCase() : value;
	}

	week(): number {
		return this.isoWeek;
	}

	weekYear(): number {
		return this.isoWeekYear;
	}

	weeksInYear(): number {
		return this.weeksInISOYear;
	}

	daysFromStartOfWeek(weekStartsOn = 1): number {
		return (this.dayOfWeek - weekStartsOn + 7) % 7;
	}

	setDaysFromStartOfWeek(days: number, weekStartsOn = 1): this {
		return this.startOfWeek({ weekStartsOn }).addDays(days);
	}

	formatIntl(options?: Intl.DateTimeFormatOptions & { readonly locale?: string }): string {
		const { locale, ...dateTimeOptions } = options ?? {};

		return new Intl.DateTimeFormat(locale, {
			timeZone: this.zone,
			...dateTimeOptions,
		}).format(this.value);
	}

	toDate(): Date {
		return new Date(this.value.getTime());
	}

	toDateTime(): Date {
		return this.toDate();
	}

	toDateTimeImmutable(): Date {
		return this.toDate();
	}

	toDateString(): string {
		return this.format('YYYY-MM-DD');
	}

	toTimeString(precision: TimeStringPrecision = 'second'): string {
		const base = this.format('HH:mm:ss');

		return precision === 'millisecond' ? `${base}.${pad(this.millisecond, 3)}` : base;
	}

	toDateTimeString(): string {
		return this.format('YYYY-MM-DD HH:mm:ss');
	}

	toFormattedDateString(): string {
		return this.format('MMM D, YYYY');
	}

	toFormattedDayDateString(): string {
		return this.format('ddd, MMM D, YYYY');
	}

	toDayDateTimeString(): string {
		return this.format('ddd, MMM D, YYYY h:mm A');
	}

	toDateTimeLocalString(precision: TimeStringPrecision = 'second'): string {
		return `${this.toDateString()}T${this.toTimeString(precision)}`;
	}

	toISOString(): string {
		return this.value.toISOString();
	}

	toIso8601String(): string {
		return this.format('YYYY-MM-DDTHH:mm:ssZ');
	}

	toIso8601ZuluString(precision: TimeStringPrecision = 'second'): string {
		return `${this.utc().toDateTimeLocalString(precision)}Z`;
	}

	toRfc3339String(precision: TimeStringPrecision = 'second'): string {
		return `${this.toDateTimeLocalString(precision)}${this.offsetString(':')}`;
	}

	toRfc7231String(): string {
		return this.utc().format('ddd, DD MMM YYYY HH:mm:ss [GMT]');
	}

	toRfc822String(): string {
		return this.format('ddd, DD MMM YY HH:mm:ss ZZ');
	}

	toRfc850String(): string {
		return this.format('dddd, DD-MMM-YY HH:mm:ss ZZ');
	}

	toRfc1036String(): string {
		return this.toRfc822String();
	}

	toRfc1123String(): string {
		return this.toRssString();
	}

	toRfc2822String(): string {
		return this.toRssString();
	}

	toW3cString(): string {
		return this.toRfc3339String();
	}

	toCookieString(): string {
		return this.utc().format('ddd, DD-MMM-YYYY HH:mm:ss [GMT]');
	}

	toAtomString(): string {
		return this.toRfc3339String();
	}

	toRssString(): string {
		return this.format('ddd, DD MMM YYYY HH:mm:ss ZZ');
	}

	toUnixString(): string {
		return String(this.timestamp);
	}

	unix(): number {
		return this.timestamp;
	}

	toJSON(): string {
		return this.policy.serializer?.(this) ?? this.toISOString();
	}

	jsonSerialize(): string {
		return this.toJSON();
	}

	serialize(): string {
		return this.toJSON();
	}

	toObject(): TempoObject {
		const parts = this.parts();

		return {
			...parts,
			offsetMinutes: this.offsetMinutes,
			timeZone: this.zone,
			weekday: parts.weekday,
		};
	}

	toMap(): Map<keyof TempoObject, TempoObject[keyof TempoObject]> {
		return new Map(Object.entries(this.toObject()) as Array<[keyof TempoObject, TempoObject[keyof TempoObject]]>);
	}

	toArray(): [number, number, number, number, number, number, number] {
		const parts = this.parts();

		return [parts.year, parts.month, parts.day, parts.hour, parts.minute, parts.second, parts.millisecond];
	}

	valueOf(): number {
		return this.timestampMs;
	}

	preciseTimestamp(precision = 6): number {
		assertFiniteNumber(precision, 'Precision');

		return Math.round(this.timestampMs * 10 ** (precision - 3));
	}

	toString(): string {
		return this.policy.toStringFormat === null ? this.toISOString() : this.format(this.policy.toStringFormat);
	}

	intervalUntil(end: TempoInput): TempoInterval {
		return new TempoInterval(this, end);
	}

	periodUntil(end: TempoInput, options?: PeriodOptions): TempoPeriod {
		return new TempoPeriod(this, end, options);
	}

	toPeriod(end: TempoInput, options?: PeriodOptions): TempoPeriod {
		return this.periodUntil(end, options);
	}

	until(end: TempoInput, options?: PeriodOptions): TempoPeriod {
		return this.periodUntil(end, options);
	}

	range(end: TempoInput, options?: PeriodOptions): TempoPeriod {
		return this.periodUntil(end, options);
	}

	private parts(): ZonedParts {
		return getZonedParts(this.value, this.zone);
	}

	private offsetFor(timeZone: string): number {
		const parts = getZonedParts(this.value, timeZone);
		const localAsUTC = dateFromPartsAsUTC(parts);

		return Math.trunc((localAsUTC - this.value.getTime()) / millisecondsPerMinute);
	}

	private offsetForDate(date: Date): number {
		const parts = getZonedParts(date, this.zone);
		const localAsUTC = dateFromPartsAsUTC(parts);

		return Math.trunc((localAsUTC - date.getTime()) / millisecondsPerMinute);
	}

	private comparableValue(unit: ComparisonUnit): number {
		return tempoComparison.comparableValue(this, unit);
	}

	private diffFilteredDays(other: TempoInput, predicate: (item: TempoImmutable) => boolean, options?: DiffOptions): number {
		const otherTempo = TempoImmutable.parse(other, { timeZone: this.zone });
		const sign = this.isBefore(otherTempo, 'day') ? -1 : 1;
		const start = sign < 0 ? this.startOf('day') : otherTempo.startOf('day');
		const end = sign < 0 ? otherTempo.startOf('day') : this.startOf('day');

		let current = start;
		let count = 0;

		while (current.isBefore(end, 'day')) {
			current = current.addDays(1);

			if (current.isSameOrBefore(end, 'day') && predicate(current)) {
				count += 1;
			}
		}

		const result = options?.absolute ? count : count * sign;

		return options?.float ? result : Math.trunc(result);
	}
}

export class Tempo extends TempoImmutable {
	static override now(options?: TempoOptions): Tempo {
		const policy = resolveTempoPolicy(options);

		return new Tempo(policy.testNow ?? new Date(), optionsFromPolicy(policy));
	}

	static override today(options?: TempoOptions): Tempo {
		return Tempo.now(options).startOfDay();
	}

	static override tomorrow(options?: TempoOptions): Tempo {
		return Tempo.today(options).addDays(1);
	}

	static override yesterday(options?: TempoOptions): Tempo {
		return Tempo.today(options).subDays(1);
	}

	static override parse(input: TempoInput, options?: TempoOptions): Tempo {
		return new Tempo(input, options);
	}

	static override fromJSON(input: string, options?: TempoOptions): Tempo {
		const value = JSON.parse(input) as unknown;

		if (typeof value !== 'string') {
			throw new RangeError('Tempo JSON must be a string');
		}

		return Tempo.parse(value, options);
	}

	static override fromFormat(input: string, pattern: string, options?: TempoOptions): Tempo {
		return new Tempo(parseFromPattern(input, pattern, options), options);
	}

	static override createNormalized(components: TempoComponents, options?: TempoOptions): Tempo {
		const policy = resolveTempoPolicy(options);
		const timeZone = normalizeTimeZone(components.timeZone ?? policy.timeZone);

		return new Tempo(dateFromZonedComponents(components), {
			...optionsFromPolicy(policy, options),
			strictMode: false,
			timeZone,
		});
	}

	static override create(components: TempoComponents, options?: TempoOptions): Tempo {
		const policy = resolveTempoPolicy(options);
		const timeZone = normalizeTimeZone(components.timeZone ?? policy.timeZone);
		const date = dateFromZonedComponents(components, timeZone);

		assertSafeZonedComponents(components, date, timeZone);

		return new Tempo(date, optionsFromPolicy(policy, { timeZone }));
	}

	static override fromDate(year: number, month = 1, day = 1, options?: TempoOptions): Tempo {
		return Tempo.create({ day, month, timeZone: options?.timeZone, year }, options);
	}

	static override fromTime(hour = 0, minute = 0, second = 0, millisecond = 0, options?: TempoOptions): Tempo {
		return Tempo.today(options).setTime(hour, minute, second, millisecond);
	}

	static override fromTimeString(time: string, options?: TempoOptions): Tempo {
		return Tempo.today(options).setTimeFromTimeString(time);
	}

	static override fromObject(components: TempoComponents): Tempo {
		return Tempo.create(components);
	}

	static override fromTimestamp(timestamp: number, options?: TempoOptions): Tempo {
		return new Tempo(fromNumericTimestamp(timestamp), options);
	}

	static override fromTimestampMs(timestamp: number, options?: TempoOptions): Tempo {
		assertFiniteNumber(timestamp, 'Timestamp');

		return new Tempo(new Date(timestamp), options);
	}

	static override fromTimestampUTC(timestamp: number): Tempo {
		return Tempo.fromTimestamp(timestamp, { timeZone: defaultTimeZone });
	}

	static override fromTimestampMsUTC(timestamp: number): Tempo {
		return Tempo.fromTimestampMs(timestamp, { timeZone: defaultTimeZone });
	}
}
