package graph

import (
	"context"
	"errors"
	"math/rand"

	"github.com/Gratheon/swarm-api/logger"
)

// RandomHiveName is the resolver for the randomHiveName field.
func (r *queryResolver) RandomHiveName(ctx context.Context, language *string) (*string, error) {
	langCode := "en" // Default to English
	if language != nil {
		langCode = *language
	}

	// Access the map from the resolver
	names, ok := r.Resolver.femaleNamesMap[langCode]
	if !ok || len(names) == 0 {
		// Fallback to English if the requested language is not found or empty
		names, ok = r.Resolver.femaleNamesMap["en"]
		if !ok || len(names) == 0 {
			// Should not happen if en names exist in the json
			logger.ErrorWithContext(ctx, errors.New("fallback to 'en' names failed or 'en' list is empty").Error()) // Log specific error
			defaultName := "Bee"                                                                                    // Provide a fallback name
			return &defaultName, nil
		}
	}

	// Select a random name from the chosen list
	randomIndex := rand.Intn(len(names))
	randomName := names[randomIndex]

	return &randomName, nil
}
