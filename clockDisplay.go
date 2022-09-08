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
	"fmt"
	"time"

	"benjamin.barenblat.name/audiotrond/cfa635"
)

const (
	lowerHalfSprite = iota
	upperHalfSprite
	lowerHalfEdgeSprite
	upperHalfEdgeSprite
	fullBlockSprite
	fullBlockEdgeSprite
	lowerRightSprite
	rightLowerEdgeSprite
)

func initializeClockDisplay(lcd *cfa635.Module) error {
	// Load sprites.

	if err := lcd.SetCharacter(lowerHalfSprite, &[...]byte{
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_111111,
		0b00_111111,
		0b00_111111,
	}); err != nil {
		return err
	}

	if err := lcd.SetCharacter(upperHalfSprite, &[...]byte{
		0b00_111111,
		0b00_111111,
		0b00_111111,
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_000000,
	}); err != nil {
		return err
	}

	if err := lcd.SetCharacter(lowerHalfEdgeSprite, &[...]byte{
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_001111,
		0b00_001111,
		0b00_001111,
	}); err != nil {
		return err
	}

	if err := lcd.SetCharacter(upperHalfEdgeSprite, &[...]byte{
		0b00_001111,
		0b00_001111,
		0b00_001111,
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_000000,
		0b00_000000,
	}); err != nil {
		return err
	}

	if err := lcd.SetCharacter(fullBlockSprite, &[...]byte{
		0b00_111111,
		0b00_111111,
		0b00_111111,
		0b00_111111,
		0b00_111111,
		0b00_111111,
		0b00_111111,
		0b00_111111,
	}); err != nil {
		return err
	}

	if err := lcd.SetCharacter(fullBlockEdgeSprite, &[...]byte{
		0b00_001111,
		0b00_001111,
		0b00_001111,
		0b00_001111,
		0b00_001111,
		0b00_001111,
		0b00_001111,
		0b00_001111,
	}); err != nil {
		return err
	}

	if err := lcd.SetCharacter(lowerRightSprite, &[...]byte{
		0b00_000001,
		0b00_000011,
		0b00_000111,
		0b00_001111,
		0b00_011111,
		0b00_111111,
		0b00_111111,
		0b00_111111,
	}); err != nil {
		return err
	}

	if err := lcd.SetCharacter(rightLowerEdgeSprite, &[...]byte{
		0b00_000001,
		0b00_000011,
		0b00_000111,
		0b00_001111,
		0b00_001111,
		0b00_001111,
		0b00_001111,
		0b00_001111,
	}); err != nil {
		return err
	}

	return nil
}

