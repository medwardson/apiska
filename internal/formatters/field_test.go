// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package formatters

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
	"github.com/stretchr/testify/assert"
)

func TestFieldToString(t *testing.T) {
	t.Run("formats null field", func(t *testing.T) {
		field := &types.FieldMemberIsNull{Value: true}

		assert.Equal(t, "NULL", FieldToString(field))
	})

	t.Run("formats boolean true", func(t *testing.T) {
		field := &types.FieldMemberBooleanValue{Value: true}

		assert.Equal(t, "true", FieldToString(field))
	})

	t.Run("formats boolean false", func(t *testing.T) {
		field := &types.FieldMemberBooleanValue{Value: false}

		assert.Equal(t, "false", FieldToString(field))
	})

	t.Run("formats long value", func(t *testing.T) {
		field := &types.FieldMemberLongValue{Value: 42}

		assert.Equal(t, "42", FieldToString(field))
	})

	t.Run("formats negative long value", func(t *testing.T) {
		field := &types.FieldMemberLongValue{Value: -12345}

		assert.Equal(t, "-12345", FieldToString(field))
	})

	t.Run("formats double value", func(t *testing.T) {
		field := &types.FieldMemberDoubleValue{Value: 3.14}

		assert.Equal(t, "3.14", FieldToString(field))
	})

	t.Run("formats double with scientific notation for large values", func(t *testing.T) {
		field := &types.FieldMemberDoubleValue{Value: 1.23e10}

		assert.Equal(t, "1.23e+10", FieldToString(field))
	})

	t.Run("formats string value", func(t *testing.T) {
		field := &types.FieldMemberStringValue{Value: "hello"}

		assert.Equal(t, "hello", FieldToString(field))
	})

	t.Run("formats empty string value", func(t *testing.T) {
		field := &types.FieldMemberStringValue{Value: ""}

		assert.Equal(t, "", FieldToString(field))
	})

	t.Run("formats blob value", func(t *testing.T) {
		field := &types.FieldMemberBlobValue{Value: []byte{0x01, 0x02, 0x03}}

		assert.Equal(t, "<blob:3 bytes>", FieldToString(field))
	})

	t.Run("formats empty blob value", func(t *testing.T) {
		field := &types.FieldMemberBlobValue{Value: []byte{}}

		assert.Equal(t, "<blob:0 bytes>", FieldToString(field))
	})

	t.Run("formats boolean array", func(t *testing.T) {
		field := &types.FieldMemberArrayValue{
			Value: &types.ArrayValueMemberBooleanValues{Value: []bool{true, false, true}},
		}

		assert.Equal(t, "{true,false,true}", FieldToString(field))
	})

	t.Run("formats long array", func(t *testing.T) {
		field := &types.FieldMemberArrayValue{
			Value: &types.ArrayValueMemberLongValues{Value: []int64{1, 2, 3}},
		}

		assert.Equal(t, "{1,2,3}", FieldToString(field))
	})

	t.Run("formats double array", func(t *testing.T) {
		field := &types.FieldMemberArrayValue{
			Value: &types.ArrayValueMemberDoubleValues{Value: []float64{1.1, 2.2, 3.3}},
		}

		assert.Equal(t, "{1.1,2.2,3.3}", FieldToString(field))
	})

	t.Run("formats string array", func(t *testing.T) {
		field := &types.FieldMemberArrayValue{
			Value: &types.ArrayValueMemberStringValues{Value: []string{"a", "b", "c"}},
		}

		assert.Equal(t, "{a,b,c}", FieldToString(field))
	})

	t.Run("formats empty array", func(t *testing.T) {
		field := &types.FieldMemberArrayValue{
			Value: &types.ArrayValueMemberLongValues{Value: []int64{}},
		}

		assert.Equal(t, "{}", FieldToString(field))
	})

	t.Run("formats nested array", func(t *testing.T) {
		field := &types.FieldMemberArrayValue{
			Value: &types.ArrayValueMemberArrayValues{
				Value: []types.ArrayValue{
					&types.ArrayValueMemberLongValues{Value: []int64{1, 2}},
					&types.ArrayValueMemberLongValues{Value: []int64{3, 4}},
				},
			},
		}

		assert.Equal(t, "{{1,2},{3,4}}", FieldToString(field))
	})

	t.Run("returns unknown for nil field", func(t *testing.T) {
		assert.Equal(t, "<unknown>", FieldToString(nil))
	})
}
