package config

import (
	"fmt"
	"regexp"
	"strings"
)

// Validate checks a merged configuration and compiles every orphan pattern in
// place, so a rule that reaches the scanner is always ready to use and no hot
// path ever compiles a regular expression.
//
// It takes a pointer because compiling the patterns is the point: returning a
// copy would leave the caller holding uncompiled rules.
func Validate(c *Config) error {
	if c.Version == 0 {
		c.Version = Version
	}

	if c.Version != Version {
		return fmt.Errorf("%w: %d (treex understands %d)", ErrUnsupportedVersion, c.Version, Version)
	}

	if c.Defaults.MaxDepth <= 0 {
		c.Defaults.MaxDepth = 4
	}

	if c.Artifacts.MaxDepth <= 0 {
		c.Artifacts.MaxDepth = 6
	}

	seen := make(map[string]struct{}, len(c.Providers))

	for index := range c.Providers {
		provider := &c.Providers[index]

		if err := validateProvider(provider, seen); err != nil {
			return err
		}
	}

	return nil
}

func validateProvider(provider *Provider, seen map[string]struct{}) error {
	provider.Name = strings.TrimSpace(provider.Name)

	if provider.Name == "" {
		return fmt.Errorf("%w: a provider has no name", ErrInvalidProvider)
	}

	if _, duplicate := seen[provider.Name]; duplicate {
		return fmt.Errorf("%w: %q is defined twice", ErrInvalidProvider, provider.Name)
	}

	seen[provider.Name] = struct{}{}

	if strings.TrimSpace(provider.Root) == "" {
		return fmt.Errorf("%w: %q has no root", ErrInvalidProvider, provider.Name)
	}

	if provider.Kind == "" {
		provider.Kind = KindAgent
	}

	if provider.Kind != KindAgent && provider.Kind != KindToolchain {
		return fmt.Errorf("%w: %q has unknown kind %q", ErrInvalidProvider, provider.Name, provider.Kind)
	}

	for source := range provider.Worktrees {
		if provider.Worktrees[source].Depth <= 0 {
			provider.Worktrees[source].Depth = 2
		}
	}

	for rule := range provider.Orphans {
		if err := compileOrphan(provider.Name, &provider.Orphans[rule]); err != nil {
			return err
		}
	}

	return nil
}

func compileOrphan(provider string, rule *OrphanRule) error {
	if strings.TrimSpace(rule.Name) == "" {
		return fmt.Errorf("%w: %s has an unnamed rule", ErrInvalidOrphanRule, provider)
	}

	if rule.Liveness == "" {
		rule.Liveness = LivenessNone
	}

	if rule.Liveness != LivenessPID && rule.Liveness != LivenessNone {
		return fmt.Errorf("%w: %s/%s has unknown liveness %q", ErrInvalidOrphanRule, provider, rule.Name, rule.Liveness)
	}

	pattern, err := regexp.Compile(rule.Match)

	if err != nil {
		return fmt.Errorf("%w: compile %s/%s: %w", ErrInvalidOrphanRule, provider, rule.Name, err)
	}

	// A pid rule that cannot extract a pid would treat every match as dead,
	// which for ~/.agent-browser would delete the sockets of running browsers.
	if rule.Liveness == LivenessPID && pattern.SubexpIndex(rule.Group) < 0 {
		return fmt.Errorf(
			"%w: %s/%s asks for pid liveness but captures no group %q",
			ErrInvalidOrphanRule, provider, rule.Name, rule.Group,
		)
	}

	rule.pattern = pattern

	return nil
}
