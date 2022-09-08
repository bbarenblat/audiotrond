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

import (
	"math"
	"time"

	"golang.org/x/text/transform"

	"benjamin.barenblat.name/audiotrond/cfa635"
)

var (
	encoder = cfa635.NewEncoder()
)

func encode(s string) []byte {
	r, _, err := transform.String(encoder, s)
	if err != nil {
		panic(err)
	}
	return []byte(r)
}

func setPlaybackIcon(model *model, lcdState *cfa635.LCDState) {
	switch model.State {
	case stopped:
		lcdState[0][19] = 0xd0
	case playing:
		lcdState[0][19] = 0x10
	case paused:
		lcdState[0][19] = pauseIconSprite
	}
}

func rotate(s []byte, width int, start, now time.Time) []byte {
	const delay int = 10
	const rotationRate float64 = 1 / float64(333*time.Millisecond)

	if len(s) <= width {
		return s
	}

	ticks := int(float64(now.Sub(start).Nanoseconds()) * rotationRate)

	if len(s) < 2*width {
		// Just scroll back and forth.
		period := len(s) - width + delay
		ticks %= 2 * period

		i := 0
		if ticks <= delay {
		} else if ticks <= period {
			i = ticks - delay
		} else if ticks <= period+delay {
			i = len(s) - width
		} else {
			i = 2*period - ticks
		}
		return s[i : i+width]
	}

	// Actually rotate.
	for i := 0; i < 7; i++ {
		s = append(s, byte(' '))
	}

	period := len(s) + delay
	ticks %= period

	i := 0
	if ticks > delay {
		i = ticks - delay
	}
	return append(s[i:], s[:i]...)[:width]
}

func setTrackInfo(model *model, now time.Time, lcdState *cfa635.LCDState) {
	// Cut off the track one character short so we don't overwrite the
	// playback icon.
	copy(lcdState[0][:], rotate(encode(model.Track), 19, model.LastTrackInfoUpdate, now))
	copy(lcdState[1][:], rotate(encode(model.Artist), 20, model.LastTrackInfoUpdate, now))
	copy(lcdState[2][:], rotate(encode(model.Album), 20, model.LastTrackInfoUpdate, now))
}

func setTimeElapsed(model *model, lcdState *cfa635.LCDState) int {
	elapsed := fmtTime(model.Elapsed, model.Duration)
	copy(lcdState[3][:], elapsed)
	return len(elapsed)
}

func setTimeRemaining(model *model, lcdState *cfa635.LCDState) int {
	remaining := fmtTime(model.Elapsed-model.Duration, model.Duration)
	copy(lcdState[3][20-len(remaining):], remaining)
	return len(remaining)
}

func setProgressBar(model *model, barStart, barEnd int, lcdState *cfa635.LCDState) {
	fraction := float64(model.Elapsed) / float64(model.Duration)

	// Convert the fraction played to the number of columns that should be
	// colored in the bar. Each cell has 6 columns; leave one extra column
	// free at the left of the bar so it doesn't crash into the time
	// elapsed.
	c := int(math.Round(fraction * (6*float64(barEnd-barStart) - 1)))
	if c == 0 {
		return
	}

	// There are precomposed block glyphs of widths 1-5 that leave a blank
	// column at the beginning. Use one of those in the first cell.
	w := c
	if w > 5 {
		w = 5
	}
	lcdState[3][barStart] = byte(0xdb - w)
	c -= w
	if c == 0 {
		return
	}

	// Fill in the full cells in the bar.
	x := barStart + 1
	for ; c >= 6; x, c = x+1, c-6 {
		lcdState[3][x] = progressBarSpriteFull
	}
	if c == 0 {
		return
	}

	// Fill in the last, partial bar.
	lcdState[3][x] = byte(progressBarSprite1 - 1 + c)
}

const (
	ellipsisSprite = iota
	pauseIconSprite
	progressBarSprite1
	progressBarSprite2
	progressBarSprite3
	progressBarSprite4
	progressBarSprite5
	progressBarSpriteFull
)

func initializeMPDDisplay(lcd *cfa635.Module) error {
	// Load sprites.

	if err := lcd.SetCharacter(ellipsisSprite, &[...]byte{
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_010101,
		0b00_000000,
	}); err != nil {
		return err
	}

	if err := lcd.SetCharacter(pauseIconSprite, &[...]byte{
		0b00_000000,
		0b00_011011,
		0b00_011011,
		0b00_011011,
		0b00_011011,
		0b00_011011,
		0b00_000000,
		0b00_000000,
	}); err != nil {
		return err
	}

	for w := 1; w <= 6; w++ {
		r := byte((1<<w - 1) << (6 - w))
		if err := lcd.SetCharacter(progressBarSprite1-1+w, &[...]byte{
			r,
			r,
			r,
			r,
			r,
			r,
			r,
			0b00_000000,
		}); err != nil {
			return err
		}
	}

	return nil
}

func brightnessStep(pm float64, then, now time.Time, old float64) float64 {
	const brightnessRampRate float64 = 20.0 / float64(500*time.Millisecond)

	dt := float64(now.Sub(then).Nanoseconds())
	return old + pm*dt*brightnessRampRate
}

func setBrightness(model *model, now time.Time, old *view) float64 {
	var z time.Time
	if old.Mtime == z {
		return old.DisplayBrightness
	}

	if model.State == playing || now.Sub(model.LastStateChange).Seconds() < 15 {
		if old.DisplayBrightness >= 20 {
			return 20
		}
		return brightnessStep(+1, old.Mtime, now, old.DisplayBrightness)
	} else {
		if old.DisplayBrightness == 0 {
			return 0
		}
		return math.Max(0, brightnessStep(-1, old.Mtime, now, old.DisplayBrightness))
	}
}

func mpdView(model *model, now time.Time, old *view) *view {
	var new view
	new.LCD = cfa635.ClearedLCDState()
	setPlaybackIcon(model, new.LCD)
	setTrackInfo(model, now, new.LCD)
	if model.Duration > 0 {
		barStart := setTimeElapsed(model, new.LCD)
		barEnd := 20 - setTimeRemaining(model, new.LCD)
		setProgressBar(model, barStart, barEnd, new.LCD)
	}

	new.DisplayBrightness = setBrightness(model, now, old)

	new.Mtime = now

	return &new
}
