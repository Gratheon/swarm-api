//go:build !integration
// +build !integration

package graph

import (
	"testing"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/stretchr/testify/assert"
)

func TestValidateFrameCapacity(t *testing.T) {
	assert.NoError(t, validateFrameCapacity(model.BoxTypeLargeHorizontalSection, 0))
	assert.NoError(t, validateFrameCapacity(model.BoxTypeLargeHorizontalSection, 32))
	assert.Error(t, validateFrameCapacity(model.BoxTypeLargeHorizontalSection, 33))
	assert.Error(t, validateFrameCapacity(model.BoxTypeLargeHorizontalSection, 40))
	assert.NoError(t, validateFrameCapacity(model.BoxTypeDeep, 100))
}
