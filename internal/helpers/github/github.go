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
func GetReleaseNotes(charts map[string]apis.Repoes, tokens []string, owners []string) string {
	client := github.NewClient(nil)

	clients := map[string]*github.Client{}
	if len(tokens) != 0 {
		for i, token := range tokens {
			clients[owners[i]] = client.WithAuthToken(token)
		}
	}

	finalReleaseNotes := ""

	for _, chart := range charts {
		if chart.AppVersionPrevious == "" {
			log.Warn().Msg("empty previous version, using automatic option")
		}
		for _, owner := range owners {
			log.Info().Msgf("Generating release notes for %s with tag range %s ... %s", chart.ImageName, chart.AppVersionPrevious, chart.AppVersion)
			release, response, err := clients[owner].Repositories.GenerateReleaseNotes(context.Background(), owner, chart.ImageName, &github.GenerateNotesOptions{
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
					release, response, errr := clients[owner].Repositories.GenerateReleaseNotes(context.Background(), owner, value, &github.GenerateNotesOptions{
						TagName: chart.Version,
					})
					if errr != nil {
						log.Warn().Err(err).Msgf("%s: there was an error generating the release for the chart", value)
						bodyData, _ := io.ReadAll(response.Body)
						log.Warn().Msgf("Body %s", string(bodyData))
					} else {
						finalReleaseNotes += fmt.Sprintf("## %s v%s\n### What's Changed\n%s\n\n", value, chart.Version, formatReleaseNotes(release.Body))
						break
					}
				}
			} else {
				finalReleaseNotes += fmt.Sprintf("## %s v%s\n### What's Changed\n%s\n\n", chart.ImageName, chart.AppVersion, formatReleaseNotes(release.Body))
				break
			}
		}
	}

	return finalReleaseNotes
}

func CreateInstallerRelease(releaseNotes string, config configuration.Configuration) {
	client := github.NewClient(nil)

	clients := map[string]*github.Client{}
	if len(config.Tokens) != 0 {
		for i, token := range config.Tokens {
			clients[config.Organizations[i]] = client.WithAuthToken(token)
		}
	}

	release, _, err := clients[config.InstallerOrganization].Repositories.GetReleaseByTag(context.Background(), config.InstallerOrganization, config.InstallerChartGithubRepository, config.InstallerChartVersion)
	if err != nil {
		log.Info().Msgf("Release not found for tag %s", config.InstallerChartVersion)
		_, response, errr := clients[config.InstallerOrganization].Repositories.CreateRelease(context.Background(), config.InstallerOrganization, config.InstallerChartGithubRepository, &github.RepositoryRelease{
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
		_, response, errr := clients[config.InstallerOrganization].Repositories.EditRelease(context.Background(), config.InstallerOrganization, config.InstallerChartGithubRepository, *release.ID, release)
		if errr != nil {
			log.Error().Err(err).Msgf("could not edit release")
			bodyData, _ := io.ReadAll(response.Body)
			log.Error().Msgf("Body %s", string(bodyData))
		} else {
			log.Info().Msgf("Release edited for tag %s", config.InstallerChartVersion)
		}
	}

	// --- Append release notes to RELEASE_NOTES.md in config.KrateoRepository ---
	ctx := context.Background()

	// Get the current contents of RELEASE_NOTES.md
	fileContent, _, resp, err := clients[config.InstallerOrganization].Repositories.GetContents(
		ctx,
		config.InstallerOrganization,
		config.KrateoRepository,
		"RELEASE_NOTES.md",
		nil,
	)

	var newContent string
	var sha *string

	if err != nil {
		// If file not found, start a new one
		if resp != nil && resp.StatusCode == 404 {
			log.Info().Msg("RELEASE_NOTES.md not found, creating a new one")
			newContent = fmt.Sprintf("# Release %s\n\n%s\n\n", config.InstallerChartVersion, releaseNotes)
		} else {
			log.Error().Err(err).Msg("could not retrieve RELEASE_NOTES.md")
			return
		}
	} else {
		// Append to existing contents
		content, err := fileContent.GetContent()
		if err != nil {
			log.Error().Err(err).Msg("could not decode RELEASE_NOTES.md")
			return
		}
		newContent = fmt.Sprintf("# Release %s\n\n%s\n<br><br>\n%s", config.InstallerChartVersion, releaseNotes, content)
		sha = fileContent.SHA
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("chore: update release notes for %s", config.InstallerChartVersion)),
		Content: []byte(newContent),
		SHA:     sha,
	}

	_, _, err = clients[config.InstallerOrganization].Repositories.UpdateFile(
		ctx,
		config.InstallerOrganization,
		config.KrateoRepository,
		"RELEASE_NOTES.md",
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("could not update RELEASE_NOTES.md")
	} else {
		log.Info().Msgf("RELEASE_NOTES.md updated for version %s", config.InstallerChartVersion)
	}
}
