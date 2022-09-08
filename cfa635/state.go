// Copyright 2022 Benjamin Barenblat
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package cfa635

// State represents the state of the CFA635 LCD.
type LCDState [4][20]byte

func ClearedLCDState() *LCDState {
	var r LCDState
	for y := range r {
		for x := range r[y] {
			r[y][x] = 0x20
		}
	}
	return &r
}

func Update(m *Module, old, new *LCDState) error {
	if *old == *new {
		return nil
	}

	for y := range old {
		var first, last int

		for ; first < len(old[y]) && new[y][first] == old[y][first]; first++ {
		}
		if first == len(old[y]) {
			continue
		}

		for last = len(old[y]) - 1; last > first && new[y][last] == old[y][last]; last-- {
		}
		if err := m.Put(first, y, new[y][first:last+1]); err != nil {
			return err
		}
	}

	return nil
}
