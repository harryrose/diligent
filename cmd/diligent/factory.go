package main

import (
	"fmt"
	"github.com/senseyeio/diligent/python/pip"
	"github.com/senseyeio/diligent/python/pypi"

	"path/filepath"

	"github.com/senseyeio/diligent"
	"github.com/senseyeio/diligent/dep"
	"github.com/senseyeio/diligent/github"
	"github.com/senseyeio/diligent/go"
	"github.com/senseyeio/diligent/govendor"
	"github.com/senseyeio/diligent/npm"
)

var (
	gh        = github.New("https://api.github.com")
	goLG      = _go.NewLicenseGetter(gh)
	npmAPIURL = "https://registry.npmjs.org"
	pypiAPIURL = "https://pypi.python.org"
)

var depers = []diligent.Deper{
	npm.New(npmAPIURL),
	govendor.New(goLG),
	dep.New(goLG),
	&pip.Deper{
		Client: &pypi.Client{
			BaseURL: pypiAPIURL,
		},
	},
}

func getDeper(path string) (diligent.Deper, error) {
	filename := filepath.Base(path)
	for _, deper := range depers {
		if deper.IsCompatible(filename) {
			return deper, nil
		}
	}
	return nil, fmt.Errorf("Diligent does not know how to process '%s' files", filename)
}
