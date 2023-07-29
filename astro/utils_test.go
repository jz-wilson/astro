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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCartesian(t *testing.T) {
	t.Parallel()

	params := [][]interface{}{
		{"a", "b", "c"},
		{1, 2, 3},
		{"x", "y"},
	}
	res := cartesian(params...)

	assert.Equal(t, [][]interface{}{
		{"a", 1, "x"},
		{"a", 1, "y"},
		{"a", 2, "x"},
		{"a", 2, "y"},
		{"a", 3, "x"},
		{"a", 3, "y"},
		{"b", 1, "x"},
		{"b", 1, "y"},
		{"b", 2, "x"},
		{"b", 2, "y"},
		{"b", 3, "x"},
		{"b", 3, "y"},
		{"c", 1, "x"},
		{"c", 1, "y"},
		{"c", 2, "x"},
		{"c", 2, "y"},
		{"c", 3, "x"},
		{"c", 3, "y"},
	}, res)
}
