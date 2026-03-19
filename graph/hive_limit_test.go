package graph

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeBillingPlan(t *testing.T) {
	assert.Equal(t, "free", normalizeBillingPlan(""))
	assert.Equal(t, "free", normalizeBillingPlan("   "))
	assert.Equal(t, "hobbyist", normalizeBillingPlan("  HOBBYIST "))
}

func TestGetHiveLimitForBillingPlan(t *testing.T) {
	assert.Equal(t, 3, getHiveLimitForBillingPlan("free"))
	assert.Equal(t, 200, getHiveLimitForBillingPlan("enterprise"))
	assert.Equal(t, 3, getHiveLimitForBillingPlan("unknown-plan"))
	assert.Equal(t, 3, getHiveLimitForBillingPlan("   "))
}

func TestGetBillingPlanFromContext(t *testing.T) {
	assert.Equal(t, "free", getBillingPlanFromContext(context.Background()))

	withString := context.WithValue(context.Background(), "billingPlan", "  PROFESSIONAL ")
	assert.Equal(t, "professional", getBillingPlanFromContext(withString))

	withInvalidType := context.WithValue(context.Background(), "billingPlan", 123)
	assert.Equal(t, "free", getBillingPlanFromContext(withInvalidType))
}
