/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

// GeoCache stores a value
type GeoCache struct {
	Value       string `json:"value"`
	XcoordRange [2]int
	YcoordRange [2]int
}
