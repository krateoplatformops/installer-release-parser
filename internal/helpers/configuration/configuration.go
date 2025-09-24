package configuration

import (
	"flag"
	"strings"

	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/rs/zerolog/log"
)

type Configuration struct {
	InstallerChartRegistry         string   `json:"installerChartRegistry" yaml:"installerChartRegistry"`
	InstallerChartRepository       string   `json:"installerChartRepository" yaml:"installerChartRepository"`
	InstallerChartGithubRepository string   `json:"installerChartGithubRepository" yaml:"installerChartGithubRepository"`
	InstallerChartVersion          string   `json:"installerChartVersion" yaml:"installerChartVersion"`
	InstallerChartVersionPrevious  string   `json:"installerChartVersionPrevious" yaml:"installerChartVersionPrevious"`
	Tokens                         []string `json:"token" yaml:"token"`
	InstallerOrganization          string   `json:"installerOrganization" yaml:"installerOrganization"`
	Organizations                  []string `json:"organization" yaml:"organizations"`
	KrateoRepository               string   `json:"krateoRepository" yaml:"krateoRepository"`
}

func ParseConfig() Configuration {
	installerChartRegistry := flag.String("installerchartregistry",
		env.String("INSTALLER_CHART_REGISTRY", "https://charts.krateo.io/"), "Installer Chart Registry")

	installerChartRepository := flag.String("installerchartrepository",
		env.String("INSTALLER_CHART_REPOSITORY", "installer"), "Installer Chart Reporitory")

	installerChartGithubRepository := flag.String("installerchartgithubrepository",
		env.String("INSTALLER_CHART_GITHUB_REPOSITORY", "installer-chart"), "Installer Chart Github Reporitory")

	installerChartVersion := flag.String("installerchartversion",
		env.String("INSTALLER_CHART_VERSION", "2.5.1"), "Installer Chart Version")

	installerChartVersionPrevious := flag.String("installerchartversionprevious",
		env.String("INSTALLER_CHART_VERSION_PREVIOUS", "2.5.0"), "Installer Chart Version to generate the release notes from")

	tokens := flag.String("token",
		env.String("TOKEN", ""), "GitHub bearer/app token for the API")

	tokenList := strings.Split(*tokens, ",")

	installerOrganization := flag.String("installerorganization",
		env.String("INSTALLER_ORGANIZATION", "krateoplatformops"), "GitHub Organization to get/publish release notes for the installer")

	organization := flag.String("organizations",
		env.String("ORGANIZATIONS", "krateoplatformops,krateoplatformops-blueprints"), "Comma separetaed list of GitHub Organization to retrieve release notes from")
	log.Logger.Debug().Msgf("List of organizations: %s", *organization)

	organizations := strings.Split(*organization, ",")
	log.Logger.Debug().Msgf("Parsed list of organizations: %s", organizations)

	krateoRepository := flag.String("krateorepository",
		env.String("KRATEO_REPOSITORY", "krateo"), "Repository to append the release notes in /RELEASE_NOTES.md")

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
		Tokens:                         tokenList,
		InstallerOrganization:          *installerOrganization,
		Organizations:                  organizations,
		KrateoRepository:               *krateoRepository,
	}
}
