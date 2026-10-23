# SDK migration comparison at the source baseline

This comparison uses the original Eshu module tags and the SDK source baseline
`7c1eb180b017359e7dee453bcf37056a00851fca`. The baseline is a source
snapshot for the proposed new paths. It is not a candidate tag; the planned
`collector/v0.1.0-rc.1` and `factschema/v0.1.0-rc.1` remain unpublished.

| Original module at `v0.2.0` | Original tag | Origin commit | Module zip checksum | `go.mod` checksum |
| --- | --- | --- | --- | --- |
| `github.com/eshu-hq/eshu/sdk/go/collector` | `sdk/go/collector/v0.2.0` | `2c8ab5c228c87a7b47f16063d4b1f43027d33ad6` | `h1:wOlmnu26Q/MAo50lNGlqj+KJvYDnI6x1CYOThhzNlEg=` | `h1:DtRBLGu3mClyJdlcsuJX9mx0Gb4nvTk2J0U+HvFwNjY=` |
| `github.com/eshu-hq/eshu/sdk/go/factschema` | `sdk/go/factschema/v0.2.0` | `2c8ab5c228c87a7b47f16063d4b1f43027d33ad6` | `h1:47bNIgc95iYH7Thf0YWwIKbcHqf+m1UHC+En6IjOMKU=` | `h1:pVInY0h5obTaV0qlr4f1zkKjDXfIben/rxXZEzbrU5w=` |

The Go checksums verify the downloaded module archives and module files. They
do not prove tree equality between the original and relocated source, nor do
they establish compatibility with an old core consumer.

At this preparation review, the original modules' `v0.2.0` tags and public
proxy resolution are proven by the command below. Eshu's
[core SDK compatibility matrix](https://github.com/eshu-hq/eshu/blob/main/docs/public/extend/sdk-compatibility.md)
still labels its `v0.2.0` row “Pending tag cut.” That row has not been updated
here; a core consumer qualification must reconcile it with the verified tags
and then supply its own compatibility evidence.

## Reproduce the schema comparison

From a clone of this SDK repository containing the baseline commit, run this
command with Go, Git, and Python 3. It downloads the original immutable module
versions from the public Go proxy, reads the new schemas from the pinned Git
commit, and compares parsed JSON. It does not use the current worktree's
schemas, a workspace, or a local replacement.

```sh
python3 - <<'PY'
"""Compare pinned old module schemas with the SDK source baseline."""

import json
import os
import subprocess
from pathlib import Path
from typing import Any, Iterator

base = "7c1eb180b017359e7dee453bcf37056a00851fca"
origin = "2c8ab5c228c87a7b47f16063d4b1f43027d33ad6"
modules = {
    "collector": (
        "h1:wOlmnu26Q/MAo50lNGlqj+KJvYDnI6x1CYOThhzNlEg=",
        "h1:DtRBLGu3mClyJdlcsuJX9mx0Gb4nvTk2J0U+HvFwNjY=",
    ),
    "factschema": (
        "h1:47bNIgc95iYH7Thf0YWwIKbcHqf+m1UHC+En6IjOMKU=",
        "h1:pVInY0h5obTaV0qlr4f1zkKjDXfIben/rxXZEzbrU5w=",
    ),
}
env = dict(os.environ, GOWORK="off", GOPROXY="https://proxy.golang.org")
for name, (zip_sum, mod_sum) in modules.items():
    path = f"github.com/eshu-hq/eshu/sdk/go/{name}"
    output = subprocess.check_output(
        ["go", "mod", "download", "-json", f"{path}@v0.2.0"], env=env
    )
    info = json.loads(output)
    if (info["Sum"], info["GoModSum"]) != (zip_sum, mod_sum):
        raise SystemExit(f"unexpected checksums for {name}")
    if info["Origin"]["Hash"] != origin:
        raise SystemExit(f"unexpected origin commit for {name}")
    if info["Origin"]["Ref"] != f"refs/tags/sdk/go/{name}/v0.2.0":
        raise SystemExit(f"unexpected tag ref for {name}")
    if name == "factschema":
        old_dir = Path(info["Dir"]) / "schema"

old = {p.name: json.loads(p.read_text()) for p in old_dir.glob("*.schema.json")}
paths = subprocess.check_output(
    ["git", "ls-tree", "-r", "--name-only", base, "factschema/schema"],
    text=True,
).splitlines()
new = {
    Path(path).name: json.loads(
        subprocess.check_output(["git", "show", f"{base}:{path}"])
    )
    for path in paths
    if path.endswith(".schema.json")
}
print("counts", len(old), len(new))
print("added", sorted(new.keys() - old.keys()))
print("removed", sorted(old.keys() - new.keys()))

def changes(
    left: Any, right: Any, pointer: str = ""
) -> Iterator[tuple[str, str, Any]]:
    """Yield parsed JSON changes with JSON Pointer locations."""
    if isinstance(left, dict) and isinstance(right, dict):
        for key in sorted(left.keys() | right.keys()):
            child = pointer + "/" + key.replace("~", "~0").replace(
                "/", "~1"
            )
            if key not in left:
                yield child, "added", right[key]
            elif key not in right:
                yield child, "removed", left[key]
            else:
                yield from changes(left[key], right[key], child)
    elif left != right:
        yield pointer, "changed", {"old": left, "new": right}

for name in sorted(old.keys() & new.keys()):
    for pointer, kind, value in changes(old[name], new[name]):
        print(name, pointer, kind, json.dumps(value, sort_keys=True))
PY
```

The observed output at this baseline is 176 old schemas and 179 new schemas.
Three files were added: `config_key_divergence.v1.schema.json`,
`reducer_code_drifted_finding.v1.schema.json`, and
`ref_overlay.branch_observation.v1.schema.json`. No file was removed. The
complete parsed-JSON diff for the four existing changed files is:

| Schema | Added JSON pointer | Value |
| --- | --- | --- |
| `repository.v1.schema.json` | `/properties/empty_scope` | `{"type":["boolean","null"]}` |
| `service_catalog.entity.v1.schema.json` | `/properties/workload_repository_id` | `{"type":["string","null"]}` |
| `service_catalog.repository_link.v1.schema.json` | `/properties/workload_repository_id` | `{"type":["string","null"]}` |
| `vulnerability.suppression.v1.schema.json` | `/properties/scope/properties/workload_repository_id` | `{"type":["string","null"]}` |

These are additive optional nullable fields in the schema comparison. The
result does not prove that the original core accepts all new SDK outputs; a
qualified consumer/core pair still needs its own contract and runtime proof.
