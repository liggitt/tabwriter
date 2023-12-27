// From https://github.com/rivo/uniseg/blob/master/grapheme.go
//
// Licensed under MIT License
//
// Copyright (c) 2019 Oliver Kuederle
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package unicode

import "unicode/utf8"

// The number of bits the grapheme property must be shifted to make place for
// grapheme states.
const shiftGraphemePropState = 4

// The bit mask used to extract the state returned by the [Step] function, after
// shifting. These values must correspond to the shift constants.
const maskGraphemeState = 0xf

// FirstGraphemeCluster returns the first grapheme cluster found in the given
// byte slice according to the rules of [Unicode Standard Annex #29, Grapheme
// Cluster Boundaries]. This function can be called continuously to extract all
// grapheme clusters from a byte slice, as illustrated in the example below.
//
// If you don't know the current state, for example when calling the function
// for the first time, you must pass -1. For consecutive calls, pass the state
// and rest slice returned by the previous call.
//
// The "rest" slice is the sub-slice of the original byte slice "b" starting
// after the last byte of the identified grapheme cluster. If the length of the
// "rest" slice is 0, the entire byte slice "b" has been processed. The
// "cluster" byte slice is the sub-slice of the input slice containing the
// identified grapheme cluster.
//
// The returned width is the width of the grapheme cluster for most monospace
// fonts where a value of 1 represents one character cell.
//
// Given an empty byte slice "b", the function returns nil values.
//
// While slightly less convenient than using the Graphemes class, this function
// has much better performance and makes no allocations. It lends itself well to
// large byte slices.
//
// [Unicode Standard Annex #29, Grapheme Cluster Boundaries]: http://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
func FirstGraphemeCluster(b []byte, state int) (cluster, rest []byte, width, newState int) {
	// An empty byte slice returns nothing.
	if len(b) == 0 {
		return
	}

	// Extract the first rune.
	r, length := utf8.DecodeRune(b)
	if len(b) <= length { // If we're already past the end, there is nothing else to parse.
		var prop int
		if state < 0 {
			prop = property(graphemeCodePoints, r)
		} else {
			prop = state >> shiftGraphemePropState
		}
		return b, nil, runeWidth(r, prop), grAny | (prop << shiftGraphemePropState)
	}

	// If we don't know the state, determine it now.
	var firstProp int
	if state < 0 {
		state, firstProp, _ = transitionGraphemeState(state, r)
	} else {
		firstProp = state >> shiftGraphemePropState
	}
	width += runeWidth(r, firstProp)

	// Transition until we find a boundary.
	for {
		var (
			prop     int
			boundary bool
		)

		r, l := utf8.DecodeRune(b[length:])
		state, prop, boundary = transitionGraphemeState(state&maskGraphemeState, r)

		if boundary {
			return b[:length], b[length:], width, state | (prop << shiftGraphemePropState)
		}

		if r == vs16 {
			width = 2
		} else if firstProp != prExtendedPictographic && firstProp != prRegionalIndicator && firstProp != prL {
			width += runeWidth(r, prop)
		} else if firstProp == prExtendedPictographic {
			if r == vs15 {
				width = 1
			} else {
				width = 2
			}
		}

		length += l
		if len(b) <= length {
			return b, nil, width, grAny | (prop << shiftGraphemePropState)
		}
	}
}
