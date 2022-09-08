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

import (
	"encoding/binary"
	"errors"
)

var (
	errUnknownReportType = errors.New("failed to read report: unknown type")
	errUnknownKey        = errors.New("failed to read key activity report: unknown key")
	errFBSCAB            = errors.New("failed to read fan speed report: FBSCAB module disconnected")
	errDOW               = errors.New("failed to read temperature report: DOW sensor error")
)

// KeyActivity is a report that a key has been pressed or released.
type KeyActivity struct {
	K       Key
	Pressed bool
}

// Key represents a button on the CFA635.
type Key int

const (
	_ Key = iota
	UpButton
	DownButton
	LeftButton
	RightButton
	EnterButton
	ExitButton
)

// FanSpeed is a report on the speed of a system fan.
type FanSpeed struct {
	N          int
	TachCycles int
	TimerTicks int
}

// Temperature is a report on the system temperature.
type Temperature struct {
	N       int // Sensor number.
	Celsius float64
}

// decodeReport converts a packet into either a *KeyActivity, a *FanSpeed, or a
// *Temperature, depending on the packet's type tag.
func decodeReport(p []byte) (any, error) {
	switch p[0] {
	case 0x80:
		return decodeKeyActivity(p[2])
	case 0x81:
		return decodeFanSpeed(p[2:6])
	case 0x82:
		return decodeTemperature(p[2:6])
	default:
		return nil, errUnknownReportType
	}
}

// decodeKeyActivity converts the data byte from a key activity report to a
// KeyActivity.
func decodeKeyActivity(b byte) (*KeyActivity, error) {
	if b == 0 || b > 12 {
		return nil, errUnknownKey
	}

	var a KeyActivity
	if b < 7 {
		a.Pressed = true
	} else {
		b -= 6
	}
	a.K = Key(b)
	return &a, nil
}

// decodeFanSpeed converts the data part of a fan speed report packet to a
// FanSpeed.
func decodeFanSpeed(b []byte) (*FanSpeed, error) {
	if b[1] == 0 && b[2] == 0 && b[3] == 0 {
		return nil, errFBSCAB
	}
	return &FanSpeed{int(b[0]), int(b[1]), int(binary.BigEndian.Uint16(b[2:4]))}, nil
}

// decodeTemperature converts the data part of a temperature report packet to a
// Temperature.
func decodeTemperature(b []byte) (*Temperature, error) {
	if b[3] == 0 {
		return nil, errDOW
	}
	return &Temperature{int(b[0]), float64(binary.BigEndian.Uint16(b[1:3])) / 16}, nil
}
