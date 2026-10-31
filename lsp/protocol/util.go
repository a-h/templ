// SPDX-FileCopyrightText: 2019 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package protocol

// NewVersion returns a pointer to i.
//
// Deprecated: use new(i) instead.
//
//go:fix inline
func NewVersion(i int32) *int32 {
	return new(i)
}
