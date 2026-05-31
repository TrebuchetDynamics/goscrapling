package browser

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrUnsupportedBrowserChallenge = errors.New("browser challenge solving unsupported")

const browserStealthUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"

// BrowserStealthOptions are explicit browser fingerprint controls. They are not
// enabled by normal browser fetching, and they do not claim automatic anti-bot
// bypass behavior.
type BrowserStealthOptions struct {
	Enabled         bool
	GenerateHeaders bool
	GoogleReferer   bool
	HideCanvas      bool
	BlockWebRTC     bool
	DisableWebGL    bool
	SolveCloudflare bool
}

func normalizeBrowserStealthOptions(stealth BrowserStealthOptions) BrowserStealthOptions {
	if stealth.GenerateHeaders || stealth.GoogleReferer || stealth.HideCanvas || stealth.BlockWebRTC || stealth.DisableWebGL || stealth.SolveCloudflare {
		stealth.Enabled = true
	}
	return stealth
}

func validateBrowserStealth(stealth BrowserStealthOptions) error {
	if stealth.SolveCloudflare {
		return fmt.Errorf("%w: %w: automatic Cloudflare/Turnstile challenge solving is not implemented", ErrBrowserOptions, ErrUnsupportedBrowserChallenge)
	}
	return nil
}

func applyBrowserStealthHeaders(headers http.Header, stealth BrowserStealthOptions) {
	if !stealth.Enabled {
		return
	}
	if stealth.GenerateHeaders {
		setBrowserHeaderDefault(headers, "User-Agent", browserStealthUserAgent)
		setBrowserHeaderDefault(headers, "Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		setBrowserHeaderDefault(headers, "Accept-Language", "en-US,en;q=0.9")
		setBrowserHeaderDefault(headers, "Upgrade-Insecure-Requests", "1")
	}
	if stealth.GoogleReferer {
		headers.Set("Referer", "https://www.google.com/")
	}
}

func setBrowserHeaderDefault(headers http.Header, key, value string) {
	if headers.Get(key) != "" {
		return
	}
	headers.Set(key, value)
}

func browserStealthExtraFlags(stealth BrowserStealthOptions, current []string) []string {
	flags := append([]string(nil), current...)
	if !stealth.Enabled {
		return flags
	}
	if stealth.BlockWebRTC {
		flags = appendUniqueBrowserFlag(flags, "--force-webrtc-ip-handling-policy=disable_non_proxied_udp")
	}
	if stealth.DisableWebGL {
		flags = appendUniqueBrowserFlag(flags, "--disable-features=WebGL,WebGL2")
	}
	return flags
}

func appendUniqueBrowserFlag(flags []string, flag string) []string {
	for _, existing := range flags {
		if existing == flag {
			return flags
		}
	}
	return append(flags, flag)
}

func mergeBrowserStealthOptions(defaults, overrides BrowserStealthOptions) BrowserStealthOptions {
	merged := normalizeBrowserStealthOptions(defaults)
	overrides = normalizeBrowserStealthOptions(overrides)
	if overrides.Enabled {
		merged.Enabled = true
	}
	if overrides.GenerateHeaders {
		merged.GenerateHeaders = true
	}
	if overrides.GoogleReferer {
		merged.GoogleReferer = true
	}
	if overrides.HideCanvas {
		merged.HideCanvas = true
	}
	if overrides.BlockWebRTC {
		merged.BlockWebRTC = true
	}
	if overrides.DisableWebGL {
		merged.DisableWebGL = true
	}
	if overrides.SolveCloudflare {
		merged.SolveCloudflare = true
	}
	return merged
}

func browserStealthInitScript(request BrowserRequest) string {
	if !request.Stealth.Enabled || !request.Stealth.HideCanvas {
		return ""
	}
	return `(() => {
  const patchCanvas = (prototype, method) => {
    const original = prototype && prototype[method];
    if (typeof original !== "function") return;
    Object.defineProperty(prototype, method, {
      value: function(...args) {
        try {
          const context = this.getContext && this.getContext("2d");
          if (context) {
            const width = Math.min(this.width || 1, 8);
            const height = Math.min(this.height || 1, 8);
            const image = context.getImageData(0, 0, width, height);
            for (let i = 0; i < image.data.length; i += 4) image.data[i] = image.data[i] ^ 1;
            context.putImageData(image, 0, 0);
          }
        } catch (_) {}
        return original.apply(this, args);
      }
    });
  };
  patchCanvas(HTMLCanvasElement.prototype, "toDataURL");
  patchCanvas(HTMLCanvasElement.prototype, "toBlob");
})();`
}
