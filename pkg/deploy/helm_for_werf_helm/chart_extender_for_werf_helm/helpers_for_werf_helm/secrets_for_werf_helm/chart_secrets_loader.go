package secrets_for_werf_helm

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"sigs.k8s.io/yaml"

	chart "github.com/werf/3p-helm-for-werf-helm/pkg/chart"
	chartutil "github.com/werf/3p-helm-for-werf-helm/pkg/chartutil"
	"github.com/werf/common-go/pkg/util"
	secret "github.com/werf/nelm-for-werf-helm/pkg/secret"
)

const (
	DefaultSecretValuesFileName = "secret-values.yaml"
	SecretDirName               = "secret"
)

func GetDefaultSecretValuesFile(loadedChartFiles []*chart.ChartExtenderBufferedFile) *chart.ChartExtenderBufferedFile {
	for _, file := range loadedChartFiles {
		if file.Name == DefaultSecretValuesFileName {
			return file
		}
	}

	return nil
}

func GetSecretDirFiles(loadedChartFiles []*chart.ChartExtenderBufferedFile) []*chart.ChartExtenderBufferedFile {
	var res []*chart.ChartExtenderBufferedFile

	for _, file := range loadedChartFiles {
		if !util.IsSubpathOfBasePath(SecretDirName, file.Name) {
			continue
		}
		res = append(res, file)
	}

	return res
}

func LoadChartSecretValueFiles(
	chartDir string,
	secretDirFiles []*chart.ChartExtenderBufferedFile,
	encoder *secret.YamlEncoder,
) (map[string]interface{}, error) {
	var res map[string]interface{}

	for _, file := range secretDirFiles {
		decodedData, err := encoder.DecryptYamlData(file.Data)
		if err != nil {
			return nil, fmt.Errorf("cannot decode file %q secret data: %w", filepath.Join(chartDir, file.Name), err)
		}

		rawValues := map[string]interface{}{}
		if err := yaml.Unmarshal(decodedData, &rawValues); err != nil {
			return nil, fmt.Errorf("cannot unmarshal secret values file %s: %w", filepath.Join(chartDir, file.Name), err)
		}

		res = chartutil.CoalesceTables(rawValues, res)
	}

	return res, nil
}

func LoadChartSecretDirFilesData(
	chartDir string,
	secretFiles []*chart.ChartExtenderBufferedFile,
	encoder *secret.YamlEncoder,
) (map[string]string, error) {
	res := make(map[string]string)

	for _, file := range secretFiles {
		if !util.IsSubpathOfBasePath(SecretDirName, file.Name) {
			continue
		}

		decodedData, err := encoder.Decrypt([]byte(strings.TrimRightFunc(string(file.Data), unicode.IsSpace)))
		if err != nil {
			return nil, fmt.Errorf("error decoding %s: %w", filepath.Join(chartDir, file.Name), err)
		}

		relPath := util.GetRelativeToBaseFilepath(SecretDirName, file.Name)
		res[filepath.ToSlash(relPath)] = string(decodedData)
	}

	return res, nil
}
