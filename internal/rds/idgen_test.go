// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package rds

import (
	"regexp"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIDGen_Generate(t *testing.T) {
	t.Run("returns ID matching prefix-NNN format", func(t *testing.T) {
		gen := NewIDGeneratorWithSeed(42)

		id := gen.Generate()

		assert.Regexp(t, `^[a-z-]+-\d{4}$`, id)
	})

	t.Run("increments counter on each call", func(t *testing.T) {
		gen := NewIDGeneratorWithSeed(42)
		counterPattern := regexp.MustCompile(`-(\d+)$`)

		ids := []string{gen.Generate(), gen.Generate(), gen.Generate()}

		assert.Equal(t, "0001", counterPattern.FindStringSubmatch(ids[0])[1])
		assert.Equal(t, "0002", counterPattern.FindStringSubmatch(ids[1])[1])
		assert.Equal(t, "0003", counterPattern.FindStringSubmatch(ids[2])[1])
	})

	t.Run("uses prefixes from the list", func(t *testing.T) {
		gen := NewIDGeneratorWithSeed(42)
		prefixPattern := regexp.MustCompile(`^(.+)-\d+$`)

		for i := range 10000 {
			id := gen.Generate()
			match := prefixPattern.FindStringSubmatch(id)

			require.NotNil(t, match, "iteration %d: ID %q doesn't match expected pattern", i, id)
			assert.Contains(t, idGenPrefixes, match[1], "iteration %d: unexpected prefix", i)
		}
	})

	t.Run("safe for concurrent use", func(t *testing.T) {
		gen := NewIDGenerator()
		ids := make(chan string, 10000)

		var wg sync.WaitGroup
		for range 100 {
			wg.Go(func() {
				for range 100 {
					ids <- gen.Generate()
				}
			})
		}
		wg.Wait()
		close(ids)

		seen := make(map[string]bool)
		for id := range ids {
			assert.False(t, seen[id], "duplicate ID: %s", id)
			seen[id] = true
		}
		assert.Len(t, seen, 10000)
	})
}

func TestIDGen_NewIDGeneratorWithSeed(t *testing.T) {
	t.Run("produces deterministic sequence for same seed", func(t *testing.T) {
		gen1 := NewIDGeneratorWithSeed(12345)
		gen2 := NewIDGeneratorWithSeed(12345)

		for i := range 10 {
			assert.Equal(t, gen1.Generate(), gen2.Generate(), "iteration %d", i)
		}
	})

	t.Run("produces different sequences for different seeds", func(t *testing.T) {
		gen1 := NewIDGeneratorWithSeed(11111)
		gen2 := NewIDGeneratorWithSeed(22222)
		prefixPattern := regexp.MustCompile(`^(.+)-\d+$`)

		var prefixes1, prefixes2 []string
		for range 20 {
			prefixes1 = append(prefixes1, prefixPattern.FindStringSubmatch(gen1.Generate())[1])
			prefixes2 = append(prefixes2, prefixPattern.FindStringSubmatch(gen2.Generate())[1])
		}

		assert.NotEqual(t, prefixes1, prefixes2)
	})
}

func TestIDGen_NewIDGenerator(t *testing.T) {
	t.Run("creates generator", func(t *testing.T) {
		gen := NewIDGenerator()

		ids := []string{gen.Generate(), gen.Generate()}

		assert.Len(t, ids, 2)
		assert.NotEqual(t, ids[0], ids[1])
	})
}
