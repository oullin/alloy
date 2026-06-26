package parser

import (
	"regexp"

	"github.com/oullin/alloy/api/tempo/config"
)

var parserSettings = config.DefaultParserSettings()

var (
	dateOnlyPattern = regexp.MustCompile(parserSettings.DateOnlyPattern)
	localPattern    = regexp.MustCompile(parserSettings.LocalPattern)
	zonePattern     = regexp.MustCompile(parserSettings.ZonePattern)
)
