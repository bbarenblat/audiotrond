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

package main

import "time"

import (
	"math"

	"benjamin.barenblat.name/audiotrond/cfa635"
)

type view struct {
	LCD               *cfa635.LCDState
	DisplayBrightness float64
	Mtime             time.Time
}

func updateView(lcd *cfa635.Module, old, new *view) error {
	if err := cfa635.Update(lcd, old.LCD, new.LCD); err != nil {
		return err
	}

	if new.DisplayBrightness != old.DisplayBrightness {
		bright := new.DisplayBrightness
		if bright < 0 {
			bright = 0
		}
		if err := lcd.SetBacklight(int(math.Round(new.DisplayBrightness)), 0); err != nil {
			return err
		}
	}

	return nil
}
