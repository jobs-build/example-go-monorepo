# `example-go-monorepo` — sibling sources in a Go monorepo

A worked example of **sibling sources + source closure** (JOBS
sibling-sources design 2026-07-26; source-closure design 2026-07-27): a
`BUILD.jobs` in a monorepo subdirectory whose module depends on a **sibling
directory** through a plain Go `replace` directive. `services/api` requires
`example.com/lib/common` via `replace … => ../../lib/common`; the go plugin
([`plugin-go` v0.1.0](https://github.com/jobs-build/plugin-go/releases/tag/v0.1.0))
walks the transitive import closure from `go.mod`, the engine covers exactly
that, and the build sees the sibling at `$SRC_ROOT/lib/common` with the
repo-relative layout intact. Requires
[jobs-iroh ≥ v0.12.0](https://github.com/fables-for-robots/jobs-iroh/releases/tag/v0.12.0).

## What's here

```
example-go-monorepo/
├── README.md
├── go.work                 ← workspace sugar for local editing (the build ignores it: GOWORK=off)
├── docs/
│   └── notes.md            ← deliberately unrelated — the memo demo (see below)
├── lib/
│   └── common/             ← module example.com/lib/common; one function, one greeting
└── services/
    └── api/                ← module example.com/services/api; prints the greeting
        ├── BUILD.jobs      ← the recipe: go plugin in monorepo mode + offline go build
        ├── go.mod          ← require example.com/lib/common v0.0.0 + replace => ../../lib/common
        └── main.go
```

Zero external dependencies — no `go.sum`, no module fetches; the only imports
are the pinned Go toolchain and the pinned plugin build.

## How it works

1. **Context widening.** From a checkout, `--source services/api` defaults the
   ingest root to the **git repo root** (`.git` itself is never ingested), so
   the whole monorepo is the build's context and `dir = services/api`.
2. **`plugins()`** declares the go plugin (pinned tarball of
   [`jobs-build/plugin-go`](https://github.com/jobs-build/plugin-go)).
3. **`build()`** calls it in **closure mode** — `go_mod = source.read("go.mod")`
   and `go_closure = ["."]` alongside `go_sum` — and the response switches to
   `{modules, sources, closure}`: the plugin walks the transitive local
   import graph from the entry package (resolving imports across the
   manifest's relative `replace` targets — Go honors replaces only in the
   main module, so the consumer's `go.mod` anchors the whole local sibling
   closure) and answers `closure = ["//lib/common", "//services/api/…", …]` —
   the reached sibling package dirs, this module's own files, and the
   manifests.
4. The recipe forwards `res["closure"]` into the `closure =` field of the
   `build()` return: a **complete cover** (source-closure design, jobs-iroh
   v0.12.0) — no implicit dir seed. Exactly that closure — nothing else —
   keys the build (KP).
5. The build sandbox materializes the covered tree at `$SRC_ROOT`
   (`/build/src`) with CWD `$SRC = /build/src/services/api`, so
   `../../lib/common` resolves exactly as it does in the checkout and
   `go build` needs no path surgery. The script is the standard offline Go
   build: pinned `go1.26.4` toolchain, `GOPROXY=off`, `CGO_ENABLED=0`,
   `GOWORK=off`.

## Run it

From a checkout of this repo (Linux, `jobs-client` from
[jobs-iroh](https://github.com/fables-for-robots/jobs-iroh)):

```bash
jobs-client build --source services/api    # hermetic offline build → …/api
jobs-client run   --source services/api    # build, then execute → "hello from lib/common"
```

The git-root default widens the context automatically — no flags needed.
(`--source-root` overrides the root; `--no-repo-root` disables the widening.)

## The memo demo (early cutoff at closure granularity)

The build is keyed by **KP** — the content of the plugin-computed closure
(this module's files + `lib/common`), not the whole repo:

```bash
echo "- meeting notes" >> docs/notes.md
jobs-client build --source services/api
#   ✓ build example-go-monorepo api  (cached)   ← outside the closure: memo hit

sed -i 's/hello from lib\/common/hello v2 from lib\/common/' lib/common/common.go
jobs-client run --source services/api
#   rebuilds — the sibling IS covered — and prints the new greeting
```

Edits outside the closure (docs, CI config, `go.work`, other services,
unimported sibling packages) re-run only the cheap eval stages and memo-hit
the build; edits to covered paths rebuild it. Timestamps don't matter either — the covered
tree is normalized before hashing, so `touch` and fresh checkouts of the same
bytes land on the same KP.

## Notes

- The go plugin's monorepo mode also understands `go.work` (`use` directives)
  via a `go_work` kwarg; this example keeps `go.work` as editor sugar only and
  pins the build to the consumer module with `GOWORK=off`.
- With no external requires there is no `go.sum`; the plugin requires the
  kwarg, so the recipe passes `go_sum = ""`. Add a `require` + `go.sum` and
  the module staging loop in `BUILD.jobs` starts fetching, exactly like the
  [`go-build` example](https://github.com/jobs-build/examples/tree/main/go-build).
- A new external `require` in **lib/common** flows through the same plugin
  call unchanged: path-replaced modules have no `go.sum` entries — they are
  pinned by the covered tree instead.
