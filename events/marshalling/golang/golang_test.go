package golang

import (
	"fmt"
	"testing"
)

func TestMarshalString(t *testing.T) {
	/*
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
			require.EqualValues(t, string(expected), actual)
			require.NoError(t, err)
		}
	*/
	_, err := MarshalString(`{ hi: string }`)
	fmt.Println(err)

}
