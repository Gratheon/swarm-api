package logger

import (
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToFields(t *testing.T) {
	fields := toFields("alpha", 2, true)
	assert.Equal(t, "alpha", fields["meta_0"])
	assert.Equal(t, 2, fields["meta_1"])
	assert.Equal(t, true, fields["meta_2"])
}

func TestUserCycleFormatter_Format(t *testing.T) {
	formatter := &UserCycleFormatter{}
	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Time:    time.Date(2026, 3, 19, 12, 34, 56, 0, time.UTC),
		Level:   logrus.InfoLevel,
		Message: "hello",
		Data: logrus.Fields{
			"user_id": "u-1",
		},
	}

	formatted, err := formatter.Format(entry)
	require.NoError(t, err)

	output := string(formatted)
	assert.Contains(t, output, "hello")
	assert.Contains(t, output, "info")
	assert.Contains(t, output, "user_id")
	assert.True(t, strings.HasSuffix(output, "\n"))
}

func TestInitLogging(t *testing.T) {
	logInstance := InitLogging()
	assert.NotNil(t, logInstance)
	assert.Equal(t, logrus.InfoLevel, logInstance.Level)
	_, ok := logInstance.Formatter.(*UserCycleFormatter)
	assert.True(t, ok)
}
