package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJsonMarshal(t *testing.T) {
	bytes, err := json.Marshal(make(map[string]int))
	require.NoError(t, err)
	require.Equal(t, "{}", string(bytes))
	bytes, err = json.Marshal((map[string]int)(nil))
	require.NoError(t, err)
	require.Equal(t, "null", string(bytes))
}

func TestSetVarOnCompile(t *testing.T) {
	t.Logf(TestVar)
}
