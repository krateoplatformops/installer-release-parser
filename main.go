package main

import (
	"fmt"
	"installer-release-parser/apis"
	"installer-release-parser/internal/helpers/configuration"
	"installer-release-parser/internal/helpers/github"
	"installer-release-parser/internal/helpers/helm"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("Starting up")

	log.Info().Msg("Parsing configuration")
	config := configuration.ParseConfig()

	// Pull the current installer chart
	log.Info().Msg("Downloading current installer chart...")
	err := helm.Pull(apis.Chart{
		Registry:   config.InstallerChartRegistry,
		Repository: config.InstallerChartRepository,
		Version:    config.InstallerChartVersion,
	})
	if err != nil {
		log.Error().Err(err).Msg("there was an error while pulling the current installer chart")
		cleanup()
		return
	}

	// Pull all charts to get the appVersion
	log.Info().Msg("Downloading all charts...")
	allCharts, err := helm.ParseValues()
	if err != nil {
		log.Error().Err(err).Msg("there was an error while parsing all repositories from the installer chart values file")
		cleanup()
		return
	}
	log.Debug().Msg("=== Current Installer Versions")
	for key := range allCharts {
		log.Debug().Msgf("%s: %s", key, allCharts[key])
	}

	// Remove current release
	cleanup()

	// Pull the previous installer chart
	log.Info().Msg("Downloading previous installer chart...")
	err = helm.Pull(apis.Chart{
		Registry:   config.InstallerChartRegistry,
		Repository: config.InstallerChartRepository,
		Version:    config.InstallerChartVersionPrevious,
	})
	if err != nil {
		log.Error().Err(err).Msg("there was an error while pulling the previous installer chart")
		cleanup()
		return
	}
	// Pull all previous charts to get the appVersion
	log.Info().Msg("Downloading all charts...")
	allPreviousCharts, err := helm.ParseValues()
	if err != nil {
		log.Error().Err(err).Msg("there was an error while parsing all repositories from the installer chart values file")
		cleanup()
		return
	}
	log.Debug().Msg("=== Previous Installer Versions")
	for key := range allPreviousCharts {
		log.Debug().Msgf("%s: %s", key, allPreviousCharts[key])
	}

	// Get the removed charts
	removedChartsTextBuilder := strings.Builder{}
	removedChartsTextBuilder.WriteString("## Removed Charts\n")
	for key := range allPreviousCharts {
		if _, ok := allCharts[key]; !ok {
			removedChartsTextBuilder.WriteString(fmt.Sprintf("- %s v%s: Removed\n", allPreviousCharts[key].ImageName, allPreviousCharts[key].AppVersion))
		}
	}

	// Get the version changes
	allRangeCharts := map[string]apis.Repoes{}
	for key := range allCharts {
		if _, ok := allPreviousCharts[key]; ok {
			if allCharts[key].ImageName != allPreviousCharts[key].ImageName {
				removedChartsTextBuilder.WriteString(fmt.Sprintf("- %s v%s: Removed\n", allPreviousCharts[key].ImageName, allPreviousCharts[key].AppVersion))
				allRangeCharts[key] = apis.Repoes{
					ImageName: allCharts[key].ImageName,
					Chart: apis.Chart{
						Repository: allCharts[key].Chart.Repository,
						Version:    allCharts[key].Chart.Version,
						AppVersion: allCharts[key].Chart.AppVersion,
						Registry:   allCharts[key].Chart.Registry,
					},
				}
			} else {
				allRangeCharts[key] = apis.Repoes{
					ImageName: allCharts[key].ImageName,
					Chart: apis.Chart{
						Repository:         allCharts[key].Chart.Repository,
						Version:            allCharts[key].Chart.Version,
						AppVersion:         allCharts[key].Chart.AppVersion,
						Registry:           allCharts[key].Chart.Registry,
						AppVersionPrevious: allPreviousCharts[key].Chart.AppVersion,
					},
				}
			}
		} else {
			allRangeCharts[key] = apis.Repoes{
				ImageName: allCharts[key].ImageName,
				Chart: apis.Chart{
					Repository: allCharts[key].Chart.Repository,
					Version:    allCharts[key].Chart.Version,
					AppVersion: allCharts[key].Chart.AppVersion,
					Registry:   allCharts[key].Chart.Registry,
				},
			}
		}
	}

	removedChartsText := removedChartsTextBuilder.String()
	if removedChartsText == "## Removed Charts\n" {
		removedChartsTextBuilder.WriteString("Nothing removed\n")
		removedChartsText = removedChartsTextBuilder.String()
	}

	// Call the Github API to get the release notes
	// If config.CreateReleases is set to true, create the release notes for the tag appVersion (if it does not exist)
	log.Info().Msg("Generating release notes...")
	finalReleaseNotes := fmt.Sprintf("%s\n%s", removedChartsText, github.GetReleaseNotes(allRangeCharts, config.Token, config.Organization))

	// Write the result to file
	log.Info().Msg("Writing the release notes to file...")
	err = os.WriteFile("./release_notes.md", []byte(finalReleaseNotes), 0644)
	if err != nil {
		log.Error().Err(err).Msg("there was an error while writing the release notes to file")
		cleanup()
		return
	}

	// Publish the release notes on a github release for the given repository
	log.Info().Msgf("Publishing release on installer repository %s/%s:%s", config.Organization, config.InstallerChartGithubRepository, config.InstallerChartVersion)
	github.CreateInstallerRelease(finalReleaseNotes, config)
	cleanup()
}

// Cleanup downloaded charts
func cleanup() {
	os.RemoveAll("./charts/")
}
