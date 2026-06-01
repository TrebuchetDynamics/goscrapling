package curlcommand

import "testing"

func TestMethodWithBodyDataMakesDefaultPromotionExplicit(t *testing.T) {
	cases := []struct {
		name           string
		currentMethod  string
		explicitMethod bool
		want           string
	}{
		{name: "default get data promotes to post", currentMethod: "get", want: "post"},
		{name: "explicit get data stays get", currentMethod: "get", explicitMethod: true, want: "get"},
		{name: "explicit delete data stays delete", currentMethod: "delete", explicitMethod: true, want: "delete"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := methodWithBodyData(tc.currentMethod, tc.explicitMethod); got != tc.want {
				t.Fatalf("methodWithBodyData(%q, %v) = %q, want %q", tc.currentMethod, tc.explicitMethod, got, tc.want)
			}
		})
	}
}

func TestParsePreservesDollarPrefixedDataValues(t *testing.T) {
	request, err := Parse(`curl 'https://example.com/form' --data '$amount=10' --data-raw $'line=one'`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if request.Body != "$amount=10&line=one" {
		t.Fatalf("Body = %q, want literal dollars preserved unless they mark ANSI-C quoting", request.Body)
	}
}

func TestParsePreservesBackslashesInsideSingleQuotedValues(t *testing.T) {
	request, err := Parse(`curl 'https://example.com/form' --data 'path=C:\Temp\kit'`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if request.Body != `path=C:\Temp\kit` {
		t.Fatalf("Body = %q, want single-quoted curl value to preserve literal backslashes", request.Body)
	}
}

func TestParseSkipsShellLineContinuations(t *testing.T) {
	request, err := Parse("curl 'https://example.com/search' \\\n\t-H 'X-Trace: copied' \\\n\t--data-raw 'q=kit'")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if request.URL != "https://example.com/search" {
		t.Fatalf("URL = %q, want copied curl URL without line-continuation artifacts", request.URL)
	}
	if got := request.Headers.Get("X-Trace"); got != "copied" {
		t.Fatalf("Headers.Get(%q) = %q, want %q", "X-Trace", got, "copied")
	}
	if request.Body != "q=kit" {
		t.Fatalf("Body = %q, want data after shell line continuations", request.Body)
	}
}

func TestParsePreservesEmptyQuotedDataValue(t *testing.T) {
	request, err := Parse(`curl 'https://example.com/form' --data ''`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if request.Method != "post" {
		t.Fatalf("Method = %q, want post", request.Method)
	}
	if request.Body != "" {
		t.Fatalf("Body = %q, want empty quoted data preserved as an explicit empty body", request.Body)
	}
}

func TestParseHonorsExplicitGetWithBody(t *testing.T) {
	request, err := Parse(`curl -X GET 'https://example.com/search' --data 'q=kit'`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if request.Method != "get" {
		t.Fatalf("Method = %q, want explicit -X GET to stay get", request.Method)
	}
	if request.Body != "q=kit" {
		t.Fatalf("Body = %q, want data retained as request body without -G", request.Body)
	}
	if got := request.Params.Get("q"); got != "" {
		t.Fatalf("Params.Get(%q) = %q, want explicit GET body not converted to query params", "q", got)
	}
}

func TestParsePreservesRepeatedDataValues(t *testing.T) {
	t.Run("get data becomes query params without dropping earlier fields", func(t *testing.T) {
		request, err := Parse(`curl -G 'https://example.com/search?existing=1' --data 'q=kit' --data-raw 'page=2'`)
		if err != nil {
			t.Fatalf("Parse returned error: %v", err)
		}

		if request.Method != "get" {
			t.Fatalf("Method = %q, want get", request.Method)
		}
		if request.Body != "" {
			t.Fatalf("Body = %q, want empty body for curl -G", request.Body)
		}
		for key, want := range map[string]string{"existing": "1", "q": "kit", "page": "2"} {
			if got := request.Params.Get(key); got != want {
				t.Fatalf("Params.Get(%q) = %q, want %q (all params: %v)", key, got, want, request.Params)
			}
		}
	})

	t.Run("post data joins repeated flags in curl order", func(t *testing.T) {
		request, err := Parse(`curl 'https://example.com/form' --data 'q=kit' --data-raw 'page=2'`)
		if err != nil {
			t.Fatalf("Parse returned error: %v", err)
		}

		if request.Method != "post" {
			t.Fatalf("Method = %q, want post", request.Method)
		}
		if request.Body != "q=kit&page=2" {
			t.Fatalf("Body = %q, want repeated data joined", request.Body)
		}
	})
}
