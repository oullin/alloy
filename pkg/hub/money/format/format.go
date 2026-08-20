package format

import (
	"math"
	"strconv"
	"strings"
)

// Renderer handles the formatting of monetary amounts.
type Renderer struct {
	Fraction int    // Number of decimal places (e.g., 2 for SGD, 0 for JPY)
	Decimal  string // The character separating whole numbers from decimals (e.g., ".")
	Thousand string // The character separating thousands (e.g., ",")
	Grapheme string // The currency symbol (e.g. "$", "£", "€")
	Template string // The pattern for the output (e.g. "$1")
}

// NewFormatter creates a new Renderer instance with the specified configuration.
func NewFormatter(fraction int, decimal, thousand, grapheme, template string) *Renderer {
	return &Renderer{
		Fraction: fraction,
		Decimal:  decimal,
		Thousand: thousand,
		Grapheme: grapheme,
		Template: template,
	}
}

// Format returns string of formatted integer using given currency template.
func (f *Renderer) Format(amount int64) string {
	// Work with absolute amount value
	sa := strconv.FormatInt(f.abs(amount), 10)

	if len(sa) <= f.Fraction {
		sa = strings.Repeat("0", f.Fraction-len(sa)+1) + sa
	}

	if f.Thousand != "" {
		for i := len(sa) - f.Fraction - 3; i > 0; i -= 3 {
			sa = sa[:i] + f.Thousand + sa[i:]
		}
	}

	if f.Fraction > 0 {
		sa = sa[:len(sa)-f.Fraction] + f.Decimal + sa[len(sa)-f.Fraction:]
	}

	sa = strings.Replace(f.Template, "1", sa, 1)
	sa = strings.Replace(sa, "$", f.Grapheme, 1)

	// Add minus sign for negative amount.
	if amount < 0 {
		sa = "-" + sa
	}

	return sa
}

// FormatWhole returns the amount as whole major units, with no decimal part.
//
// A per-instalment line needs its cents; a headline figure -- a reference
// price, a due amount stated once on its own line -- reads them as noise.
// Rounding is half away from zero, so this states a figure rather than quietly
// truncating it.
//
// The TypeScript twin's MoneyFormatter.formatWhole must agree, and
// conformance/money.json pins that.
func (f *Renderer) FormatWhole(amount int64) string {
	minorPerMajor := int64(1)

	for range f.Fraction {
		minorPerMajor *= 10
	}

	absolute := f.abs(amount)
	major := divideRounding(absolute, minorPerMajor)

	if amount < 0 {
		major = -major
	}

	return NewFormatter(0, f.Decimal, f.Thousand, f.Grapheme, f.Template).Format(major)
}

// FormatCompactSignificant returns the amount abbreviated to a fixed number of
// significant digits, such as "$47.2M" or "$4.26M" at three.
//
// FormatCompact keeps one decimal, which reads well for a single figure but not
// for a column: "47.2M" beside "4.2M" gives the second figure a digit less of
// information than the first. Fixing the significant digits instead keeps every
// row equally precise, which is what a table of totals wants.
//
// Trailing zeros are dropped -- "4B", never "4.00B" -- so the digit count is a
// ceiling rather than padding.
//
// The TypeScript twin's MoneyFormatter.formatCompactSignificant must agree, and
// conformance/money.json pins that.
func (f *Renderer) FormatCompactSignificant(amount int64, significantDigits int) string {
	absolute := f.abs(amount)
	minorPerMajor := int64(1)

	for range f.Fraction {
		minorPerMajor *= 10
	}

	if absolute < compactFloorMajorUnits*minorPerMajor {
		return f.Format(amount)
	}

	for _, scale := range compactScales {
		unit := scale.divisor * minorPerMajor

		if absolute < unit {
			continue
		}

		return f.renderSignificant(absolute, unit, significantDigits, amount < 0, scale.suffix)
	}

	return f.Format(amount)
}

// renderSignificant renders absolute scaled to unit at a fixed significant-digit
// count.
//
// The decimal count falls out of how many digits the integer part already
// spends: "47" spends two of three, leaving one decimal; "4" spends one, leaving
// two. Trailing zeros are then dropped so the count is a ceiling.
func (f *Renderer) renderSignificant(absolute, unit int64, significantDigits int, negative bool, suffix string) string {
	integerDigits := len(strconv.FormatInt(absolute/unit, 10))

	decimals := significantDigits - integerDigits

	if decimals < 0 {
		decimals = 0
	}

	factor := int64(1)

	for range decimals {
		factor *= 10
	}

	scaled := divideRounding(absolute*factor, unit)

	for decimals > 0 && scaled%10 == 0 {
		scaled /= 10
		decimals--
	}

	if negative {
		scaled = -scaled
	}

	rendered := NewFormatter(decimals, f.Decimal, f.Thousand, f.Grapheme, f.Template).Format(scaled)

	return appendSuffix(rendered, suffix)
}

