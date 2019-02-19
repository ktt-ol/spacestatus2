package state

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func Test_ParseOpenValue(t *testing.T) {
	isParseError(t, "moin")
	isParseError(t, "openx")

	// normal values
	isOpenValue(t, "", NONE)
	isOpenValue(t, "none", NONE)
	isOpenValue(t, "open", OPEN)

	// test legacy values
	isOpenValue(t, "closed", NONE)
	isOpenValue(t, "opened", OPEN)

	// special state is not parsable
	isParseError(t, "closing")
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
