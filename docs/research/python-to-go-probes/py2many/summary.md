# py2many Static Reference Probe

Date: 2026-05-14

Purpose: reference-only Python-to-Go static translation evidence for the goscrapling upstream app map.

Command shape: `py2many --go <python-file>`

Source: https://github.com/py2many/py2many

Policy:

- Generated output is reference evidence only.
- Generated output must not be copied into production Go packages.
- Parity still requires progress rows and Go tests.

## Inputs

| Input | Output | Result |
|---|---|---|
| `references/Scrapling/scrapling/parser.py` | `docs/research/python-to-go-probes/py2many/go/parser.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/core/custom_types.py` | `docs/research/python-to-go-probes/py2many/go/core__custom_types.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/core/shell.py` | `docs/research/python-to-go-probes/py2many/go/core__shell.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/core/ai.py` | `docs/research/python-to-go-probes/py2many/go/core__ai.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/cli.py` | `docs/research/python-to-go-probes/py2many/go/cli.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/engines/static.py` | `docs/research/python-to-go-probes/py2many/go/engines__static.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/engines/toolbelt/proxy_rotation.py` | `docs/research/python-to-go-probes/py2many/go/engines__toolbelt__proxy_rotation.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/engines/toolbelt/navigation.py` | `docs/research/python-to-go-probes/py2many/go/engines__toolbelt__navigation.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/engines/_browsers/_validators.py` | `docs/research/python-to-go-probes/py2many/go/engines___browsers___validators.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/spiders/robotstxt.py` | `docs/research/python-to-go-probes/py2many/go/spiders__robotstxt.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/spiders/cache.py` | `docs/research/python-to-go-probes/py2many/go/spiders__cache.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/spiders/checkpoint.py` | `docs/research/python-to-go-probes/py2many/go/spiders__checkpoint.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/spiders/links.py` | `docs/research/python-to-go-probes/py2many/go/spiders__links.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/spiders/templates/crawler.py` | `docs/research/python-to-go-probes/py2many/go/spiders__templates__crawler.go.txt` | `missing-tool` |
| `references/Scrapling/scrapling/spiders/templates/sitemap.py` | `docs/research/python-to-go-probes/py2many/go/spiders__templates__sitemap.go.txt` | `missing-tool` |
