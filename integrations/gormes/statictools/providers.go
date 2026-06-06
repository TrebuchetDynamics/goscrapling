package statictools

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

type StaticProvider struct {
	Name       string
	Shortcut   string
	Categories []string
	Enabled    bool
	Timeout    time.Duration
	Weight     float64
	Recipe     string
}

func (a StaticExtractionAdapter) ProvidersByCategory(category string) []StaticProvider {
	category = strings.TrimSpace(strings.ToLower(category))
	providers := make([]StaticProvider, 0, len(a.Providers))
	for _, provider := range a.Providers {
		if !provider.Enabled {
			continue
		}
		if category != "" && !providerHasCategory(provider, category) {
			continue
		}
		providers = append(providers, cloneProvider(provider))
	}
	sort.Slice(providers, func(left, right int) bool {
		return providers[left].Name < providers[right].Name
	})
	return providers
}

func (a StaticExtractionAdapter) ResolveProvider(key string) (StaticProvider, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return StaticProvider{}, fmt.Errorf("%s: provider is required", ToolWebExtract)
	}

	for _, provider := range a.Providers {
		if strings.EqualFold(provider.Name, key) {
			if !provider.Enabled {
				return StaticProvider{}, fmt.Errorf("%s: provider %q is disabled", ToolWebExtract, provider.Name)
			}
			return cloneProvider(provider), nil
		}
	}

	var matches []StaticProvider
	for _, provider := range a.Providers {
		if strings.EqualFold(provider.Shortcut, key) {
			matches = append(matches, provider)
		}
	}
	if len(matches) == 0 {
		return StaticProvider{}, fmt.Errorf("%s: unknown provider %q", ToolWebExtract, key)
	}
	if len(matches) > 1 {
		names := make([]string, 0, len(matches))
		for _, provider := range matches {
			names = append(names, provider.Name)
		}
		sort.Strings(names)
		return StaticProvider{}, fmt.Errorf("%s: ambiguous provider shortcut %q: %s", ToolWebExtract, key, strings.Join(names, ", "))
	}
	provider := matches[0]
	if !provider.Enabled {
		return StaticProvider{}, fmt.Errorf("%s: provider %q is disabled", ToolWebExtract, provider.Name)
	}
	return cloneProvider(provider), nil
}

func (a StaticExtractionAdapter) extractProvider(ctx context.Context, call StaticToolCall) (WebExtractResult, error) {
	provider, err := a.ResolveProvider(call.Provider)
	if err != nil {
		return WebExtractResult{}, err
	}
	if strings.TrimSpace(provider.Recipe) == "" {
		return WebExtractResult{}, fmt.Errorf("%s: provider %q recipe is required", ToolWebExtract, provider.Name)
	}
	call.Recipe = provider.Recipe
	if call.Opts.Timeout == 0 && provider.Timeout > 0 {
		call.Opts.Timeout = provider.Timeout
	}
	result, err := a.extractRecipe(ctx, call)
	if err != nil {
		return WebExtractResult{}, err
	}
	if result.Extraction != nil {
		result.Extraction.Provider = provider.Name
		result.Extraction.ProviderShortcut = provider.Shortcut
		result.Extraction.ProviderWeight = provider.Weight
		result.Extraction.ProviderTimeout = provider.Timeout
	}
	return result, nil
}

func providerHasCategory(provider StaticProvider, category string) bool {
	for _, candidate := range provider.Categories {
		if strings.EqualFold(strings.TrimSpace(candidate), category) {
			return true
		}
	}
	return false
}

func cloneProvider(provider StaticProvider) StaticProvider {
	provider.Categories = append([]string(nil), provider.Categories...)
	return provider
}
