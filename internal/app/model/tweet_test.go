package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTweet_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		tw      *Tweet
		isValid bool
	}{
		{
			name: "valid",
			tw: &Tweet{
				Message: "Some message",
			},
			isValid: true,
		},
		{
			name: "empty msg",
			tw: &Tweet{
				Message: "",
			},
			isValid: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isValid {
				assert.NoError(t, tc.tw.Validate())
			} else {
				assert.Error(t, tc.tw.Validate())
			}
		})
	}
}
