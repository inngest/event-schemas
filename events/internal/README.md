The functionality within these packages is used to parse our cue
definitions within `/defs`, then generate well-defined Event structs
from those instances.

`go generate` should run before any commit, which means that this
package in itself is only needed when developing and adding types
at build time.  Its API is not necessary for working with events.
