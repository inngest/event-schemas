package typescript

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypescriptGeneration(t *testing.T) {
	entries, err := os.ReadDir("./testdata")
	require.NoError(t, err)

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".cue") {
			continue
		}

		name := e.Name()
		contents, err := ioutil.ReadFile(path.Join("./testdata", name))
		require.NoError(t, err)

		expected, err := ioutil.ReadFile(path.Join("./testdata", strings.ReplaceAll(name, ".cue", ".ts")))
		require.NoError(t, err)

		actual, err := MarshalString(string(contents))
		require.NoError(t, err)
		require.EqualValues(t, string(expected), actual)
	}

}
