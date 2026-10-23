# Release and cutover checklist

No SDK module release is selected or published. Before selecting tags or replacing in-tree dependencies:

1. Confirm an independently reviewed source-to-target mapping, including all schemas, fixtures, public APIs, and the 179 decoder bindings.
2. Run standalone module CI and consumer compatibility tests against exact candidate commits.
3. Choose module versions and confirm the tag prefixes `collector/vX.Y.Z` and `factschema/vX.Y.Z` with the accountable owner.
4. Update Eshu consumers and prove the full contract and runtime gates on the candidate dependencies.
5. Validate deployment configuration and runtime behavior in the designated environment, with rollback to the in-tree modules available.
6. Only then document support, publication, and retirement of the old paths. Existing immutable tags must remain intact.

The collector transport helpers are copied into this candidate but still need
consumer validation and a cutover decision. Host conformance orchestration and
example entrypoints remain outside this candidate and need separate decisions
and proof.
