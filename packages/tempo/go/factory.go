package tempo

import (
	"time"

	factorypkg "github.com/oullin/alloy/tempo/factory"
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

func NewFactoryWithTestNow(input Tempo, options ...Option) (Factory, error) {
	cfg := config{
		location:       input.location,
		runtime:        input.Runtime(),
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

func (factory Factory) Now() Tempo {
	return factory.newTempo(factory.clock.Now())
}

func (factory Factory) Runtime() Runtime {
	if factory.runtime.Locale() == "" {
		settings := factory.settingsSnapshot()

		return NewRuntime(
			RuntimeLocale(settings.Locale),
			RuntimeFallbackLocale(settings.FallbackLocale),
		)
	}

	return factory.runtime
}

func (factory Factory) WithRuntime(runtime Runtime) Factory {
	factory.runtime = runtime

	return factory
}

func (factory Factory) WithTranslator(translator Translator) Factory {
	factory.runtime = factory.Runtime().With(RuntimeTranslator(translator))

	return factory
}

func (factory Factory) Today() Tempo {
	return factory.Now().StartOfDay()
}

func (factory Factory) Tomorrow() Tempo {
	return factory.Today().AddDays(1)
}

func (factory Factory) Yesterday() Tempo {
	return factory.Today().SubDays(1)
}

func (factory Factory) ImmutableNow() Tempo {
	return factory.Now()
}

func (factory Factory) MutableNow() *MutableTempo {
	return NewMutable(factory.Now())
}

func (factory Factory) FromTime(value time.Time) Tempo {
	return factory.newTempo(value)
}

func (factory Factory) Parse(input string) (Tempo, error) {
	return factory.parser().Parse(input)
}

func (factory Factory) FromFormat(input string, pattern string) (Tempo, error) {
	return factory.parser().FromFormat(input, pattern)
}

func (factory Factory) Create(components Components) (Tempo, error) {
	location := factory.location

	if components.Timezone != "" {
		nextLocation, err := loadLocation(components.Timezone)

		if err != nil {
			return Tempo{}, err
		}

		location = nextLocation
	}

	value := timeFromComponents(components, location)

	if !componentsMatchTime(components, value, location) {
		return Tempo{}, errInvalidComponents
	}

	return newTempoWithPolicy(value, location, factory.runtime, factory.settingsSnapshot(), factory.serializer, factory.toStringFormat), nil
}

func (factory Factory) CreateNormalized(components Components) (Tempo, error) {
	location := factory.location

	if components.Timezone != "" {
		nextLocation, err := loadLocation(components.Timezone)

		if err != nil {
			return Tempo{}, err
		}

		location = nextLocation
	}

	return newTempoWithPolicy(timeFromComponents(components, location), location, factory.runtime, factory.settingsSnapshot(), factory.serializer, factory.toStringFormat), nil
}

func (factory Factory) CreateFromDate(year int, month int, day int) (Tempo, error) {
	return factory.Create(Components{Year: year, Month: month, Day: day})
}

func (factory Factory) CreateFromTime(hour int, minute int, second int, millisecond int) (Tempo, error) {
	return factory.Today().SetTime(hour, minute, second, millisecond)
}

func (factory Factory) CreateFromTimeString(input string) (Tempo, error) {
	return factory.Today().SetTimeFromTimeString(input)
}

func (factory Factory) FromObject(components Components) (Tempo, error) {
	return factory.Create(components)
}

func (factory Factory) FromTimestamp(timestamp int64) Tempo {
	return factory.newTempo(time.Unix(timestamp, 0))
}

func (factory Factory) FromTimestampMs(timestamp int64) Tempo {
	return factory.newTempo(time.UnixMilli(timestamp))
}

func (factory Factory) settingsSnapshot() Settings {
	return cloneSettings(normalizeSettings(factory.settings))
}

func (factory Factory) newTempo(value time.Time) Tempo {
	return newTempoWithPolicy(value, factory.location, factory.runtime, factory.settingsSnapshot(), factory.serializer, factory.toStringFormat)
}

func (factory Factory) parser() Parser {
	return newParserWithPolicy(factory.location, factory.runtime, factory.settingsSnapshot(), factory.serializer, factory.toStringFormat)
}
