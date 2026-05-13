package progress

import (
	"fmt"
	"strings"
)

func RenderBuilderLoopHandoff(p *Progress) string {
	meta := p.Meta.BuilderLoop
	var b strings.Builder
	b.WriteString("## Control Plane\n\n")
	fmt.Fprintf(&b, "- Entrypoint: `%s`\n", meta.Entrypoint)
	fmt.Fprintf(&b, "- Plan: `%s`\n", meta.Plan)
	if meta.CoverageLedger != "" {
		fmt.Fprintf(&b, "- Coverage ledger: `%s`\n", meta.CoverageLedger)
	}
	if meta.AgentQueue != "" {
		fmt.Fprintf(&b, "- Agent queue: `%s`\n", meta.AgentQueue)
	}
	fmt.Fprintf(&b, "- Progress schema: `%s`\n", meta.ProgressSchema)
	fmt.Fprintf(&b, "- Candidate source: `%s`\n", meta.CandidateSource)
	fmt.Fprintf(&b, "- Unit tests: `%s`\n", meta.UnitTest)
	if len(meta.CandidatePolicy) > 0 {
		b.WriteString("\n## Candidate Policy\n\n")
		for _, policy := range meta.CandidatePolicy {
			fmt.Fprintf(&b, "- %s\n", policy)
		}
	}
	return b.String()
}

func RenderAgentQueue(p *Progress) string {
	rows := AgentQueueRows(p)
	if len(rows) == 0 {
		return "_No unblocked contract rows are ready for implementation._\n"
	}
	var b strings.Builder
	for i, row := range rows {
		item := row.Item
		fmt.Fprintf(&b, "## %d. %s\n\n", i+1, item.Name)
		fmt.Fprintf(&b, "- Phase: `%s / %s`\n", row.PhaseKey, row.SubphaseKey)
		fmt.Fprintf(&b, "- Priority: `%s`\n", item.Priority)
		fmt.Fprintf(&b, "- Owner: `%s`\n", item.ExecutionOwner)
		fmt.Fprintf(&b, "- Size: `%s`\n", item.SliceSize)
		fmt.Fprintf(&b, "- Contract status: `%s`\n", item.ContractStatus)
		fmt.Fprintf(&b, "- Contract: %s\n", item.Contract)
		fmt.Fprintf(&b, "- Ready when: %s\n", joinOrDash(item.ReadyWhen))
		if len(item.NotReadyWhen) > 0 {
			fmt.Fprintf(&b, "- Not ready when: %s\n", joinOrDash(item.NotReadyWhen))
		}
		fmt.Fprintf(&b, "- Write scope: %s\n", joinCodeOrDash(item.WriteScope))
		fmt.Fprintf(&b, "- Test commands: %s\n", joinCodeOrDash(item.TestCommands))
		if item.NoTestRequired != "" {
			fmt.Fprintf(&b, "- No test required: %s\n", item.NoTestRequired)
		}
		fmt.Fprintf(&b, "- Acceptance: %s\n", joinOrDash(item.Acceptance))
		fmt.Fprintf(&b, "- Done signal: %s\n", joinOrDash(item.DoneSignal))
		fmt.Fprintf(&b, "- Source refs: %s\n\n", joinCodeOrDash(item.SourceRefs))
	}
	return b.String()
}

func RenderNextSlices(p *Progress) string {
	rows := AgentQueueRows(p)
	if len(rows) == 0 {
		return "_No next slices are currently ready._\n"
	}
	if len(rows) > 10 {
		rows = rows[:10]
	}
	var b strings.Builder
	b.WriteString("| Phase | Slice | Owner | Size | Contract status | Why now |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, row := range rows {
		item := row.Item
		fmt.Fprintf(&b, "| %s | %s | `%s` | `%s` | `%s` | %s |\n",
			escapeCell(row.PhaseKey+" / "+row.SubphaseKey),
			escapeCell(item.Name),
			escapeCell(item.ExecutionOwner),
			escapeCell(item.SliceSize),
			escapeCell(item.ContractStatus),
			escapeCell(whyNow(item)),
		)
	}
	return b.String()
}

func RenderBlockedSlices(p *Progress) string {
	rows := BlockedRows(p)
	if len(rows) == 0 {
		return "_No contract rows are currently blocked._\n"
	}
	var b strings.Builder
	b.WriteString("| Phase | Slice | Blocked by | Ready when | Unblocks |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, row := range rows {
		item := row.Item
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n",
			escapeCell(row.PhaseKey+" / "+row.SubphaseKey),
			escapeCell(item.Name),
			escapeCell(joinOrDash(item.BlockedBy)),
			escapeCell(joinOrDash(item.ReadyWhen)),
			escapeCell(joinOrDash(item.Unblocks)),
		)
	}
	return b.String()
}

func RenderUmbrellaCleanup(p *Progress) string {
	rows := UmbrellaRows(p)
	if len(rows) == 0 {
		return "_No umbrella rows need cleanup._\n"
	}
	var b strings.Builder
	b.WriteString("| Phase | Umbrella row | Owner | Not ready when |\n")
	b.WriteString("|---|---|---|---|\n")
	for _, row := range rows {
		item := row.Item
		fmt.Fprintf(&b, "| %s | %s | `%s` | %s |\n",
			escapeCell(row.PhaseKey+" / "+row.SubphaseKey),
			escapeCell(item.Name),
			escapeCell(item.ExecutionOwner),
			escapeCell(joinOrDash(item.NotReadyWhen)),
		)
	}
	return b.String()
}

func whyNow(item Item) string {
	if item.Priority == "P0" {
		return "P0 port parity row with complete handoff metadata."
	}
	if item.ContractStatus == "fixture_ready" {
		return "Fixture-ready row with complete handoff metadata."
	}
	return "Contract metadata is present and the row is unblocked."
}

func joinOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}

func joinCodeOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, "`"+value+"`")
	}
	return strings.Join(out, ", ")
}

func escapeCell(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")
	if value == "" {
		return "-"
	}
	return value
}
