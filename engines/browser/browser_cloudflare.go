package browser

// CloudflareChallengeStatus describes goscrapling's owned boundary for
// Cloudflare/Turnstile challenge handling. It is intentionally descriptive so
// operators can inspect capability without probing a live protected website.
type CloudflareChallengeStatus struct {
	Supported      bool
	DefaultEnabled bool
	OptionName     string
	Message        string
	Err            error
}

func CloudflareChallengeBoundary() CloudflareChallengeStatus {
	return CloudflareChallengeStatus{
		Supported:      false,
		DefaultEnabled: false,
		OptionName:     "SolveCloudflare",
		Err:            ErrUnsupportedBrowserChallenge,
		Message:        "Cloudflare/Turnstile challenge solving is unsupported and disabled by default; shipping it requires a deterministic local challenge fixture, explicit tests, and operator-visible controls before goscrapling makes bypass claims.",
	}
}
