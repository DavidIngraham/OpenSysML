// Copyright 2025 Open‐MBEE Foundation. All rights reserved.
// Use of this source code is governed by the LICENSE file.

package jsonrpc

const (
	largeResponseThreshold = 16 << 20
	largeResponseQuantum   = 4 << 20
)

// ResponseCapacity groups large JSON buffers into bounded allocation classes.
// Small buffers retain their requested size; large buffers add less than 25% slack.
func ResponseCapacity(size int) int {
	if size < largeResponseThreshold {
		return size
	}
	remainder := size % largeResponseQuantum
	if remainder == 0 {
		return size
	}
	padding := largeResponseQuantum - remainder
	maxInt := int(^uint(0) >> 1)
	if size > maxInt-padding {
		return size
	}
	return size + padding
}
