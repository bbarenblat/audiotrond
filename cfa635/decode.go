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
	"io"
	"log"
	"time"
)

const (
	maxPacketBytes = 26

	msgPacketFailed = "failed to read packet from CFA635:"
	msgTimedOut     = "timed out"
)

var (
	timeout = 250 * time.Millisecond // maximum response latency
)

// buffer copies bytes from an io.Reader into a channel, logging any errors as
// it goes. It closes the channel when no more bytes are left or when it
// receives on the done channel.
func buffer(r io.ByteReader, w chan<- byte) {
	defer close(w)
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print("failed to read byte from CFA635: ", err)
			continue
		}
		w <- b
	}
}

// decode reassembles bytes into packets, logging any errors as it goes.
func decode(bytes <-chan byte, packets chan<- []byte) {
	defer close(packets)
Outer:
	for {
		p := make([]byte, 0, maxPacketBytes)

		// Read packet type.
		typ, ok := <-bytes
		if !ok {
			break
		}
		p = append(p, typ)

		// The rest of the packet should come in fairly quickly.
		timedout := time.After(timeout)

		// Read packet length.
		var length byte
		select {
		case <-timedout:
			log.Print(msgPacketFailed, ' ', msgTimedOut)
			continue

		case length, ok = <-bytes:
			if !ok {
				break Outer
			}
			if length > 22 {
				log.Print(msgPacketFailed, " got too-long data_length ", length)
				continue
			}
			p = append(p, length)
		}

		// Read the data and CRC.
		for i := 0; i < int(length)+2; i++ {
			select {
			case <-timedout:
				log.Print(msgPacketFailed, ' ', msgTimedOut)
				continue
			case b, ok := <-bytes:
				if !ok {
					break Outer
				}
				p = append(p, b)
			}
		}

		if p, ok = popCRC(p); !ok {
			log.Printf("%s CRC failure\n", msgPacketFailed)
			continue
		}

		// Save the packet.
		packets <- p
	}
}

// route splits a channel of packets into channels of reports and responses.
func route(packets <-chan []byte, reports chan<- any, responses chan<- []byte) {
	defer close(reports)
	defer close(responses)
	for p := range packets {
		switch p[0] & 0b1100_0000 >> 6 {
		case 0b10:
			r, err := decodeReport(p)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			reports <- r
		default:
			responses <- p
		}
	}
}
