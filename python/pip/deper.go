package pip

import (
	"fmt"
	"github.com/senseyeio/diligent"
	"github.com/senseyeio/diligent/python/pypi"
)

type ProjectMetadataFetcher interface {
	ProjectMetadata(projectName string, projectVersion string) (*pypi.ProjectMetadata, error)
}

type Deper struct {
	Client ProjectMetadataFetcher
}

func (*Deper) Name() string {
	return "pip"
}

// Dependencies interrogates the manifest file and returns the licenses associated with each dependency
// If a single dependency cannot be processed, a warning should be returned
// If no dependencies can be processed, an error should be returned
func (d *Deper) Dependencies(file []byte) ([]diligent.Dep, []diligent.Warning, error) {
	var reqs []Requirement
	err := Unmarshal(file, &reqs)

	if err != nil {
		return nil, nil, err
	}

	deps := make([]diligent.Dep, 0, len(reqs))
	var warns []diligent.Warning

	for _, req := range reqs {
		meta, err := d.Client.ProjectMetadata(req.ProjectName, req.Version)
		if err != nil {
			warns = append(warns, &warning{project: req.ProjectName, reason: err.Error()})
			continue
		}

		if meta.Info.License == "" {
			warns = append(warns, &warning{project: req.ProjectName, reason: "empty license field"})
			continue
		}

		lic, err := diligent.GetLicenseFromIdentifier(meta.Info.License)

		if err != nil {
			warns = append(warns, &warning{project: req.ProjectName, reason: err.Error()})
			continue
		}

		deps = append(deps, diligent.Dep{
			Name: req.ProjectName,
			License: lic,
		})
	}

	return deps, warns, nil
}

type warning struct {
	project string
	reason string
}

func (w *warning) Warning() string {
	return fmt.Sprintf("%s: %s", w.project, w.reason)
}

// IsCompatible should return true if the Deper can handle the provided manifest file
func (d *Deper) IsCompatible(filename string) bool {
	return filename == "requirements.txt"
}