func blitClockDigit(n int, lcd *cfa635.LCDState, x int) {
	switch n {
	case 0:
		lcd[0][x] = rightLowerEdgeSprite
		lcd[0][x+1] = fullBlockSprite
		lcd[0][x+2] = fullBlockSprite
		lcd[1][x] = fullBlockEdgeSprite
		lcd[1][x+2] = fullBlockSprite
		lcd[2][x] = fullBlockEdgeSprite
		lcd[2][x+2] = fullBlockSprite
		lcd[3][x] = fullBlockEdgeSprite
		lcd[3][x+1] = fullBlockSprite
		lcd[3][x+2] = fullBlockSprite

	case 1:
		lcd[0][x+2] = lowerRightSprite
		lcd[1][x+2] = fullBlockSprite
		lcd[2][x+2] = fullBlockSprite
		lcd[3][x+2] = fullBlockSprite

	case 2:
		lcd[0][x] = rightLowerEdgeSprite
		lcd[0][x+1] = fullBlockSprite
		lcd[0][x+2] = fullBlockSprite
		lcd[1][x] = lowerHalfEdgeSprite
		lcd[1][x+1] = lowerHalfSprite
		lcd[1][x+2] = fullBlockSprite
		lcd[2][x] = fullBlockEdgeSprite
		lcd[2][x+1] = upperHalfSprite
		lcd[2][x+2] = upperHalfSprite
		lcd[3][x] = fullBlockEdgeSprite
		lcd[3][x+1] = fullBlockSprite
		lcd[3][x+2] = fullBlockSprite

	case 3:
		lcd[0][x] = rightLowerEdgeSprite
		lcd[0][x+1] = fullBlockSprite
		lcd[0][x+2] = fullBlockSprite
		lcd[1][x] = lowerHalfEdgeSprite
		lcd[1][x+1] = lowerHalfSprite
		lcd[1][x+2] = fullBlockSprite
		lcd[2][x] = upperHalfEdgeSprite
		lcd[2][x+1] = upperHalfSprite
		lcd[2][x+2] = fullBlockSprite
		lcd[3][x] = fullBlockEdgeSprite
		lcd[3][x+1] = fullBlockSprite
		lcd[3][x+2] = fullBlockSprite

	case 4:
		lcd[0][x] = rightLowerEdgeSprite
		lcd[0][x+2] = fullBlockSprite
		lcd[1][x] = fullBlockEdgeSprite
		lcd[1][x+1] = lowerHalfSprite
		lcd[1][x+2] = fullBlockSprite
		lcd[2][x] = upperHalfEdgeSprite
		lcd[2][x+1] = upperHalfSprite
		lcd[2][x+2] = fullBlockSprite
		lcd[3][x+2] = fullBlockSprite

	case 5:
		lcd[0][x] = rightLowerEdgeSprite
		lcd[0][x+1] = fullBlockSprite
		lcd[0][x+2] = fullBlockSprite
		lcd[1][x] = fullBlockEdgeSprite
		lcd[1][x+1] = lowerHalfSprite
		lcd[1][x+2] = lowerHalfSprite
		lcd[2][x] = upperHalfEdgeSprite
		lcd[2][x+1] = upperHalfSprite
		lcd[2][x+2] = fullBlockSprite
		lcd[3][x] = fullBlockEdgeSprite
		lcd[3][x+1] = fullBlockSprite
		lcd[3][x+2] = fullBlockSprite

	case 6:
		lcd[0][x] = rightLowerEdgeSprite
		lcd[0][x+1] = fullBlockSprite
		lcd[0][x+2] = fullBlockSprite
		lcd[1][x] = fullBlockEdgeSprite
		lcd[1][x+1] = lowerHalfSprite
		lcd[1][x+2] = lowerHalfSprite
		lcd[2][x] = fullBlockEdgeSprite
		lcd[2][x+1] = upperHalfSprite
		lcd[2][x+2] = fullBlockSprite
		lcd[3][x] = fullBlockEdgeSprite
		lcd[3][x+1] = fullBlockSprite
		lcd[3][x+2] = fullBlockSprite

	case 7:
		lcd[0][x] = rightLowerEdgeSprite
		lcd[0][x+1] = fullBlockSprite
		lcd[0][x+2] = fullBlockSprite
		lcd[1][x+2] = fullBlockSprite
		lcd[2][x+2] = fullBlockSprite
		lcd[3][x+2] = fullBlockSprite

	case 8:
		lcd[0][x] = rightLowerEdgeSprite
		lcd[0][x+1] = fullBlockSprite
		lcd[0][x+2] = fullBlockSprite
		lcd[1][x] = fullBlockEdgeSprite
		lcd[1][x+1] = lowerHalfSprite
		lcd[1][x+2] = fullBlockSprite
		lcd[2][x] = fullBlockEdgeSprite
		lcd[2][x+1] = upperHalfSprite
		lcd[2][x+2] = fullBlockSprite
		lcd[3][x] = fullBlockEdgeSprite
		lcd[3][x+1] = fullBlockSprite
		lcd[3][x+2] = fullBlockSprite

	case 9:
		lcd[0][x] = rightLowerEdgeSprite
		lcd[0][x+1] = fullBlockSprite
		lcd[0][x+2] = fullBlockSprite
		lcd[1][x] = fullBlockEdgeSprite
		lcd[1][x+1] = lowerHalfSprite
		lcd[1][x+2] = fullBlockSprite
		lcd[2][x] = upperHalfEdgeSprite
		lcd[2][x+1] = upperHalfSprite
		lcd[2][x+2] = fullBlockSprite
		lcd[3][x+2] = fullBlockSprite

	default:
		panic(fmt.Sprint("unknown digit ", n))
	}
}

func clockView(now time.Time) *view {
	var new view
	new.LCD = cfa635.ClearedLCDState()

	now = now.Local()
	h := now.Hour() % 12
	if h == 0 {
		h = 12
	}
	if h >= 10 {
		blitClockDigit(1, new.LCD, -2)
		h -= 10
	}
	blitClockDigit(h, new.LCD, 1)

	new.LCD[1][4] = 0xbb
	new.LCD[2][4] = 0xbb

	m := now.Minute()
	blitClockDigit(m/10, new.LCD, 5)
	blitClockDigit(m%10, new.LCD, 8)

	new.LCD[1][11] = 0xbb
	new.LCD[2][11] = 0xbb

	s := now.Second()
	blitClockDigit(s/10, new.LCD, 12)
	blitClockDigit(s%10, new.LCD, 15)

	if now.Hour() < 12 {
		new.LCD[2][19] = 'a'
	} else {
		new.LCD[2][19] = 'p'
	}
	new.LCD[3][19] = 'm'

	return &new
}
