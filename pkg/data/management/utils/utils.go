// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	cli "github.com/rancher/machine/libmachine/mcnflag"
	"github.com/rancher/norman/types/convert"
	v32 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
)

func ToLowerCamelCase(nodeFlagName string) (string, error) {
	parts := strings.SplitN(nodeFlagName, "-", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("parameter %s does not follow expected naming convention [DRIVER]-[FLAG-NAME]", nodeFlagName)
	}
	flagNameParts := strings.Split(parts[1], "-")
	flagName := flagNameParts[0]
	for _, flagNamePart := range flagNameParts[1:] {
		flagName = flagName + strings.ToUpper(flagNamePart[:1]) + flagNamePart[1:]
	}
	return flagName, nil
}

func FlagToField(flag cli.Flag) (string, v32.Field, error) {
	field := v32.Field{
		Create: true,
		Update: true,
		Type:   "string",
	}

	name, err := ToLowerCamelCase(flag.String())
	if err != nil {
		return name, field, err
	}

	switch v := flag.(type) {
	case *cli.StringFlag:
		field.Description = v.Usage
		field.Default.StringValue = v.Value
	case *cli.IntFlag:
		// This will make the int flag appear as a string field in the rancher API, but we are doing this to maintain
		// backward compatibility, at least until we fix a bug that prevents nodeDriver schemas from updating upon
		// a Rancher upgrade
		field.Description = v.Usage
		field.Default.StringValue = strconv.Itoa(v.Value)
	case *cli.BoolFlag:
		field.Type = "boolean"
		field.Description = v.Usage
	case *cli.StringSliceFlag:
		field.Type = "array[string]"
		field.Description = v.Usage
		field.Nullable = true
		field.Default.StringSliceValue = v.Value
	case *BoolPointerFlag:
		field.Type = "boolean"
		field.Description = v.Usage
	default:
		return name, field, fmt.Errorf("unknown type of flag %v: %v", flag, reflect.TypeOf(flag))
	}

	return name, field, nil
}

func UpdateDefault(credField v32.Field, val, kind string) v32.Field {
	switch kind {
	case "int":
		i, err := strconv.Atoi(val)
		if err == nil {
			credField.Default = v32.Values{IntValue: i}
		} else {
			logrus.Errorf("error converting %s to int %v", val, err)
		}
	case "boolean":
		credField.Default = v32.Values{BoolValue: convert.ToBool(val)}
	case "array[string]":
		credField.Default = v32.Values{StringSliceValue: convert.ToStringSlice(val)}
	case "password", "string":
		credField.Default = v32.Values{StringValue: val}
	default:
		logrus.Errorf("unsupported kind for default val:%s kind:%s", val, kind)
	}
	return credField
}

func CredentialConfigSchemaName(driverName string) string {
	return fmt.Sprintf("%s%s", driverName, "credentialconfig")
}

func GetCredFields(annotations map[string]string) (map[string]bool, map[string]bool, map[string]bool, map[string]string, map[string]bool) {
	getMap := func(fields string) map[string]bool {
		data := map[string]bool{}
		for _, field := range strings.Split(fields, ",") {
			data[field] = true
		}
		return data
	}
	getDefaults := func(fields string) map[string]string {
		data := map[string]string{}
		for _, pattern := range strings.Split(fields, ",") {
			split := strings.SplitN(pattern, ":", 2)
			if len(split) == 2 {
				data[split[0]] = split[1]
			}
		}
		return data
	}
	return getMap(annotations["publicCredentialFields"]),
		getMap(annotations["privateCredentialFields"]),
		getMap(annotations["passwordFields"]),
		getDefaults(annotations["defaults"]),
		getMap(annotations["optionalCredentialFields"])
}
