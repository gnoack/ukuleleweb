# TODO

## `uku static`: rewrite wiki links at the goldmark AST level

`cmd/uku/static.go` currently rewrites wiki links (e.g. `href="/PageName"`)
using a regex substitution on the rendered HTML string. This is fragile: any
change to how goldmark emits link nodes (extra attributes, different quote
style, etc.) will silently break it.

The correct fix is to perform the rewrite as a goldmark AST transform, before
rendering, where link destination nodes can be accessed and modified directly
with typed access. See `render.go` and `linkext.go` for how existing
transformers are wired into the goldmark pipeline.
