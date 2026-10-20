package graph

import (
	"fmt"

	"github.com/Gratheon/swarm-api/graph/model"
)

// maxFramesPerBoxType caps how many active frames a single box can hold.
// Only horizontal sections are limited for now: long Langstroth hives hold
// 30-33 frames, Lazutin ~25, Layens ~20 (up to 30), so 33 fits every common
// design while still catching runaway additions. Partition and feeder frames
// count too, since they occupy a slot like any other frame.
var maxFramesPerBoxType = map[model.BoxType]int{
	model.BoxTypeLargeHorizontalSection: 33,
}

func validateFrameCapacity(boxType model.BoxType, existingFrameCount int) error {
	limit, ok := maxFramesPerBoxType[boxType]
	if !ok || existingFrameCount < limit {
		return nil
	}
	return fmt.Errorf("box can hold at most %d frames", limit)
}
