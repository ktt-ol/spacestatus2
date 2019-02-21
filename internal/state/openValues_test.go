package state

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func Test_ParseOpenValue(t *testing.T) {
	isParseError(t, "moin")
	isParseError(t, "openx")
	isParseError(t, "")

	// normal values
	isOpenValue(t, "none", NONE)
	isOpenValue(t, "open", OPEN)

	// test legacy values
	isOpenValue(t, "closed", NONE)
	isOpenValue(t, "opened", OPEN)

	// special state comes from the db
	isOpenValue(t, "closing", CLOSING)
}

func isParseError(t *testing.T, valueToTest string) {
	_, err := ParseOpenValue(valueToTest)
	require.NotNil(t, err)
}

func isOpenValue(t *testing.T, valueToTest string, expectedResult OpenValue) {
	result, err := ParseOpenValue(valueToTest)
	require.Nil(t, err)
	require.Equal(t, expectedResult, result)
}
