/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

// GeoCache stores a value
type GeoCache struct {
	Id          string
	Name        string
	Description string
	XcoordRange [2]int
	YcoordRange [2]int
	Owner       User
	Reports     []Report
	Visitors    []User
	Trackable   Trackable
}

type Trackable struct {
	Id    string
	Value string
}

type User struct {
	Id   string
	Name string
}

type Report struct {
	Id       string
	Message  string
	Notifier User
}