// divideRounding divides half away from zero, on non-negative operands.
func divideRounding(value, unit int64) int64 {
	quotient := value / unit

	if (value%unit)*2 < unit {
		return quotient
	}

	return quotient + 1
}

// ToMajorUnits converts the integer amount to a float representing the major currency units.
// e.g., 100 cents becomes 1.00 dollars if fraction is 2.
func (f *Renderer) ToMajorUnits(amount int64) float64 {
	if f.Fraction == 0 {
		return float64(amount)
	}

	return float64(amount) / math.Pow10(f.Fraction)
}

func (f *Renderer) abs(amount int64) int64 {
	if amount < 0 {
		return -amount
	}

	return amount
}

// compactScales are the abbreviation steps in major units, ordered largest first.
var compactScales = []struct {
	divisor int64
	suffix  string
}{
	{1_000_000_000, "B"},
	{1_000_000, "M"},
	{1_000, "K"},
}

// compactFloorMajorUnits is the threshold below which FormatCompact defers to Format.
const compactFloorMajorUnits = 1_000

// FormatCompact returns the amount abbreviated to a scale suffix, such as
// "$1.3M" or "750K".
//
// Amounts below one thousand major units are returned at full precision by
// Format: rounding 950.00 to 1K overstates it, and a figure that short does not
// need shortening.
//
// A scale is chosen only when the amount fills at least one of it, largest
// first, so a value never renders in a smaller scale than it belongs to. One
// decimal is kept only when it carries information -- "4B", not "4.0B" -- and
// the suffix lands directly after the last digit so it sits inside the number
// rather than after a trailing grapheme.
//
// Everything is computed in int64, so no precision is lost on the way to the
// rounded display string. The TypeScript twin's MoneyFormatter.formatCompact
// must agree, and conformance/money.json pins that.
func (f *Renderer) FormatCompact(amount int64) string {
	absolute := f.abs(amount)
	minorPerMajor := int64(1)

	for range f.Fraction {
		minorPerMajor *= 10
	}

	if absolute < compactFloorMajorUnits*minorPerMajor {
		return f.Format(amount)
	}

	for _, scale := range compactScales {
		tenths := tenthsOf(absolute, scale.divisor*minorPerMajor)

		if tenths < 10 {
			continue
		}

		if amount < 0 {
			tenths = -tenths
		}

		return f.renderCompact(tenths, scale.suffix)
	}

	return f.Format(amount)
}

// renderCompact renders a tenths count at this currency's separators and
// template. The fraction is chosen per value rather than fixed, so a whole
// number of tenths asks for no decimals at all -- which is what keeps "4B" from
// becoming "4.0B" without a second rounding pass.
func (f *Renderer) renderCompact(tenths int64, suffix string) string {
	fraction := 1

	if tenths%10 == 0 {
		fraction = 0
		tenths /= 10
	}

	scaled := NewFormatter(fraction, f.Decimal, f.Thousand, f.Grapheme, f.Template)

	return appendSuffix(scaled.Format(tenths), suffix)
}

// tenthsOf returns how many tenths of unit the amount is, rounded half away from zero.
func tenthsOf(absolute, unit int64) int64 {
	scaled := absolute * 10
	quotient := scaled / unit

	if scaled%unit*2 < unit {
		return quotient
	}

	return quotient + 1
}

// appendSuffix places the suffix after the last digit.
//
// Currencies differ on whether the grapheme leads or trails, so appending to the
// end would land the suffix after the trailing grapheme for AED while working
// for USD. Scanning for an ASCII digit is sufficient: the separators are ASCII
// and no grapheme in the dataset contains an ASCII digit.
func appendSuffix(formatted, suffix string) string {
	for index := len(formatted) - 1; index >= 0; index-- {
		if formatted[index] >= '0' && formatted[index] <= '9' {
			return formatted[:index+1] + suffix + formatted[index+1:]
		}
	}

	return formatted + suffix
}
