package config

type ParserSettings struct {
	DateOnlyPattern string
	LocalPattern    string
	ZonePattern     string
	FormatTokens    []string
}

func DefaultParserSettings() ParserSettings {
	return ParserSettings{
		DateOnlyPattern: `^(\d{4})-(\d{2})-(\d{2})$`,
		LocalPattern:    `^(\d{4})-(\d{2})-(\d{2})(?:[T\s](\d{2})(?::?(\d{2}))?(?::?(\d{2})(?:\.(\d{1,9}))?)?)?$`,
		ZonePattern:     `(?:Z|[+-]\d{2}:?\d{2})$`,
		FormatTokens:    []string{"YYYY", "MMMM", "dddd", "MMM", "ddd", "SSS", "Do", "YY", "ZZ", "MM", "DD", "HH", "hh", "mm", "ss", "Z", "M", "D", "H", "h", "m", "s", "A", "a"},
	}
}
