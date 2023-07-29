/*
 *  Copyright (c) 2018 Uber Technologies, Inc.
 *
 *     Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package astro

import (
	"fmt"
	"testing"

	"github.com/uber/astro/astro/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testReadResults reads all results channel from an astro operation and
// returns them as a map, indexed by execution ID.
func testReadResults(resultChan <-chan *Result) map[string]*Result {
	ret := map[string]*Result{}
	for result := range resultChan {
		ret[result.ID()] = result
	}
	return ret
}

// testResultErrs returns a map of the results and whether each one
// is an error.
func testResultErrs(results map[string]*Result) map[string]error {
	errors := map[string]error{}
	for id, result := range results {
		errors[id] = result.Err()
	}
	return errors
}

func TestPlanSuccess(t *testing.T) {
	t.Parallel()

	c, err := NewProjectFromConfigFile("fixtures/test-plan-success/astro.yaml")
	require.NoError(t, err)

	c.config.TerraformDefaults.Path = absolutePath("fixtures/mock-terraform/success")

	_, resultChan, err := c.Plan(PlanExecutionParameters{
		ExecutionParameters: ExecutionParameters{
			UserVars: &UserVariables{
				Values: map[string]string{
					"aws_region": "east1",
				},
			},
		},
	})
	require.NoError(t, err)

	// there should be no errors
	assert.Equal(t, map[string]error{
		"app-east1-dev":          nil,
		"app-east1-prod":         nil,
		"app-east1-staging":      nil,
		"database-east1-dev":     nil,
		"database-east1-prod":    nil,
		"database-east1-staging": nil,
		"network-east1-dev":      nil,
		"network-east1-prod":     nil,
		"network-east1-staging":  nil,
		"network-east1-mgmt":     nil,
		"mgmt-east1":             nil,
		"users":                  nil,
	}, testResultErrs(testReadResults(resultChan)))
}

func TestPlanModulesFiltered(t *testing.T) {
	c, err := NewProjectFromConfigFile("fixtures/foosite.yaml")
	require.NoError(t, err)

	c.config.TerraformDefaults.Path = absolutePath("fixtures/mock-terraform/success")

	modulesToPlan := []string{
		"app",
		"database",
	}

	_, resultChan, err := c.Plan(PlanExecutionParameters{
		ExecutionParameters: ExecutionParameters{
			ModuleNames: modulesToPlan,
			UserVars: &UserVariables{
				Values: map[string]string{
					"aws_region": "east1",
				},
			},
		},
	})
	require.NoError(t, err)

	// only the two modules above should be in the plan results
	assert.Equal(t, map[string]error{
		"app-east1-dev":          nil,
		"app-east1-prod":         nil,
		"app-east1-staging":      nil,
		"database-east1-dev":     nil,
		"database-east1-prod":    nil,
		"database-east1-staging": nil,
	}, testResultErrs(testReadResults(resultChan)))
}

func TestPlanVariablesFiltered(t *testing.T) {
	c, err := NewProjectFromConfigFile("fixtures/foosite.yaml")
	require.NoError(t, err)

	c.config.TerraformDefaults.Path = absolutePath("fixtures/mock-terraform/success")

	_, resultChan, err := c.Plan(PlanExecutionParameters{
		ExecutionParameters: ExecutionParameters{
			UserVars: &UserVariables{
				Values: map[string]string{
					"aws_region":  "east1",
					"environment": "dev",
				},
				Filters: map[string]bool{
					"environment": true,
				},
			},
		},
	})
	require.NoError(t, err)

	// only the two modules above should be in the plan results
	assert.Equal(t, map[string]error{
		"app-east1-dev":      nil,
		"database-east1-dev": nil,
		"network-east1-dev":  nil,
	}, testResultErrs(testReadResults(resultChan)))
}

func TestApplySuccess(t *testing.T) {
	t.Parallel()

	c, err := NewProjectFromConfigFile("fixtures/foosite.yaml")
	require.NoError(t, err)

	_, resultChan, err := c.Apply(ApplyExecutionParameters{
		ExecutionParameters: ExecutionParameters{
			UserVars: &UserVariables{
				Values: map[string]string{
					"aws_region": "east1",
				},
			},
		},
	})
	require.NoError(t, err)

	// there should be no errors
	assert.Equal(t, map[string]error{
		"app-east1-dev":          nil,
		"app-east1-prod":         nil,
		"app-east1-staging":      nil,
		"database-east1-dev":     nil,
		"database-east1-prod":    nil,
		"database-east1-staging": nil,
		"network-east1-dev":      nil,
		"network-east1-prod":     nil,
		"network-east1-staging":  nil,
		"network-east1-mgmt":     nil,
		"mgmt-east1":             nil,
		"users":                  nil,
	}, testResultErrs(testReadResults(resultChan)))
}

func TestApplyFailModule(t *testing.T) {
	t.Parallel()

	c, err := NewProjectFromConfigFile("fixtures/test-apply-fail-module/astro.yaml")
	require.NoError(t, err)

	_, resultChan, err := c.Apply(ApplyExecutionParameters{
		ExecutionParameters: ExecutionParameters{
			UserVars: &UserVariables{
				Values: map[string]string{
					"aws_region": "east1",
				},
			},
		},
	})
	require.NoError(t, err)

	results := testResultErrs(testReadResults(resultChan))

	var executionsRan []string
	for id := range results {
		executionsRan = append(executionsRan, id)
	}

	// users module should have failed
	assert.Error(t, results["users"])

	// check that the following modules were skipped
	for _, id := range []string{
		"app-east1-dev",
		"app-east1-prod",
		"app-east1-staging",
		"database-east1-dev",
		"database-east1-prod",
		"database-east1-staging",
	} {
		if utils.StringSliceContains(executionsRan, id) {
			assert.Fail(t, fmt.Sprintf("%s was not skipped", id))
		}
	}

	// check that the following modules had no errors
	for _, id := range []string{
		"network-east1-dev",
		"network-east1-prod",
		"network-east1-staging",
		"network-east1-mgmt",
		"mgmt-east1",
	} {
		assert.NoError(t, results[id])
	}
}

// Tests that variables are passed to the modules that declare them and not
// passed to the modules that didn't
func TestPassVariables(t *testing.T) {
	t.Parallel()

	c, err := NewProjectFromConfigFile("fixtures/test-pass-variables/astro.yaml")
	require.NoError(t, err)

	c.config.TerraformDefaults.Path = absolutePath("fixtures/mock-terraform/success")

	_, resultChan, err := c.Plan(PlanExecutionParameters{
		ExecutionParameters: ExecutionParameters{
			UserVars: &UserVariables{
				Values: map[string]string{
					"region": "east1",
				},
			},
		},
	})
	require.NoError(t, err)

	results := testReadResults(resultChan)

	// Mocked terraform prints parameters it was called with to stderr.
	// Check that variables were passed to the module that declared it and not
	// passed to the one that didn't.
	assert.Contains(t, results["bar-east1"].TerraformResult().Stderr(), "-var region=east1")
	assert.NotContains(t, results["foo"].TerraformResult().Stderr(), "-var")
}
