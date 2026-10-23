# StableID extraction and migration boundary

This candidate copies `StableID` from Eshu core
`go/internal/facts/encode/stable.go` at commit
`eb62fa3056f5ed3518bb5aa676bf2e3adc96c1d5` into
`factschema/identity/stable.go`. The exported function and its private
normalizer retain the same algorithm. The four original tests are copied with
the package name changed; fixed digest vectors and edge tests add independent
checks of the candidate. Neither the core runtime nor its callers change in
this slice. No SDK version or tag is selected.

The two source files overlap temporarily. Core currently routes
`facts.StableID` through `go/internal/facts/encode.StableID`. After an immutable
SDK release exists, a later core change can forward to
`github.com/eshu-hq/eshu-sdk/factschema/identity.StableID`. That change needs
consumer compatibility proof against the exact release commit, including
representative collector fact keys and the existing core contract gates.
Only after that proof should core remove the duplicate algorithm. Permanent
parallel implementations would risk divergent durable identities.

The candidate test vectors cover nested maps and slices, map order, nil and
empty maps, time zone and nanosecond precision, JSON escaping, large integers,
unsupported-value panic, input immutability, and concurrent calls. Confluence,
Jira, and PagerDuty vectors preserve identity-map keys from their pinned core
call sites and use sample values. Their fixed expected digests were calculated
from canonical JSON outside Go. This is local compatibility evidence, not a
production consumer or deployment result.
