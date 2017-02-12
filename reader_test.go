package strfile

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeader(t *testing.T) {
	f, err := NewStrFileReader("./strings", "./strings.dat")
	require.NoError(t, err)
	require.NotNil(t, f)
	h, err := f.Header()
	require.NoError(t, err)
	require.EqualValues(t, Header{Version: 1, Numstr: 2, Delim: 0x25, LongLen: 0xf, ShortLen: 0xc}, *h)
}

func TestMessage(t *testing.T) {
	f, err := NewStrFileReader("./strings", "./strings.dat")
	require.NoError(t, err)
	require.NotNil(t, f)
	s, err := f.String(1)
	require.NoError(t, err)
	require.EqualValues(t, "What the heck?\n", s)
}

func TestMessageCount(t *testing.T) {
	f, err := NewStrFileReader("./strings", "./strings.dat")
	require.NoError(t, err)
	require.NotNil(t, f)
	max, err := f.StringCount()
	require.NoError(t, err)
	require.EqualValues(t, 2, max)
}
