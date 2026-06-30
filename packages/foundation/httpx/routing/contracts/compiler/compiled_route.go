package compiler

import "regexp"

// Token is one element of a compiled route pattern.
type Token struct {
	Kind      string
	Prefix    string
	Regexp    string
	Name      string
	Utf8      bool
	Important bool
}

// CompiledRoute is the result of compiling a route URI and optional host.
type CompiledRoute struct {
	staticPrefix  string
	regex         string
	tokens        []Token
	pathVariables []string
	hostRegex     string
	hostTokens    []Token
	hostVariables []string
	variables     []string

	compiledRegex     *regexp.Regexp
	compiledHostRegex *regexp.Regexp
}

// NewCompiledRoute constructs a CompiledRoute.
func NewCompiledRoute(
	staticPrefix, regex string,
	tokens []Token,
	pathVariables []string,
	hostRegex string,
	hostTokens []Token,
	hostVariables []string,
	variables []string,
) *CompiledRoute {
	cr := &CompiledRoute{
		staticPrefix:  staticPrefix,
		regex:         regex,
		tokens:        tokens,
		pathVariables: pathVariables,
		hostRegex:     hostRegex,
		hostTokens:    hostTokens,
		hostVariables: hostVariables,
		variables:     variables,
	}

	if regex != "" {
		cr.compiledRegex = regexp.MustCompile(regex)
	}

	if hostRegex != "" {
		cr.compiledHostRegex = regexp.MustCompile(hostRegex)
	}

	return cr
}

func (c *CompiledRoute) StaticPrefix() string              { return c.staticPrefix }
func (c *CompiledRoute) Regex() string                     { return c.regex }
func (c *CompiledRoute) Tokens() []Token                   { return c.tokens }
func (c *CompiledRoute) PathVariables() []string           { return c.pathVariables }
func (c *CompiledRoute) HostRegex() string                 { return c.hostRegex }
func (c *CompiledRoute) HostTokens() []Token               { return c.hostTokens }
func (c *CompiledRoute) HostVariables() []string           { return c.hostVariables }
func (c *CompiledRoute) Variables() []string               { return c.variables }
func (c *CompiledRoute) CompiledRegex() *regexp.Regexp     { return c.compiledRegex }
func (c *CompiledRoute) CompiledHostRegex() *regexp.Regexp { return c.compiledHostRegex }
