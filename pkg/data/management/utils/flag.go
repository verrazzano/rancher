// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package utils

type BoolPointerFlag struct {
	Name   string
	Usage  string
	EnvVar string
}

func (f BoolPointerFlag) String() string {
	return f.Name
}

func (f BoolPointerFlag) Default() interface{} {
	return nil
}
