package goscrapling

import (
	"strings"
	"unicode"
)

const defaultMinScore = 0.65

func scoreFingerprint(candidate Fingerprint, target Fingerprint) float64 {
	const (
		tagWeight       = 0.24
		textWeight      = 0.24
		attrNameWeight  = 0.12
		attrValueWeight = 0.16
		parentWeight    = 0.12
		siblingWeight   = 0.05
		pathWeight      = 0.07
	)

	return exactScore(candidate.Tag, target.Tag)*tagWeight +
		stringSimilarity(candidate.Text, target.Text)*textWeight +
		mapKeySimilarity(candidate.Attributes, target.Attributes)*attrNameWeight +
		attrValueSimilarity(candidate.Attributes, target.Attributes)*attrValueWeight +
		parentSimilarity(candidate, target)*parentWeight +
		sequenceSimilarity(candidate.SiblingTags, target.SiblingTags)*siblingWeight +
		sequenceSimilarity(candidate.PathTags, target.PathTags)*pathWeight
}

func parentSimilarity(candidate Fingerprint, target Fingerprint) float64 {
	const (
		tagWeight       = 0.35
		textWeight      = 0.25
		attrNameWeight  = 0.20
		attrValueWeight = 0.20
	)

	return exactScore(candidate.ParentTag, target.ParentTag)*tagWeight +
		stringSimilarity(candidate.ParentText, target.ParentText)*textWeight +
		mapKeySimilarity(candidate.ParentAttributes, target.ParentAttributes)*attrNameWeight +
		attrValueSimilarity(candidate.ParentAttributes, target.ParentAttributes)*attrValueWeight
}

func exactScore(candidate string, target string) float64 {
	candidate = normalizeSpace(strings.ToLower(candidate))
	target = normalizeSpace(strings.ToLower(target))
	if candidate == "" && target == "" {
		return 1
	}
	if candidate != "" && candidate == target {
		return 1
	}
	return 0
}

func stringSimilarity(candidate string, target string) float64 {
	candidate = normalizeSpace(strings.ToLower(candidate))
	target = normalizeSpace(strings.ToLower(target))
	if candidate == "" && target == "" {
		return 1
	}
	if candidate == "" || target == "" {
		return 0
	}
	if candidate == target {
		return 1
	}

	return tokenSetSimilarity(tokenSet(candidate), tokenSet(target))
}

func mapKeySimilarity(candidate map[string]string, target map[string]string) float64 {
	if len(candidate) == 0 && len(target) == 0 {
		return 1
	}
	if len(candidate) == 0 || len(target) == 0 {
		return 0
	}

	candidateKeys := make(map[string]struct{}, len(candidate))
	for key := range candidate {
		candidateKeys[strings.ToLower(key)] = struct{}{}
	}
	targetKeys := make(map[string]struct{}, len(target))
	for key := range target {
		targetKeys[strings.ToLower(key)] = struct{}{}
	}

	return tokenSetSimilarity(candidateKeys, targetKeys)
}

func attrValueSimilarity(candidate map[string]string, target map[string]string) float64 {
	if len(candidate) == 0 && len(target) == 0 {
		return 1
	}
	if len(candidate) == 0 || len(target) == 0 {
		return 0
	}

	return tokenSetSimilarity(attrValueTokens(candidate), attrValueTokens(target))
}

func attrValueTokens(attrs map[string]string) map[string]struct{} {
	tokens := make(map[string]struct{})
	for _, value := range attrs {
		for token := range tokenSet(value) {
			tokens[token] = struct{}{}
		}
	}
	return tokens
}

func tokenSet(value string) map[string]struct{} {
	tokens := make(map[string]struct{})
	var builder strings.Builder

	flush := func() {
		if builder.Len() == 0 {
			return
		}
		tokens[builder.String()] = struct{}{}
		builder.Reset()
	}

	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			continue
		}
		flush()
	}
	flush()

	return tokens
}

func tokenSetSimilarity(candidate map[string]struct{}, target map[string]struct{}) float64 {
	if len(candidate) == 0 && len(target) == 0 {
		return 1
	}
	if len(candidate) == 0 || len(target) == 0 {
		return 0
	}

	intersection := 0
	union := make(map[string]struct{}, len(candidate)+len(target))
	for token := range candidate {
		union[token] = struct{}{}
		if _, ok := target[token]; ok {
			intersection++
		}
	}
	for token := range target {
		union[token] = struct{}{}
	}

	return float64(intersection) / float64(len(union))
}

func sequenceSimilarity(candidate []string, target []string) float64 {
	if len(candidate) == 0 && len(target) == 0 {
		return 1
	}
	if len(candidate) == 0 || len(target) == 0 {
		return 0
	}

	lcs := longestCommonSubsequence(candidate, target)
	maxLength := len(candidate)
	if len(target) > maxLength {
		maxLength = len(target)
	}

	return float64(lcs) / float64(maxLength)
}

func longestCommonSubsequence(a []string, b []string) int {
	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				current[j] = previous[j-1] + 1
			} else if previous[j] > current[j-1] {
				current[j] = previous[j]
			} else {
				current[j] = current[j-1]
			}
		}
		previous, current = current, previous
		for j := range current {
			current[j] = 0
		}
	}

	return previous[len(b)]
}
