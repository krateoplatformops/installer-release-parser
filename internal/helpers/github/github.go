package github

import (
	"context"
	"fmt"
	"installer-release-parser/apis"

	"github.com/google/go-github/v72/github"
	"github.com/rs/zerolog/log"
)

// This function assumes that all repositories listed in the installer exist and are tagged with the installer versions
func GetReleaseNotes(charts map[string]apis.Repoes, token *string, owner string) string {
	client := github.NewClient(nil)

	if token != nil {
		client = client.WithAuthToken(*token)
	}

	finalReleaseNotes := ""

	for _, chart := range charts {
		if chart.AppVersionPrevious == "" {
			log.Warn().Msg("empty previous version, using automatic option")
		}
		log.Info().Msgf("Generating release notes for %s with tag range %s ... %s", chart.ImageName, chart.AppVersionPrevious, chart.AppVersion)
		release, _, err := client.Repositories.GenerateReleaseNotes(context.Background(), owner, chart.ImageName, &github.GenerateNotesOptions{
			TagName:         chart.AppVersion,
			PreviousTagName: &chart.AppVersionPrevious,
		})
		if err != nil {
			log.Warn().Err(err).Msgf("Skipping %s: there was an error generating the release", chart.ImageName)
		} else {
			finalReleaseNotes += fmt.Sprintf("# %s v%s\n## What's Changed\n%s\n\n", chart.ImageName, chart.AppVersion, formatReleaseNotes(release.Body))
		}
		//
	}

	return finalReleaseNotes
}
