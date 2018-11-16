package requirements

import (
	"log"
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	testCases := []struct{
		name string
		toParse string
		expectedResult []Requirement
		expectError bool
	}{
		{
			name: "package name only",
			toParse: "testPackage",
			expectedResult: []Requirement{
				{
					PackageName: "testPackage",
				},
			},
		},
		{
			name: "package name only untrimmed",
			toParse: " testPackage ",
			expectedResult: []Requirement{
				{
					PackageName: "testPackage",
				},
			},
		},
		{
			name: "package name only newline",
			toParse: " testPackage\n",
			expectedResult: []Requirement{
				{
					PackageName: "testPackage",
				},
			},
		},
		{
			name: "package name only prefixed newlines",
			toParse: "\ntestPackage\n",
			expectedResult: []Requirement{
				{
					PackageName: "testPackage",
				},
			},
		},
		{
			name: "multiple packages",
			toParse: "\ntestPackage\nanotherPackage\n",
			expectedResult: []Requirement{
				{
					PackageName: "testPackage",
				},
				{
					PackageName: "anotherPackage",
				},
			},
		},
		{
			name : "versions",
			toParse: "testPackage==1.2.3",
			expectedResult: []Requirement{
				{
					PackageName: "testPackage",
					VersionOperator: "==",
					PackageVersion: "1.2.3",
				},
			},
		},
		{
			name : "versions spaced",
			toParse: " testPackage == 1.2.3 ",
			expectedResult: []Requirement{
				{
					PackageName: "testPackage",
					VersionOperator: "==",
					PackageVersion: "1.2.3",
				},
			},
		},
		{
			name : "versions trailing newline",
			toParse: " testPackage == 1.2.3\n",
			expectedResult: []Requirement{
				{
					PackageName: "testPackage",
					VersionOperator: "==",
					PackageVersion: "1.2.3",
				},
			},
		},
		{
			name : "mix",
			toParse: `
testPackage1
testPackage2==1.2.3
testPackage3~=4.5
testPackage4>=7.8.9a
`,
			expectedResult: []Requirement{
				{
					PackageName: "testPackage1",
				},
				{
					PackageName: "testPackage2",
					VersionOperator: "==",
					PackageVersion: "1.2.3",
				},
				{
					PackageName: "testPackage3",
					VersionOperator: "~=",
					PackageVersion: "4.5",
				},
				{
					PackageName: "testPackage4",
					VersionOperator: ">=",
					PackageVersion: "7.8.9a",
				},
			},
		},
		{
			name: "empty-file",
			toParse: "",
			expectedResult: nil,
		},
		{
			name: "starts-with-invalid-char",
			toParse: "==1.2.3",
			expectError: true,
		},
		{
			name: "missing operator",
			toParse: "myPackage someVersion",
			expectError: true,
		},
		{
			name: "missing version",
			toParse: "myPackage==",
			expectError: true,
		},
		{
			name: "run-on line",
			toParse: "myPackage==1.2.3 anotherPackage",
			expectError: true,
		},
		{
			name: "bad character",
			toParse: "@",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			log.Println(tc.name)
			var got []Requirement
			err := Unmarshal([]byte(tc.toParse),&got)

			if tc.expectError {
				if err == nil {
					t.Logf("expected an error but got nil")
					t.Fail()
				}
				// no further tests - result doesn't matter on an error result
				return
			}

			if err != nil {
				t.Logf("expected no error but got '%s'", err)
				t.Fail()
			}

			if !reflect.DeepEqual(tc.expectedResult, got) {
				t.Logf("unexpected result.  got '%#v' but expected '%#v'", got, tc.expectedResult)
				t.Fail()
			}
		})
	}
}