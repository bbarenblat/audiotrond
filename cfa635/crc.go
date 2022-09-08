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

	"github.com/sigurn/crc16"
)

var crcTable = crc16.MakeTable(crc16.CRC16_X_25)

func pushCRC(b []byte) []byte {
	end := len(b)
	b = append(b, 0, 0)
	binary.LittleEndian.PutUint16(b[end:len(b)], crc16.Checksum(b[:end], crcTable))
	return b
}

func popCRC(b []byte) ([]byte, bool) {
	if len(b) < 2 {
		return b, false
	}

	end := len(b) - 2
	return b[:end], crc16.Checksum(b[:end], crcTable) == binary.LittleEndian.Uint16(b[end:len(b)])
}
