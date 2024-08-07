// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpcreplay

import "testing"

func TestRemoveExtraSpaces(t *testing.T) {
	in := `
      fields: {
        key:  "value"
        value:    {
          array_value: {
            values:  	{
              double_value: 0.30135461688041687
            }
            values: {
              double_value: 0.3094993531703949
            }
          }
        }
      }
`
	want := `
      fields: {
        key: "value"
        value: {
          array_value: {
            values: {
              double_value: 0.30135461688041687
            }
            values: {
              double_value: 0.3094993531703949
            }
          }
        }
      }
`
	got := string(removeExtraSpaces([]byte(in)))
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
