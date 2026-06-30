package tempo

import (
	"time"

	factorypkg "alloy.dev/foundation/tempo/factory"
)

func NewFactory(options ...Option) (Factory, error) {
	cfg, err := applyOptions(options...)

	if err != nil {
		return Factory{}, err
	}

	var now *time.Time

	if cfg.settings.TestNow != nil {
		value := cfg.settings.TestNow.value
		now = &value
	}

	return Factory{
		clock:          factorypkg.NewClock(now),
		location:       cfg.location,
		runtime:        cfg.runtime,
		settings:       cloneSettings(cfg.settings),
		serializer:     cfg.serializer,
		toStringFormat: cfg.toStringFormat,
	}, nil
}

func NewFactoryWithTestNow(input Time, options ...Option) (Factory, error) {
	cfg := config{
		location:       input.location,
		runtime:        input.Context(),
		settings:       input.settingsSnapshot(),
		serializer:     input.serializer,
		toStringFormat: input.toStringFormat,
	}

	for _, option := range options {
		if err := option(&cfg); err != nil {
			return Factory{}, err
		}
	}

	now := input.value

	return Factory{
		clock:          factorypkg.NewClock(&now),
		location:       cfg.location,
		runtime:        cfg.runtime,
		settings:       cloneSettings(cfg.settings),
		serializer:     cfg.serializer,
		toStringFormat: cfg.toStringFormat,
	}, nil
}

func (factory Factory) Now() Time {
	return factory.newTempo(factory.clock.Now())
}

func (factory Factory) Context() Context {
	if factory.runtime.Locale() == "" {
		settings := factory.settingsSnapshot()

		return NewRuntime(
			RuntimeLocale(settings.Locale),
			RuntimeFallbackLocale(settings.FallbackLocale),
		)
	}

	return factory.runtime
}

func (factory Factory) WithRuntime(runtime Context) Factory {
	factory.runtime = runtime

	return factory
}

func (factory Factory) WithTranslator(translator Translator) Factory {
	factory.runtime = factory.Context().With(RuntimeTranslator(translator))

	return factory
}

func (factory Factory) Today() Time {
	return factory.Now().StartOfDay()
}

func (factory Factory) Tomorrow() Time {
	return factory.Today().AddDays(1)
}

func (factory Factory) Yesterday() Time {
	return factory.Today().SubDays(1)
}

func (factory Factory) ImmutableNow() Time {
	return factory.Now()
}

func (factory Factory) MutableNow() *MutableTime {
	return NewMutable(factory.Now())
}

func (factory Factory) FromTime(value time.Time) Time {
	return factory.newTempo(value)
}

func (factory Factory) Parse(input string) (Time, error) {
	return factory.parser().Parse(input)
}

func (factory Factory) FromFormat(input string, pattern string) (Time, error) {
	return factory.parser().FromFormat(input, pattern)
}

func (factory Factory) Create(components Components) (Time, error) {
	location := factory.location

	if components.Timezone != "" {
		nextLocation, err := loadLocation(components.Timezone)

		if err != nil {
			return Time{}, err
		}

		location = nextLocation
	}

	value := timeFromComponents(components, location)

	if !componentsMatchTime(components, value, location) {
		return Time{}, errInvalidComponents
	}

	return newTempoWithPolicy(value, location, factory.runtime, factory.settingsSnapshot(), factory.serializer, factory.toStringFormat), nil
}

func (factory Factory) CreateNormalized(components Components) (Time, error) {
	location := factory.location

	if components.Timezone != "" {
		nextLocation, err := loadLocation(components.Timezone)

		if err != nil {
			return Time{}, err
		}

		location = nextLocation
	}

	return newTempoWithPolicy(timeFromComponents(components, location), location, factory.runtime, factory.settingsSnapshot(), factory.serializer, factory.toStringFormat), nil
}

func (factory Factory) CreateFromDate(year int, month int, day int) (Time, error) {
	return factory.Create(Components{Year: year, Month: month, Day: day})
}

func (factory Factory) CreateFromTime(hour int, minute int, second int, millisecond int) (Time, error) {
	return factory.Today().SetTime(hour, minute, second, millisecond)
}

func (factory Factory) CreateFromTimeString(input string) (Time, error) {
	return factory.Today().SetTimeFromTimeString(input)
}

func (factory Factory) FromObject(components Components) (Time, error) {
	return factory.Create(components)
}

func (factory Factory) FromTimestamp(timestamp int64) Time {
	return factory.newTempo(time.Unix(timestamp, 0))
}

func (factory Factory) FromTimestampMs(timestamp int64) Time {
	return factory.newTempo(time.UnixMilli(timestamp))
}

func (factory Factory) settingsSnapshot() Settings {
	return cloneSettings(normalizeSettings(factory.settings))
}

func (factory Factory) newTempo(value time.Time) Time {
	return newTempoWithPolicy(value, factory.location, factory.runtime, factory.settingsSnapshot(), factory.serializer, factory.toStringFormat)
}

func (factory Factory) parser() Parser {
	return newParserWithPolicy(factory.location, factory.runtime, factory.settingsSnapshot(), factory.serializer, factory.toStringFormat)
}
