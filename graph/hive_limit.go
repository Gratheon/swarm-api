package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/Gratheon/swarm-api/graph/model"
)

var hiveLimitByBillingPlan = map[string]int{
	"free":         3,
	"hobbyist":     15,
	"starter":      50,
	"professional": 200,
	"addon":        200,
	"enterprise":   200,
}

func normalizeBillingPlan(plan string) string {
	normalized := strings.TrimSpace(strings.ToLower(plan))
	if normalized == "" {
		return "free"
	}
	return normalized
}

func getHiveLimitForBillingPlan(plan string) int {
	normalized := normalizeBillingPlan(plan)
	if limit, ok := hiveLimitByBillingPlan[normalized]; ok {
		return limit
	}
	return hiveLimitByBillingPlan["free"]
}

func getBillingPlanFromContext(ctx context.Context) string {
	if value := ctx.Value("billingPlan"); value != nil {
		if plan, ok := value.(string); ok {
			return normalizeBillingPlan(plan)
		}
	}
	return "free"
}

func enforceHiveCreationLimit(ctx context.Context, hiveModel *model.Hive) error {
	activeHiveCount, err := hiveModel.CountActive()
	if err != nil {
		return err
	}

	billingPlan := getBillingPlanFromContext(ctx)
	hiveLimit := getHiveLimitForBillingPlan(billingPlan)
	if activeHiveCount >= hiveLimit {
		return fmt.Errorf("hive limit reached for %s plan (%d)", billingPlan, hiveLimit)
	}
	return nil
}
