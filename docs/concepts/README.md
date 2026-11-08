# Concepts

This repo is a set of agreements, not a program. Three Go modules pin
down the shapes that collectors, the Eshu core reducer, and thin API
clients swap with each other. Each side imports the shapes it needs and
nobody has to import Eshu core to get them.

## The three contracts

**`collector` is the deal between an out-of-tree collector and its host.**
It defines the request and result records, the validators that fail
closed, and the paging helpers (`PageBoundForStdout`, `FactWireBytes`,
`SplitFacts`) that size every page against the host's bound. Around
that core sit `transport` for HTTP work, `command` for running one
bounded request over stdio, `entrypoints` for generating a collector
command from a manifest, plus `marker` and `conformance`.

**`factschema` is the collector-to-reducer payload contract** (Contract
System v1). It carries the shared `Envelope` and one typed struct per
fact kind under `<family>/v1`. A missing required field is a classified
decode error naming the field, never a silent zero.

**`wirecontract` is a byte-pinned copy of the client-facing API shapes:**
response, truth, and error envelopes, query profiles, freshness values,
the MCP tool-definition shape, and the client environment registry.
Golden files pin every JSON field to core bytes.

## Where each contract stops

Don't use these modules to reach into core behaviour. Admission, grants,
graph writes, queues, and durable state stay in core on purpose. The
package comment in `wirecontract/doc.go` lists exactly what's left out
of `wirecontract` and why; [Conformance grants](../conformance-grants.md)
explains the one trusted-context seam in `collector`.

## The story in one paragraph

A collector emits facts shaped by `collector` and `factschema`; the
reducer reads those same shapes and turns them into a queryable graph;
API clients decode the answers with `wirecontract`. You only get types
and validators here — no running code, no service. The diagrams are in
[Diagrams](../diagrams/README.md).
