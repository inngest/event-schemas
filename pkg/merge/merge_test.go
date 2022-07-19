package merge

import (
	"context"
	"log"
	"os"
	"path"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"github.com/inngest/event-schemas/pkg/cueutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/txtar"
)

func TestMerge(t *testing.T) {
	entries, err := os.ReadDir("./testdata")
	require.NoError(t, err)

	// Allows us to focus on a specfic test name.
	focus := ""

	for _, e := range entries {
		t.Run(e.Name(), func(t *testing.T) {
			if !strings.HasSuffix(e.Name(), ".txtar") {
				return
			}

			// NOTE: Until https://github.com/cue-lang/cue/issues/1654 is fixed, we cannot
			// test the merging of slices due to an internal Cue bug.
			if e.Name() == "slices.txtar" {
				return
			}

			if focus != "" && e.Name() != focus {
				return
			}

			archive, err := txtar.ParseFile(path.Join("./testdata", e.Name()))
			if err != nil {
				log.Fatal(err)
			}

			r := &cue.Runtime{}

			var expected []byte
			var actual cue.Value
			for _, f := range archive.Files {
				if f.Name == "expected" {
					expected = f.Data
					continue
				}

				// Add the thingy to a mergy.
				inst, err := r.Compile(".", f.Data)
				require.NoError(t, err)
				actual, err = Merge(context.Background(), inst.Value(), actual)
				require.NoError(t, err)
			}

			// Parse expected and actual.
			instA, err := r.Compile(".", expected)
			require.NoError(t, err)
			expectedVal := instA.Value()

			// Format the expected syntax nicely
			expectedSyntax, err := cueutil.ASTToSyntax(expectedVal.Syntax())
			require.NoError(t, err)

			// for naming consistency
			actualVal := actual
			syntax, err := cueutil.ASTToSyntax(actualVal.Syntax())
			require.NoError(t, err)

			matches := expectedVal.Subsumes(actualVal)

			require.True(t, matches, "generated types do not match.  got: \n%s\nexpected:\n%s", syntax, expectedSyntax)

			// Ensure actual also subsumes expected so that they're the same.
			matches = actualVal.Subsumes(expectedVal)
			require.True(t, matches, "generated types do not reverse-match.  got: \n%s\nexpected:\n%s", syntax, expectedSyntax)
		})
	}

}
