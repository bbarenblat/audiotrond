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

// Package cfa635 interfaces with Crystalfontz CFA635-xxx-KU LCD modules.
//
// The Connect function adopts an existing serial port:
//
// 	s, err := serial.OpenPort(&serial.Config{Name: "/dev/ttyUSB0", Baud: 115200})
// 	if err != nil {
// 		panic(err)
// 	}
// 	m := cfa635.Connect(s)
// 	defer m.Close()
//
// Having connected, you can then issue commands to the module:
//
// 	msg, _, err := transform.String(cfa635.NewEncoder(), "Hello, world!")
// 	if err != nil {
// 		panic(err)
// 	}
// 	if err := m.Put(0, 0, []byte(msg)); err != nil {
// 		panic(err)
// 	}
//
// You can also use the State type and Update function to send the optimal
// sequence of commands to the module to transform its state:
//
// 	if err := m.Reset(); err != nil {
// 		panic(err)
// 	}
// 	s := new(cfa635.State)
// 	t := new(cfa635.State)
// 	copy(t.LCD[0][0:19], msg)
// 	if err := cfa635.Update(m, s, t); err != nil {
// 		panic(err)
// 	}
// 	s = t
package cfa635

import (
	"bufio"
	"errors"
	"io"
	"reflect"
	"sync"
	"time"
)

var (
	ErrPayloadTooLarge = errors.New("payload too large")
	ErrCGRAM           = errors.New("CGRAM index out of range")
	ErrSprite          = errors.New("invalid sprite")
	ErrPosition        = errors.New("position out of range")
	ErrBacklight       = errors.New("backlight brightness out of range")
	ErrLEDIndex        = errors.New("LED index out of range")
	ErrLEDDuty         = errors.New("LED duty cycle out of range")

	ErrTimeout = errors.New("timed out")

	ErrFailed = errors.New("command failed")
)

// Module is a handle to a CFA635 module.
type Module struct {
	w         io.WriteCloser
	reports   chan any    // Messages initiated by the CFA635
	responses chan []byte // Responses to messages initiated by the host

	// Ensures that only one request is in flight to the CFA635 at once
	mu sync.Mutex
}

// Connect constructs a Module from a serial connection to a CFA635.
func Connect(cfa635 io.ReadWriteCloser) *Module {
	m := Module{w: cfa635, reports: make(chan any), responses: make(chan []byte)}

	bytes := make(chan byte)
	go buffer(bufio.NewReader(cfa635), bytes)
	packets := make(chan []byte)
	go decode(bytes, packets)
	go route(packets, m.reports, m.responses)

	return &m
}

// Close releases the CFA635 and closes the underlying connection.
func (m *Module) Close() { m.w.Close() }

// ReadReport blocks until the CFA635 sends a report to the host and then
// returns it. The returned report will be a *KeyActivity, *FanSpeed, or
// *Temperature.
func (m *Module) ReadReport() any { return <-m.reports }

// RawCommand sends an arbitrary command (request and payload) to the CFA635,
// returning the response (and payload) or error.
func (m *Module) RawCommand(req byte, reqP []byte) (resp byte, respP []byte, err error) {
	p := []byte{req, byte(len(reqP))}
	p = append(p, reqP...)
	p = pushCRC(p)

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, err := m.w.Write(p); err != nil {
		return 0, nil, err
	}

	var q []byte
	select {
	case q = <-m.responses:
		break
	case <-time.After(timeout):
		return 0, nil, ErrTimeout
	}

	if len(q) < 2 || len(q) != 2+int(q[1]) {
		return 0, nil, ErrFailed
	}

	return q[0], q[2:], nil
}

func (m *Module) simple(req byte, reqP []byte, wantC byte, wantP []byte) error {
	gotC, gotP, err := m.RawCommand(req, reqP)
	if err != nil {
		return err
	}

	// DeepEqual treats nil and empty slices as non-equal, so check for
	// emptiness explicitly.
	if gotC != wantC || (len(wantP) != 0 || len(gotP) != 0) && !reflect.DeepEqual(gotP, wantP) {
		return ErrFailed
	}
	return nil
}

// Ping pings the CFA635 with a payload of up to 16 bytes.
func (m *Module) Ping(payload []byte) error {
	if len(payload) > 16 {
		return ErrPayloadTooLarge
	}
	return m.simple(0x00, payload, 0x40, payload)
}

// Clear clears the CFA635 LCD. After Clear returns successfully, all LCD cells
// hold 0x20 (space).
func (m *Module) Clear() error { return m.simple(0x06, nil, 0x46, nil) }

// SetCharacter sets a sprite (six columns by eight rows) in character generator
// RAM. The index of the sprite must be between 0 and 7, inclusive.
//
// The sprite itself is specified as an 8-element byte array, each byte
// representing a row. In each byte, the upper two bits must be 0; the lower six
// determine the six pixels in the row, with 1 bits corresponding to active
// pixels and 0 bits corresponding to inactive ones.
func (m *Module) SetCharacter(i int, data *[8]byte) error {
	if i < 0 || i > 7 {
		return ErrCGRAM
	}
	for _, b := range data {
		if b&0b11_000000 != 0 {
			return ErrSprite
		}
	}

	payload := []byte{byte(i)}
	payload = append(payload, data[:]...)
	return m.simple(0x09, payload, 0x49, nil)
}

// SetBacklight controls the LEDs backing the LCD and keypad. Each LED value can
// range from 0 to 100, inclusive, with 0 turning off the light and 100 turning
// it on to its maximum brightness.
func (m *Module) SetBacklight(lcd, keypad int) error {
	if lcd < 0 || lcd > 100 || keypad < 0 || keypad > 100 {
		return ErrBacklight
	}

	return m.simple(0x0e, []byte{byte(lcd), byte(keypad)}, 0x4e, nil)
}

// Put writes data to the LCD at a row and column. No wrapping occurs; if the
// data are too large, they are truncated. Data are interpreted in the CFA635
// character set; see NewEncoder.
func (m *Module) Put(col, row int, data []byte) error {
	if col < 0 || col >= 20 || row < 0 || row >= 4 {
		return ErrPosition
	}
	width := 20 - col
	if len(data) > width {
		data = data[:width]
	}

	payload := []byte{byte(col), byte(row)}
	payload = append(payload, data...)
	return m.simple(0x1f, payload, 0x5f, nil)
}

// SetLED controls the four red/green LEDs to the left of the LCD. The LEDs are
// numbered 0 through 3 from top to bottom; for each, the red and green
// components can be set separately to a value from 0 (off) to 100 (full duty
// cycle).
func (m *Module) SetLED(led int, green bool, duty int) error {
	if led < 0 || led > 3 {
		return ErrLEDIndex
	}
	if duty < 0 || duty > 100 {
		return ErrLEDDuty
	}

	i := byte(11 - 2*led)
	if !green {
		i++
	}
	return m.simple(0x22, []byte{i, byte(duty)}, 0x62, nil)
}
