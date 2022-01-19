package fromjson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"github.com/stretchr/testify/require"
)

func TestFromJSON(t *testing.T) {
	// For each file within testdata ensure that the generated type matches
	// the same output within $filename.cue.

	entries, err := os.ReadDir("./testdata")
	require.NoError(t, err)

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		name := e.Name()

		contents, err := ioutil.ReadFile(path.Join("./testdata", name))
		require.NoError(t, err)

		expected, err := ioutil.ReadFile(path.Join("./testdata", strings.ReplaceAll(name, ".json", ".cue")))
		require.NoError(t, err)

		data := map[string]interface{}{}
		err = json.Unmarshal(contents, &data)
		require.NoError(t, err)

		actual, err := FromJSON(data)
		require.NoError(t, err)
		require.True(t, len(actual) > 2)

		// NOTE: Golang maps do not retain any order.  This means that the
		// values are bound to be different on each test run.  We figure out
		// whether the types are the same by using Cue's "Subsumes" function,
		// which basically means "are the types the same?"

		r := &cue.Runtime{}
		instA, err := r.Compile(".", expected)
		require.NoError(t, err)

		instB, err := r.Compile(".", actual)
		require.NoError(t, err)

		matches := instA.Value().Subsumes(instB.Value())

		if !matches {
			fmt.Println(actual)
		}

		require.True(t, matches, "generated types do not match")
	}
}

const input = ``
