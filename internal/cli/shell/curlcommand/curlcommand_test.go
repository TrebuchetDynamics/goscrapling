package curlcommand

import "testing"

func TestParsePreservesDollarPrefixedDataValues(t *testing.T) {
	request, err := Parse(`curl 'https://example.com/form' --data '$amount=10' --data-raw $'line=one'`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if request.Body != "$amount=10&line=one" {
		t.Fatalf("Body = %q, want literal dollars preserved unless they mark ANSI-C quoting", request.Body)
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
