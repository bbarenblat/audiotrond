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
	"log"
	"os"
	"os/signal"
	"time"

	"benjamin.barenblat.name/audiotrond/cfa635"
	"github.com/fhs/gompd/v2/mpd"
	"github.com/tarm/serial"
)

func putWrapped(lcd *cfa635.Module, col, row int, data []byte) error {
	for row < 4 && len(data) > 0 {
		if err := lcd.Put(col, row, data); err != nil {
			return err
		}
		if len(data) <= 20-col {
			data = nil
		} else {
			data = data[20-col:]
		}
		col = 0
		row++
	}
	return nil
}

func update[T comparable](dst *T, src T, mtime *time.Time, now time.Time) {
	if *dst == src {
		return
	}
	*dst = src
	*mtime = now
}

type playbackState byte

const (
	_ playbackState = 1 + iota
	stopped
	playing
	paused
)

type foreground byte

const (
	_ foreground = iota
	mpdForeground
	clockForeground
)

type model struct {
	State           playbackState
	LastStateChange time.Time

	Duration time.Duration
	Elapsed  time.Duration

	Track               string
	Artist              string
	Album               string
	LastTrackInfoUpdate time.Time

	Foreground foreground
}

func connectToCFA635() *cfa635.Module {
	s, err := serial.OpenPort(&serial.Config{Name: "/dev/lcd", Baud: 115200})
	if err != nil {
		panic(err)
	}
	m := cfa635.Connect(s)

	go func() {
		for r := m.ReadReport(); r != nil; r = m.ReadReport() {
			log.Println("report:", r)
		}
	}()

	return m
}

func reportPanicOrClear(lcd *cfa635.Module) {
	if v := recover(); v != nil {
		putWrapped(lcd, 0, 0, encode(fmt.Sprint("panic: ", v)))
		panic(v)
	}
	lcd.Clear()
}

func poll(mpd *mpd.Client, now time.Time, model *model) error {
	cmds := mpd.BeginCommandList()
	statusP := cmds.Status()
	currentP := cmds.CurrentSong()
	if err := cmds.End(); err != nil {
		return err
	}

	status, err := statusP.Value()
	if err != nil {
		return err
	}

	switch status["state"] {
	case "stop":
		update(&model.State, stopped, &model.LastStateChange, now)
	case "play":
		update(&model.State, playing, &model.LastStateChange, now)
	case "pause":
		update(&model.State, paused, &model.LastStateChange, now)
	}

	if status["duration"] == "" {
		model.Duration = 0
		model.Elapsed = 0
	} else {
		if model.Duration, err = time.ParseDuration(status["duration"] + "s"); err != nil {
			panic(err)
		}
		if model.Elapsed, err = time.ParseDuration(status["elapsed"] + "s"); err != nil {
			panic(err)
		}
	}

	current, err := currentP.Value()
	if err != nil {
		return err
	}

	update(&model.Track, current["Title"], &model.LastTrackInfoUpdate, now)
	update(&model.Artist, current["Artist"], &model.LastTrackInfoUpdate, now)
	update(&model.Album, current["Album"], &model.LastTrackInfoUpdate, now)
	return nil
}

func ellipsize(src []byte, ellipsis byte, dst []byte) {
	if len(src) <= len(dst) {
		copy(dst, src)
	} else {
		dots := len(dst) - 1
		copy(dst, src[:dots])
		dst[dots] = ellipsis
	}
}

func fmtTime(t, max time.Duration) string {
	f := func(neg string, h, m, s int) string {
		_ = h
		return fmt.Sprintf("%s%d:%02d", neg, m, s)
	}
	if max/time.Hour > 0 {
		f = func(neg string, h, m, s int) string {
			return fmt.Sprintf("%s%d:%02d:%02d", neg, h, m, s)
		}
	} else if int(max/time.Minute)%60 >= 10 {
		f = func(neg string, h, m, s int) string {
			_ = h
			return fmt.Sprintf("%s%02d:%02d", neg, m, s)
		}
	}

	var neg string
	if t < 0 {
		t *= -1
		neg = "-"
	}
	return f(neg, int(t/time.Hour), int(t/time.Minute)%60, int(t/time.Second)%60)
}

func main() {
	lcd := connectToCFA635()
	defer reportPanicOrClear(lcd)
	defer lcd.SetBacklight(0, 0)
	if err := lcd.SetLED(0, false, 0); err != nil {
		panic(err)
	}
	if err := lcd.SetLED(0, true, 0); err != nil {
		panic(err)
	}

	if err := lcd.Clear(); err != nil {
		panic(err)
	}

	mpd, err := mpd.Dial("unix", "/run/mpd/socket")
	if err != nil {
		panic(err)
	}

	var model model

	view1 := new(view)
	view1.LCD = cfa635.ClearedLCDState()
	view1.Mtime = time.Now()

	// Create an idle timer and put it in a drained state so the event loop
	// can set it.
	idle := time.NewTimer(0)
	<-idle.C

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)

EventLoop:
	for {
		now := time.Now()

		if err := poll(mpd, now, &model); err != nil {
			panic(err)
		}

		var foreground2 foreground
		if model.State == playing || now.Sub(model.LastStateChange).Seconds() < 17 {
			foreground2 = mpdForeground
		} else {
			foreground2 = clockForeground
		}

		if foreground2 != model.Foreground {
			switch foreground2 {
			case mpdForeground:
				if err := initializeMPDDisplay(lcd); err != nil {
					panic(err)
				}
			case clockForeground:
				if err := initializeClockDisplay(lcd); err != nil {
					panic(err)
				}
			}
		}
		model.Foreground = foreground2

		var view2 *view
		switch model.Foreground {
		case mpdForeground:
			view2 = mpdView(&model, now, view1)
		case clockForeground:
			view2 = clockView(now)
		}
		if err := updateView(lcd, view1, view2); err != nil {
			panic(err)
		}
		view1 = view2

		idle.Reset(10 * time.Millisecond)
		select {
		case <-sigterm:
			break EventLoop
		case <-idle.C:
		}
	}
}
