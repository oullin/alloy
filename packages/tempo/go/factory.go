package tempo

import (
	"errors"
	"time"

	factorypkg "github.com/oullin/alloy/tempo/factory"
)

func NewFactory(options ...Option) (Factory, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Factory{}, err
	}

	var now *time.Time
	if cfg.app != nil && cfg.app.Settings.TestNow != nil {
		value := cfg.app.Settings.TestNow.value
		now = &value
	}

	return Factory{clock: factorypkg.NewClock(now), location: cfg.location, runtime: cfg.runtime}, nil
}

func NewFactoryWithTestNow(input Tempo, options ...Option) (Factory, error) {
	cfg := config{location: input.location, runtime: input.Runtime()}
	for _, option := range options {
		if err := option(&cfg); err != nil {
			return Factory{}, err
		}
	}

	now := input.value
	return Factory{clock: factorypkg.NewClock(&now), location: cfg.location, runtime: cfg.runtime}, nil
}

func (factory Factory) Now() Tempo {
	return newTempo(factory.clock.Now(), factory.location, factory.runtime)
}

func (factory Factory) Runtime() Runtime {
	if factory.runtime.Locale() == "" {
		return NewRuntime(
			RuntimeLocale(defaultConfig.Settings.Locale),
			RuntimeFallbackLocale(defaultConfig.Settings.FallbackLocale),
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
	return newTempo(value, factory.location, factory.runtime)
}

func (factory Factory) Parse(input string) (Tempo, error) {
	return newParser(factory.location, factory.runtime).Parse(input)
}

func (factory Factory) TryParse(input string) (Tempo, bool) {
	tempo, err := factory.Parse(input)
	return tempo, err == nil
}

func (factory Factory) CanParse(input string) bool {
	_, ok := factory.TryParse(input)
	return ok
}

func (factory Factory) FromFormat(input string, pattern string) (Tempo, error) {
	return newParser(factory.location, factory.runtime).FromFormat(input, pattern)
}

func (factory Factory) CreateFromFormat(input string, pattern string) (Tempo, error) {
	return factory.FromFormat(input, pattern)
}

func (factory Factory) TryFromFormat(input string, pattern string) (Tempo, bool) {
	tempo, err := factory.FromFormat(input, pattern)
	return tempo, err == nil
}

func (factory Factory) HasFormat(input string, pattern string) bool {
	_, ok := factory.TryFromFormat(input, pattern)
	return ok
}

func (factory Factory) CanBeCreatedFromFormat(input string, pattern string) bool {
	return factory.HasFormat(input, pattern)
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

	return newTempo(timeFromComponents(components, location), location, factory.runtime), nil
}

func (factory Factory) CreateSafe(components Components) (Tempo, error) {
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
		return Tempo{}, errors.New("invalid Tempo local date/time components")
	}

	return newTempo(value, location, factory.runtime), nil
}

func (factory Factory) CreateFromDate(year int, month int, day int) Tempo {
	return Tempo{
		value:    timeFromComponents(Components{Year: year, Month: month, Day: day}, factory.location).UTC(),
		location: factory.location,
		runtime:  factory.runtime,
	}
}

func (factory Factory) CreateMidnightDate(year int, month int, day int) Tempo {
	return factory.CreateFromDate(year, month, day)
}

func (factory Factory) CreateFromTime(hour int, minute int, second int, millisecond int) Tempo {
	return factory.Today().SetTime(hour, minute, second, millisecond)
}

func (factory Factory) CreateFromTimeString(input string) (Tempo, error) {
	return factory.Today().SetTimeFromTimeString(input)
}

func (factory Factory) FromObject(components Components) (Tempo, error) {
	return factory.Create(components)
}

func (factory Factory) FromTimestamp(timestamp int64) Tempo {
	return newTempo(time.Unix(timestamp, 0), factory.location, factory.runtime)
}

func (factory Factory) CreateFromTimestamp(timestamp int64) Tempo {
	return factory.FromTimestamp(timestamp)
}

func (factory Factory) FromTimestampMs(timestamp int64) Tempo {
	return newTempo(time.UnixMilli(timestamp), factory.location, factory.runtime)
}

func (factory Factory) CreateFromTimestampMs(timestamp int64) Tempo {
	return factory.FromTimestampMs(timestamp)
}
