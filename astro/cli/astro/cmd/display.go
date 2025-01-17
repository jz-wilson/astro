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

package cmd

import (
	"fmt"
	"github.com/uber/astro/astro"
	"github.com/uber/astro/astro/terraform"
	"io"

	"github.com/hashicorp/go-multierror"
	"github.com/logrusorgru/aurora"
)

// printExecStatus takes channels for status updates and exec results
// and prints them on screen as they arrive.
func (cli *AstroCLI) printExecStatus(status <-chan string, results <-chan *astro.Result) (errors error) {
	// Print status updates to stdout as they arrive
	if status != nil {
		go func() {
			var out io.Writer

			if cli.flags.verbose {
				out = cli.stdout
			} else {
				out = io.Discard
			}

			for update := range status {
				_, err := fmt.Fprintln(out, update)
				if err != nil {
					return
				}
			}
		}()
	}

	for result := range results {
		var resultType, changesInfo, runtimeInfo string
		var out = cli.stdout

		// If this was an error, append it to the list of errors to
		// return.
		if result.Err() != nil {
			errors = multierror.Append(errors, result.Err())
		}

		terraformResult := result.TerraformResult()

		// Check to see if this result is from a plan
		planResult, _ := terraformResult.(*terraform.PlanResult)

		if result.Err() == nil {
			resultType = aurora.Green("OK").String()
		} else {
			resultType = aurora.Red("ERROR").String()
			out = cli.stderr
		}

		// If this is a plan, show whether it has changes or not
		if planResult != nil {
			if planResult.HasChanges() {
				changesInfo = aurora.Brown(" Changes").String()
			} else {
				changesInfo = aurora.Gray(" No changes").String()
			}
		}

		if terraformResult != nil {
			runtimeInfo = terraformResult.Runtime()
			runtimeInfo = aurora.Sprintf(aurora.Gray(" (%s)"), result.TerraformResult().Runtime())
		}

		// Print status line
		_, err := fmt.Fprintf(out, "%s: %s%s%s\n",
			result.ID(),
			resultType,
			changesInfo,
			runtimeInfo,
		)
		if err != nil {
			return err
		}

		// If this was a plan, print the plan
		if planResult != nil && planResult.HasChanges() {
			planOutput := planResult.Changes()
			if terraform.CanDisplayReadableTerraformPolicyChanges() {
				var err error
				planOutput, err = terraform.ReadableTerraformPolicyChanges(planOutput)
				if err != nil {
					_, err := fmt.Fprintf(out, "\n%s", err)
					if err != nil {
						return err
					}
				}
			}
			_, err := fmt.Fprintf(out, "\n%s", planOutput)
			if err != nil {
				return err
			}
		}

		// If there is a stderr, print it
		if terraformResult != nil {
			_, err := fmt.Fprintf(out, terraformResult.Stderr())
			if err != nil {
				return err
			}
		} else if result.Err() != nil {
			_, err := fmt.Fprintln(out, result.Err())
			if err != nil {
				return err
			}
		}
	}

	return errors
}
