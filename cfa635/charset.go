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
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
)

// NewEncoder returns a Transformer that converts UTF-8 to the CFA635 display
// character set. The Transformer uses ¿ as the replacement character.
//
// The returned Transformer is lossy, converting various Unicode code points to
// the same byte. For example, U+DF LATIN SMALL LETTER SHARP S (ß) and U+03B2
// GREEK SMALL LETTER BETA (β) are both converted to 0xbe.
//
// The returned Transformer will never map anything to bytes in the range 0x00,
// …, 0x0f.
func NewEncoder() transform.Transformer {
	identityMapped := unicode.RangeTable{
		R16: []unicode.Range16{
			{0x20, 0x23, 1},
			{0x25, 0x3f, 1},
			{0x41, 0x5a, 1},
			{0x61, 0x7a, 1}},
		R32:         nil,
		LatinOffset: 83,
	}
	return runes.If(runes.In(&identityMapped), nil, encode{})
}

type encode struct{}

func (_ encode) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nDst < len(dst) && nSrc < len(src) {
		r, rLen := utf8.DecodeRune(src[nSrc:])
		if r == utf8.RuneError {
			err = transform.ErrShortSrc
			break
		}
		dst[nDst] = encode1(r)
		nDst++
		nSrc += rLen
	}
	return
}

func (_ encode) Reset() {}

