package github

import (
	"fmt"
	"regexp"
	"strings"
)

func formatReleaseNotes(input string) string {
	lines := strings.Split(input, "\n")

	var features, fixes, docs, other []string
	var changelogLink string
	commitRegex := regexp.MustCompile(`^\* (.*) by (@\S+) in (https://github\.com/[^)]+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		matches := commitRegex.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}

		message, author, url := matches[1], matches[2], matches[3]
		formatted := fmt.Sprintf("- %s ([link](%s)) by %s", message, url, author)

		switch {
		case strings.HasPrefix(message, "feat"):
			features = append(features, formatted)
		case strings.HasPrefix(message, "fix"):
			fixes = append(fixes, formatted)
		case strings.HasPrefix(message, "docs"):
			docs = append(docs, formatted)
		case strings.HasPrefix(message, "**Full Changelog"):
			changelogLink = message
		case regexp.MustCompile(`^.+`).MatchString(message):
			other = append(other, formatted)
		}
	}

	sb := strings.Builder{}
	if len(features) > 0 {
		sb.WriteString("\n### ✨ Features\n")
		sb.WriteString(strings.Join(features, "\n"))
		sb.WriteString("\n")
	}
	if len(fixes) > 0 {
		sb.WriteString("\n### 🐛 Bug Fixes\n")
		sb.WriteString(strings.Join(fixes, "\n"))
		sb.WriteString("\n")
	}
	if len(docs) > 0 {
		sb.WriteString("\n### 📚 Documentation\n")
		sb.WriteString(strings.Join(docs, "\n"))
		sb.WriteString("\n")
	}
	if len(other) > 0 {
		sb.WriteString("\n### 🔧 Other Changes\n")
		sb.WriteString(strings.Join(other, "\n"))
		sb.WriteString("\n")
	}

	sb.WriteString(changelogLink)
	sb.WriteString("\n")

	return sb.String()
}
