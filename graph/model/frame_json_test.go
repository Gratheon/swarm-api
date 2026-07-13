package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrameMarshalJSONUsesGraphQLWireFieldsAndDistinctSideIDs(t *testing.T) {
	leftID := 101
	rightID := 102
	leftSideID := "101"
	rightSideID := "102"

	frame := Frame{
		ID:       1,
		Position: 3,
		Type:     FrameTypeFoundation,
		LeftID:   &leftID,
		RightID:  &rightID,
		LeftSide: &FrameSide{
			ID: &leftSideID,
		},
		RightSide: &FrameSide{
			ID: &rightSideID,
		},
	}

	payload, err := json.Marshal(frame)
	require.NoError(t, err)

	var got map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(payload, &got))

	// The public JSON shape mirrors the GraphQL/web-app wire contract while
	// preserving separate DB-only left_id/right_id fields through db tags.
	require.Contains(t, got, "leftId")
	require.Contains(t, got, "rightId")
	require.Contains(t, got, "leftSide")
	require.Contains(t, got, "rightSide")
	require.NotContains(t, got, "left")
	require.NotContains(t, got, "right")
	require.JSONEq(t, `101`, string(got["leftId"]))
	require.JSONEq(t, `102`, string(got["rightId"]))

	var leftSide struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(got["leftSide"], &leftSide))
	require.Equal(t, leftSideID, leftSide.ID)

	var rightSide struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(got["rightSide"], &rightSide))
	require.Equal(t, rightSideID, rightSide.ID)
}