func encode1(c rune) byte {
	switch c {
	case '⏵', '▶', '▸', '►', '⯈':
		return 0x10
	case '⏴', '◀', '⯇':
		return 0x11
	case '⏫':
		return 0x12
	case '⏬':
		return 0x13
	case '«', '≪', '《':
		return 0x14
	case '»', '≫', '》':
		return 0x15
	case '↖', '⬉', '⭦':
		return 0x16
	case '↗', '⬈', '⭧':
		return 0x17
	case '↙', '⬋', '⭩':
		return 0x18
	case '↘', '⬊', '⭨':
		return 0x19
	case '⏶', '▲', '▴':
		return 0x1a
	case '⏷', '▼', '▾':
		return 0x1b
	case '↲', '↵', '⏎', '⮐':
		return 0x1c
	case '^', '˄', 'ˆ', '⌃':
		return 0x1d
	case 'ᵛ':
		return 0x1e
	case 0xa0, 0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200a, 0x202f, 0x2060, 0x3000:
		return 0x20
	case 0x01c3:
		return 0x21
	case 'ʺ', '˝', '״', '″', '〃':
		return 0x22
	case '℔', '⌗', '♯', '⧣':
		return 0x23
	case '¤':
		return 0x24
	case '٪', '⁒':
		return 0x25
	case 'ʹ', 'ʼ', 'ˈ', '׳', '‘', '’', '′', 'ꞌ':
		return 0x27
	case '٭', '∗', '⚹':
		return 0x2a
	case '˖':
		return 0x2b
	case '‚':
		return 0x2c
	case '˗', '‐', '‑', '‒', '–', '−', '𐆑':
		return 0x2d
	case '․':
		return 0x2e
	case '⁄', '∕', '⟋':
		return 0x2f
	case '։', '׃', '፡', '∶', '꞉':
		return 0x3a
	case ';':
		return 0x3b
	case '˂', '‹', '〈', '⟨', '〈':
		return 0x3c
	case '᐀', '⹀', '゠', '꞊', '𐆐', '🟰':
		return 0x3d
	case '˃', '›', '〉', '⟩', '〉':
		return 0x3e
	case '¡':
		return 0x40
	case 'Ä':
		return 0x5b
	case 'Ö':
		return 0x5c
	case 'Ñ':
		return 0x5d
	case 'Ü':
		return 0x5e
	case '§':
		return 0x5d
	case '¿':
		return 0x60
	case 'ä':
		return 0x7b
	case 'ö':
		return 0x7c
	case 'ñ':
		return 0x7d
	case 'ü':
		return 0x7e
	case 'à':
		return 0x7f
	case '°', '˚', 'ᴼ', 'ᵒ', '⁰':
		return 0x80
	case '¹':
		return 0x81
	case '²':
		return 0x82
	case '³':
		return 0x83
	case '⁴':
		return 0x84
	case '⁵':
		return 0x85
	case '⁶':
		return 0x86
	case '⁷':
		return 0x87
	case '⁸':
		return 0x88
	case '⁹':
		return 0x89
	case '½':
		return 0x8a
	case '¼':
		return 0x8b
	case '±':
		return 0x8c
	case '≥':
		return 0x8d
	case '≤':
		return 0x8e
	case 'µ', 'μ':
		return 0x8f
	case '♪', '𝅘𝅥𝅮':
		return 0x90
	case '♬':
		return 0x91
	case '🔔', '🕭':
		return 0x92
	case '♥', '❤', '💙', '💚', '💛', '💜', '🖤', '🤎', '🧡':
		return 0x93
	case '◆', '♦':
		return 0x94
	case '𐎂':
		return 0x95
	case '「':
		return 0x96
	case '」':
		return 0x97
	case '“', '❝':
		return 0x98
	case '”', '❞':
		return 0x99
	case 'ɑ', 'α':
		return 0x9c
	case 'ɛ', 'ε':
		return 0x9d
	case 'δ':
		return 0x9e
	case '∞':
		return 0x9f
	case '@':
		return 0xa0
	case '£':
		return 0xa1
	case '$':
		return 0xa2
	case '¥':
		return 0xa3
	case 'è':
		return 0xa4
	case 'é':
		return 0xa5
	case 'ù':
		return 0xa6
	case 'ì':
		return 0xa7
	case 'ò':
		return 0xa8
	case 'Ç':
		return 0xa9
	case 'ᵖ':
		return 0xaa
	case 'Ø':
		return 0xab
	case 'ø':
		return 0xac
	case 'ʳ':
		return 0xad
	case 'Å', 'Å':
		return 0xae
	case 'å':
		return 0xaf
	case 'Δ', '∆', '⌂':
		return 0xb0
	case '¢', 'ȼ', '₵':
		return 0xb1
	case 'Φ':
		return 0xb2
	case 'τ':
		return 0xb3
	case 'λ':
		return 0xb4
	case 'Ω', 'Ω':
		return 0xb5
	case 'π':
		return 0xb6
	case 'Ψ':
		return 0xb7
	case 'Ʃ', 'Σ', '∑':
		return 0xb8
	case 'Θ', 'ϴ', 'θ':
		return 0xb9
	case 'Ξ':
		return 0xba
	case '⏺', '⚫', '⬤', '🔴':
		return 0xbb
	case 'Æ':
		return 0xbc
	case 'æ', 'ӕ':
		return 0xbd
	case 'ß', 'β':
		return 0xbe
	case 'É':
		return 0xbf
	case 'Γ':
		return 0xc0
	case 'Λ':
		return 0xc1
	case 'Π', '∏':
		return 0xc2
	case 'Υ', 'ϓ':
		return 0xc3
	case '_', 'ˍ':
		return 0xc4
	case 'È':
		return 0xc5
	case 'Ê':
		return 0xc6
	case 'ê':
		return 0xc7
	case 'ç':
		return 0xc8
	case 'ğ', 'ǧ':
		return 0xc9
	case 'Ş':
		return 0xca
	case 'ş', 'ș':
		return 0xcb
	case 'İ':
		return 0xcc
	case 'ı':
		return 0xcd
	case '~', '˜', '⁓', '∼', '〜', '～':
		return 0xce
	case '◇', '◊', '♢':
		return 0xcf
	case 'ƒ':
		return 0xd5
	case 0x2588:
		return 0xd6
	case 0x2589, 0x258a:
		return 0xd7
	case 0x258b, 0x258c:
		return 0xd8
	case 0x258d:
		return 0xd9
	case 0x258e, 0x258f:
		return 0xda
	case '₧':
		return 0xdb
	case '◦':
		return 0xdc
	case '•', '⋅':
		return 0xdd
	case '↑', '⬆', '⭡':
		return 0xde
	case '→', '⮕', '⭢':
		return 0xdf
	case '↓', '⬇', '⭣':
		return 0xe0
	case '←', '⬅', '⭠':
		return 0xe1
	case 'Á':
		return 0xe2
	case 'Í':
		return 0xe3
	case 'Ó':
		return 0xe4
	case 'Ú':
		return 0xe5
	case 'Ý':
		return 0xe6
	case 'á':
		return 0xe7
	case 'í':
		return 0xe8
	case 'ó':
		return 0xe9
	case 'ú':
		return 0xea
	case 'ý':
		return 0xeb
	case 'Ô':
		return 0xec
	case 'ô':
		return 0xed
	case 'Č':
		return 0xf0
	case 'Ě':
		return 0xf1
	case 'Ř':
		return 0xf2
	case 'Š':
		return 0xf3
	case 'Ž':
		return 0xf4
	case 'č':
		return 0xf5
	case 'ě':
		return 0xf6
	case 'ř':
		return 0xf7
	case 'š':
		return 0xf8
	case 'ž':
		return 0xf9
	case '[':
		return 0xfa
	case '\\':
		return 0xfb
	case ']':
		return 0xfc
	case '{':
		return 0xfd
	case '|':
		return 0xfe
	case '}':
		return 0xff
	}
	return 0x60 // ¿
}
