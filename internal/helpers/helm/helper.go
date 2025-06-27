package helm

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

var (
	HARDCODED_REPOSITORIES = map[string]string{
		"composableportalstarter": "composable-portal-starter",
		"composableportalbasic":   "composable-portal-basic",
		"finopspolicies":          "finops-moving-window-policy-chart",
		"crate":                   "cratedb-chart",
		"opa":                     "opa-chart",
	}
)

// Helper function to safely get string from nested map
func getStringFromMap(m map[string]any, keys ...string) (string, error) {
	current := m
	for i, key := range keys[:len(keys)-1] {
		next, ok := current[key].(map[string]any)
		if !ok {
			return "", fmt.Errorf("key path %v not found or not a map at step %d", keys, i)
		}
		current = next
	}

	finalKey := keys[len(keys)-1]
	value, ok := current[finalKey].(string)
	if !ok {
		return "", fmt.Errorf("final key %s not found or not a string", finalKey)
	}
	return value, nil
}

// getAppVersionFromChart reads the Chart.yaml file and extracts the appVersion
func getAppVersionFromChart(chartName string) (string, error) {
	chartPath := filepath.Join(CHART_DIR, chartName, "Chart.yaml")

	chartFile, err := os.ReadFile(chartPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Chart.yaml for %s: %w", chartName, err)
	}

	var metadata chartMetadata
	if err := yaml.Unmarshal(chartFile, &metadata); err != nil {
		return "", fmt.Errorf("failed to unmarshal Chart.yaml for %s: %w", chartName, err)
	}

	if metadata.AppVersion != "" {
		return metadata.AppVersion, nil
	} else {
		return metadata.Version, nil
	}
}
