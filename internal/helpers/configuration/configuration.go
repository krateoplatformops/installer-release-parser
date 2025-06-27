package configuration

import (
	"flag"

	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/rs/zerolog/log"
)

type Configuration struct {
	InstallerChartRegistry         string  `json:"installerChartRegistry" yaml:"installerChartRegistry"`
	InstallerChartRepository       string  `json:"installerChartRepository" yaml:"installerChartRepository"`
	InstallerChartGithubRepository string  `json:"installerChartGithubRepository" yaml:"installerChartGithubRepository"`
	InstallerChartVersion          string  `json:"installerChartVersion" yaml:"installerChartVersion"`
	InstallerChartVersionPrevious  string  `json:"installerChartVersionPrevious" yaml:"installerChartVersionPrevious"`
	Token                          *string `json:"token" yaml:"token"`
	Organization                   string  `json:"organization" yaml:"organization"`
}

func ParseConfig() Configuration {
	installerChartRegistry := flag.String("installerchartregistry",
		env.String("INSTALLER_CHART_REGISTRY", "https://charts.krateo.io/"), "Installer Chart Registry")

	installerChartRepository := flag.String("installerchartrepository",
		env.String("INSTALLER_CHART_REPOSITORY", "installer"), "Installer Chart Reporitory")

	installerChartGithubRepository := flag.String("installerchartgithubrepository",
		env.String("INSTALLER_CHART_GITHUB_REPOSITORY", "installer-chart"), "Installer Chart Github Reporitory")

	installerChartVersion := flag.String("installerchartversion",
		env.String("INSTALLER_CHART_VERSION", "2.5.0"), "Installer Chart Version")

	installerChartVersionPrevious := flag.String("installerchartversionprevious",
		env.String("INSTALLER_CHART_VERSION_PREVIOUS", "2.4.3"), "Installer Chart Version to generate the release notes from")

	token := flag.String("token",
		env.String("TOKEN", ""), "GitHub bearer/app token for the API")

	organization := flag.String("organization",
		env.String("ORGANIZATION", "krateoplatformops"), "GitHub Organization to retrieve release notes from")

	// Parse flags
	flag.Parse()

	// Now dereference after parsing
	log.Logger.Debug().Msgf("args %s", flag.Args())

	return Configuration{
		InstallerChartRegistry:         *installerChartRegistry,
		InstallerChartRepository:       *installerChartRepository,
		InstallerChartGithubRepository: *installerChartGithubRepository,
		InstallerChartVersion:          *installerChartVersion,
		InstallerChartVersionPrevious:  *installerChartVersionPrevious,
		Token:                          token,
		Organization:                   *organization,
	}
}
