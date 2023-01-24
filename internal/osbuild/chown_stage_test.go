package osbuild

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewChownStage(t *testing.T) {
	expectedStage := &Stage{
		Type:    "org.osbuild.chown",
		Options: &ChownStageOptions{},
	}
	actualStage := NewChownStage(&ChownStageOptions{})
	assert.Equal(t, expectedStage, actualStage)
}
