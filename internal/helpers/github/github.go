package github

import (
	"context"
	"fmt"
	"installer-release-parser/apis"
	"installer-release-parser/internal/helpers/configuration"
	"installer-release-parser/internal/helpers/helm"
	"io"

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
		release, response, err := client.Repositories.GenerateReleaseNotes(context.Background(), owner, chart.ImageName, &github.GenerateNotesOptions{
			TagName:         chart.AppVersion,
			PreviousTagName: &chart.AppVersionPrevious,
		})
		if err != nil {
			log.Warn().Err(err).Msgf("%s: there was an error generating the release", chart.ImageName)
			bodyData, _ := io.ReadAll(response.Body)
			log.Warn().Msgf("Body %s", string(bodyData))
			log.Warn().Msg("Container probably missing, trying hardcoded values with chart version...")
			if value, ok := helm.HARDCODED_REPOSITORIES[chart.ImageName]; ok {
				log.Info().Msgf("Generating release notes for %s with tag range %s ... %s", value, chart.AppVersionPrevious, chart.AppVersion)
				release, response, errr := client.Repositories.GenerateReleaseNotes(context.Background(), owner, value, &github.GenerateNotesOptions{
					TagName: chart.Version,
				})
				if errr != nil {
					log.Warn().Err(err).Msgf("%s: there was an error generating the release for the chart", value)
					bodyData, _ := io.ReadAll(response.Body)
					log.Warn().Msgf("Body %s", string(bodyData))
				} else {
					finalReleaseNotes += fmt.Sprintf("## %s v%s\n### What's Changed\n%s\n\n", value, chart.Version, formatReleaseNotes(release.Body))
				}
			}
		} else {
			finalReleaseNotes += fmt.Sprintf("## %s v%s\n### What's Changed\n%s\n\n", chart.ImageName, chart.AppVersion, formatReleaseNotes(release.Body))
		}
	}

	return finalReleaseNotes
}

func CreateInstallerRelease(releaseNotes string, config configuration.Configuration) {
	client := github.NewClient(nil)

	if config.Token != nil {
		client = client.WithAuthToken(*config.Token)
	}

	release, _, err := client.Repositories.GetReleaseByTag(context.Background(), config.Organization, config.InstallerChartGithubRepository, config.InstallerChartVersion)
	if err != nil {
		log.Info().Msgf("Release not found for tag %s", config.InstallerChartVersion)
		_, response, errr := client.Repositories.CreateRelease(context.Background(), config.Organization, config.InstallerChartGithubRepository, &github.RepositoryRelease{
			TagName:    &config.InstallerChartVersion,
			Name:       stringPointer(fmt.Sprintf("Release Notes For Krateo %s ... %s\n", config.InstallerChartVersionPrevious, config.InstallerChartVersion)),
			Body:       &releaseNotes,
			MakeLatest: stringPointer("true"),
		})
		if errr != nil {
			log.Error().Err(err).Msgf("could not create release")
			bodyData, _ := io.ReadAll(response.Body)
			log.Error().Msgf("Body %s", string(bodyData))
		} else {
			log.Info().Msgf("Release created for tag %s", config.InstallerChartVersion)
		}
	} else {
		release.Body = &releaseNotes
		_, response, errr := client.Repositories.EditRelease(context.Background(), config.Organization, config.InstallerChartGithubRepository, *release.ID, release)
		if errr != nil {
			log.Error().Err(err).Msgf("could not edit release")
			bodyData, _ := io.ReadAll(response.Body)
			log.Error().Msgf("Body %s", string(bodyData))
		} else {
			log.Info().Msgf("Release edited for tag %s", config.InstallerChartVersion)
		}
	}

}
