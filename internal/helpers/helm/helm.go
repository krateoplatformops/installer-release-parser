package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"

	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v3"

	"installer-release-parser/apis"
)

const (
	CHART_DIR = "./charts"
)

// ChartMetadata represents the structure of Chart.yaml
type chartMetadata struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	AppVersion string `yaml:"appVersion"`
}

func Pull(chart apis.Chart) error {
	settings := cli.New()

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), "default", os.Getenv("HELM_DRIVER"), log.Logger.Debug().Msgf); err != nil {
		return err
	}

	client := action.NewPullWithOpts(action.WithConfig(actionConfig))
	client.DestDir = CHART_DIR
	client.RepoURL = chart.Registry
	client.Version = chart.Version
	client.Untar = true
	client.Settings = settings

	result, err := client.Run(chart.Repository)
	if err != nil {
		return err
	}

	log.Debug().Msgf("%s: helm pull result: %s", chart.Repository, result)
	return nil
}

func ParseValues() (map[string]apis.Repoes, error) {
	installerFile, err := os.ReadFile(filepath.Join(CHART_DIR, "installer", "values.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var installerValues map[string]any
	if err := yaml.Unmarshal(installerFile, &installerValues); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	krateoplatformopsValues, ok := installerValues["krateoplatformops"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("krateoplatformops key not found or not a map")
	}

	result := map[string]apis.Repoes{}

	for topLevelKey := range krateoplatformopsValues {
		topLevelValue, ok := krateoplatformopsValues[topLevelKey].(map[string]any)
		if !ok {
			log.Warn().Msgf("Skipping %s: not a map", topLevelKey)
			continue
		}

		// Check if chart exists
		chartValue, hasChart := topLevelValue["chart"]
		if !hasChart {
			log.Warn().Msgf("Skipping %s: no chart field", topLevelKey)
			continue
		}

		chartMap, ok := chartValue.(map[string]any)
		if !ok {
			log.Warn().Msgf("Skipping %s: chart is not a map", topLevelKey)
			continue
		}

		// Safely extract chart fields
		chartName, ok := chartMap["name"].(string)
		if !ok {
			log.Warn().Msgf("Skipping %s: chart.name is not a string", topLevelKey)
			continue
		}

		chartVersion, ok := chartMap["version"].(string)
		if !ok {
			log.Warn().Msgf("Skipping %s: chart.version is not a string", topLevelKey)
			continue
		}

		chartRepository, ok := chartMap["repository"].(string)
		if !ok {
			log.Warn().Msgf("Skipping %s: chart.repository is not a string", topLevelKey)
			continue
		}

		// Safely extract image repository
		imageURL, err := getStringFromMap(topLevelValue, "image", "repository")
		imageURLParts := strings.Split(imageURL, "/")
		imageName := imageURLParts[len(imageURLParts)-1]
		if err != nil {
			log.Warn().Err(err).Msgf("%s: failed to get image.repository", topLevelKey)
			log.Info().Msgf("Checking %s for hardcoded value", topLevelKey)
			if value, ok := HARDCODED_REPOSITORIES[topLevelKey]; ok {
				imageName = value
			} else {
				log.Warn().Err(err).Msgf("Skipping %s: no hardcoded value found", topLevelKey)
				continue
			}
		}

		err = Pull(apis.Chart{
			Repository: chartName,
			Version:    chartVersion,
			Registry:   chartRepository,
		})
		if err != nil {
			log.Warn().Err(err).Msgf("Skipping %s: failed to download chart", topLevelKey)
			continue
		}

		appVersion, err := getAppVersionFromChart(chartName)
		if err != nil {
			log.Warn().Err(err).Msgf("Skipping %s: failed to obtain chart appVersion", topLevelKey)
			continue
		}

		result[topLevelKey] = apis.Repoes{
			ImageName: imageName,
			Chart: apis.Chart{
				Repository: chartName,
				Version:    chartVersion,
				AppVersion: appVersion,
				Registry:   chartRepository,
			},
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid entries found in krateoplatformops section")
	}

	return result, nil
}
