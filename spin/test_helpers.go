package spin

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/gruntwork-io/terragrunt/options"
	"path/filepath"
	"github.com/gruntwork-io/terragrunt/locks/dynamodb"
	"github.com/gruntwork-io/terragrunt/remote"
	"github.com/gruntwork-io/terragrunt/locks"
	"github.com/gruntwork-io/terragrunt/errors"
	"sort"
)

type ByPath []TerraformModule

func (byPath ByPath) Len() int           { return len(byPath) }
func (byPath ByPath) Swap(i, j int)      { byPath[i], byPath[j] = byPath[j], byPath[i] }
func (byPath ByPath) Less(i, j int) bool { return byPath[i].Path < byPath[j].Path }

func assertModuleListsEqual(t *testing.T, expectedModules []TerraformModule, actualModules []TerraformModule, messageAndArgs ...interface{}) {
	if !assert.Equal(t, len(expectedModules), len(actualModules), messageAndArgs...) {
		t.Logf("%s != %s", expectedModules, actualModules)
		return
	}

	sort.Sort(ByPath(expectedModules))
	sort.Sort(ByPath(actualModules))

	for i := 0; i < len(expectedModules); i++ {
		expected := expectedModules[i]
		actual := actualModules[i]
		assertModulesEqual(t, expected, actual, messageAndArgs...)
	}
}

func assertModulesEqual(t *testing.T, expected TerraformModule, actual TerraformModule, messageAndArgs ...interface{}) {
	if assert.NotNil(t, actual, messageAndArgs...) {
		assert.Equal(t, expected.Config, actual.Config, messageAndArgs...)
		assert.Equal(t, expected.Path, actual.Path, messageAndArgs...)

		assertOptionsEqual(t, *expected.TerragruntOptions, *actual.TerragruntOptions, messageAndArgs...)
		assertModuleListsEqual(t, expected.Dependencies, actual.Dependencies, messageAndArgs...)
	}
}

func assertErrorsEqual(t *testing.T, expected error, actual error, messageAndArgs ...interface{}) {
	actual = errors.Unwrap(actual)
	// We can't do a simple IsError comparison for UnrecognizedDependency because that error is a struct that
	// contains an array, and in Go, trying to compare arrays gives a "comparing uncomparable type
	// spin.UnrecognizedDependency" panic. Therefore, we have to compare that error more manually.
	if expectedUnrecognized, isUnrecognizedDependencyError := expected.(UnrecognizedDependency); isUnrecognizedDependencyError {
		actualUnrecognized, isUnrecognizedDependencyError := actual.(UnrecognizedDependency)
		if assert.True(t, isUnrecognizedDependencyError, messageAndArgs...) {
			assert.Equal(t, expectedUnrecognized, actualUnrecognized, messageAndArgs...)
		}
	} else {
		assert.True(t, errors.IsError(actual, expected), messageAndArgs...)
	}
}

// We can't do a direct comparison between TerragruntOptions objects because we can't compare Logger or RunTerragrunt
// instances. Therefore, we have to manually check everything else.
func assertOptionsEqual(t *testing.T, expected options.TerragruntOptions, actual options.TerragruntOptions, messageAndArgs ...interface{}) {
	assert.NotNil(t, expected.Logger, messageAndArgs...)
	assert.NotNil(t, actual.Logger, messageAndArgs...)

	assert.Equal(t, expected.TerragruntConfigPath, actual.TerragruntConfigPath, messageAndArgs...)
	assert.Equal(t, expected.NonInteractive, actual.NonInteractive, messageAndArgs...)
	assert.Equal(t, expected.TerraformCliArgs, actual.TerraformCliArgs, messageAndArgs...)
	assert.Equal(t, expected.WorkingDir, actual.WorkingDir, messageAndArgs...)
}

// Return the absolute path for the given path
func abs(t *testing.T, path string) string {
	out, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Create a new DynamoDB lock
func lock(t *testing.T, stateFileId string) locks.Lock {
	lock, err := dynamodb.New(map[string]string{"state_file_id": stateFileId})
	if err != nil {
		t.Fatal(err)
	}
	return lock
}

// Create a RemoteState struct
func state(t *testing.T, bucket string, key string) *remote.RemoteState {
	return &remote.RemoteState{
		Backend: "s3",
		Config: map[string]string{
			"bucket": bucket,
			"key": key,
		},
	}
}