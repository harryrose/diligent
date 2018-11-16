package pip

import (
	"errors"
	"github.com/senseyeio/diligent"
	"github.com/senseyeio/diligent/python/pypi"
	"reflect"
	"testing"
)

func TestDeper_Name(t *testing.T) {
	const expected = "pip"

	uut := &Deper{}
	if got := uut.Name(); got != expected {
		t.Logf("expected %s to equal %s", got, expected)
		t.Fail()
	}
}

func TestDeper_IsCompatible(t *testing.T) {
	t.Run("requirements.txt", func(t *testing.T) {
		uut := &Deper{}

		if !uut.IsCompatible("requirements.txt") {
			t.Logf("should return true when the filename is 'requirements.txt'")
			t.Fail()
		}
	})

	t.Run("not-requirements.txt", func(t *testing.T) {
		uut := &Deper{}

		if uut.IsCompatible("not-requirements.txt") {
			t.Logf("should return false when the filename is not 'requirements.txt'")
			t.Fail()
		}
	})
}

type mockClient struct {
	ProjectMetadataFn func(projectName string, projectVersion string) (*pypi.ProjectMetadata, error)
}

func (m *mockClient) ProjectMetadata(projectName string, projectVersion string) (*pypi.ProjectMetadata, error) {
	return m.ProjectMetadataFn(projectName, projectVersion)
}

func TestDeper_Dependencies(t *testing.T) {
	const testProject = "myTest"
	const testVersion = "1.2.3"
	const validData = testProject + "==" + testVersion
	const invalidData = "^-^"

	license := diligent.GetLicenses()[0]

	testCases := []struct{
		name string
		input string
		mockClient *mockClient
		expectedDeps []diligent.Dep
		expectedWarns []diligent.Warning
		errorExpected bool
	}{
		{
			name:"unmarshal error",
			input: invalidData,
			errorExpected: true,
		},
		{
			name:"pypi error",
			input: validData,
			mockClient: &mockClient{
				ProjectMetadataFn: func(projectName string, projectVersion string) (*pypi.ProjectMetadata, error) {
					if projectName != testProject {
						t.Logf("expected project name, '%s', to equal '%s'", projectName, testProject)
						t.Fail()
					}

					if projectVersion != testVersion {
						t.Logf("expected project version, '%s', to equal '%s'", projectVersion, testVersion)
						t.Fail()
					}

					return nil, errors.New("test error")
				},
			},
			expectedDeps: []diligent.Dep{},
			expectedWarns: []diligent.Warning{
				&warning{ project: testProject, reason: "test error" },
			},
		},
		{
			name:"missing license",
			input: validData,
			mockClient: &mockClient{
				ProjectMetadataFn: func(_ string, _ string) (*pypi.ProjectMetadata, error) {
					return &pypi.ProjectMetadata{}, nil
				},
			},
			expectedDeps: []diligent.Dep{},
			expectedWarns: []diligent.Warning{
				&warning{ project: testProject, reason: "empty license field" },
			},
		},
		{
			name:"unknown license",
			input: validData,
			mockClient: &mockClient{
				ProjectMetadataFn: func(_ string, _ string) (*pypi.ProjectMetadata, error) {
					return &pypi.ProjectMetadata{
						Info:pypi.ProjectInfo{
							License: "top-secret!",
						},
					}, nil
				},
			},
			expectedDeps: []diligent.Dep{},
			expectedWarns: []diligent.Warning{
				&warning{ project: testProject, reason: "license identifier top-secret! is not known to diligent" },
			},
		},
		{
			name:"ok",
			input: validData,
			mockClient: &mockClient{
				ProjectMetadataFn: func(_ string, _ string) (*pypi.ProjectMetadata, error) {
					return &pypi.ProjectMetadata{
						Info:pypi.ProjectInfo{
							License: license.Identifier,
						},
					}, nil
				},
			},

			expectedDeps: []diligent.Dep{
				{
					Name: testProject,
					License:  license,
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T){
			t.Parallel()

			uut := &Deper{
				Client: tc.mockClient,
			}

			deps, warns, err := uut.Dependencies([]byte(tc.input))

			if tc.errorExpected {
				if err == nil {
					t.Logf("expected an error but got nil")
				}
				return
			}

			if err != nil {
				t.Logf("unexpected error, %v", err)
				t.Fail()
			}

			if !reflect.DeepEqual(deps, tc.expectedDeps) {
				t.Logf("expected deps, %#v to equal %#v",deps, tc.expectedDeps)
				t.Fail()
			}

			if !reflect.DeepEqual(warns, tc.expectedWarns) {
				t.Logf("expected warns, %#v to equal %#v",warns, tc.expectedWarns)
				t.Fail()
			}


		})
	}

}