package redisPubSub

import (
	"encoding/json"
	"strings"
	// "fmt"
	"context"
	"log"

	"github.com/Gratheon/swarm-api/graph/model"
	"github.com/jmoiron/sqlx"
	// "strings"
)
type ResourceUpdate struct {
	Delta     [][]float32     `json:"delta"`
}

func ListenFrameResourceUpdates(db *sqlx.DB) {
	// Connect to Redis
	rdb := InitRedis()

	go func() {
		// Create a wait group to ensure proper termination
		ctx := context.Background()
		log.Print("redis subscribing to *.frame_side.*.frame_resources_detected")
		// Subscribe to a Pub/Sub channel
		subscriber := rdb.PSubscribe(ctx, "*.frame_side.*.frame_resources_detected")

		for {
			msg, err := subscriber.ReceiveMessage(ctx)
			if err != nil {
				panic(err)
			}

			log.Print("got event from redis channel", msg.Channel)

			ch := strings.Split(msg.Channel, ".")
			uid := ch[0]
			frameSideId := ch[2]

			// Received string data from Pub/Sub channel
			data := msg.Payload
			log.Print(data)
			// Decode JSON data into a struct
			var event ResourceUpdate
			err = json.Unmarshal([]byte(data), &event)
			if err != nil {
				log.Printf("Error decoding JSON data: %s", err)
				continue
			}

			var counters = make(map[int]int, 6)
			//0 brood capped
			//2 honey
			//3 brood
			//4 nectar
			//5 empty
			//6 pollen

			for _, resources := range event.Delta {
				resourceType := int(resources[0])
				counters[resourceType]++
			}

			frameSideModel := &model.FrameSide{
				Db:     db,
				UserID: uid,
			}

			eggs := 100 * counters[1] / len(event.Delta)
			honey := 100 * counters[2] / len(event.Delta)
			broodCapped := 100 * counters[0] / len(event.Delta)
			brood := 100 * counters[3] / len(event.Delta)
			pollen := 100 * counters[6] / len(event.Delta)
			// nectar := counters[4] / len(detectedResources)

			frameSideModel.UpdateSide(model.FrameSideInput{
				ID:                 frameSideId,
				BroodPercent:       &brood,
				CappedBroodPercent: &broodCapped,
				// NectarPercent: 0,
				EggsPercent:   &eggs,
				PollenPercent: &pollen,
				HoneyPercent:  &honey,
				QueenDetected: false,
			})
		}
	}()
}
