// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package formatters

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
)

// FieldToString spins an RDS Data API field into a string the crowd can read.
// Ready for CSV pressings or TUI stage displays. Handles the whole setlist:
// nulls, booleans, numbers, strings, blobs, and arrays (even nested ones).
func FieldToString(field types.Field) string {
	switch v := field.(type) {
	case *types.FieldMemberIsNull:
		return "NULL"
	case *types.FieldMemberBooleanValue:
		return strconv.FormatBool(v.Value)
	case *types.FieldMemberLongValue:
		return strconv.FormatInt(v.Value, 10)
	case *types.FieldMemberDoubleValue:
		return strconv.FormatFloat(v.Value, 'g', -1, 64)
	case *types.FieldMemberStringValue:
		return v.Value
	case *types.FieldMemberBlobValue:
		return fmt.Sprintf("<blob:%d bytes>", len(v.Value))
	case *types.FieldMemberArrayValue:
		return arrayToString(v.Value)
	default:
		return "<unknown>"
	}
}

func arrayToString(arr types.ArrayValue) string {
	var elements []string

	switch v := arr.(type) {
	case *types.ArrayValueMemberBooleanValues:
		elements = make([]string, len(v.Value))
		for i, b := range v.Value {
			elements[i] = strconv.FormatBool(b)
		}
	case *types.ArrayValueMemberLongValues:
		elements = make([]string, len(v.Value))
		for i, n := range v.Value {
			elements[i] = strconv.FormatInt(n, 10)
		}
	case *types.ArrayValueMemberDoubleValues:
		elements = make([]string, len(v.Value))
		for i, n := range v.Value {
			elements[i] = strconv.FormatFloat(n, 'g', -1, 64)
		}
	case *types.ArrayValueMemberStringValues:
		elements = v.Value
	case *types.ArrayValueMemberArrayValues:
		elements = make([]string, len(v.Value))
		for i, nested := range v.Value {
			elements[i] = arrayToString(nested)
		}
	default:
		return "<array>"
	}

	return "{" + strings.Join(elements, ",") + "}"
}
