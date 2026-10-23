# Collector command boundary

An out-of-tree collector command can use `collector/command.Run` to exchange
one JSON request and one JSON result with a core host. The input is
`collector.Request`: protocol version, core-owned claim, declared contract, and
opaque configuration. The output remains the existing `collector.Result` wire
shape; this API adds no result fields or wire version.

Set `Options.SupportedProtocols` to an explicit list of known versions and set
positive `MaxRequestBytes` and `MaxResultBytes`. The request limit must be below
`math.MaxInt64`. The result limit counts the encoded newline. A handler receives
the earlier parent/claim deadline. It should cooperate with cancellation and
return a retryable or terminal result with a nil Go error when that is the
application outcome. A non-nil Go error is a command failure.

The runner decodes opaque `Config` numbers with `json.Decoder.UseNumber`.
Handlers should convert `json.Number` explicitly rather than type asserting
`float64`; this preserves integers beyond the float64 exact range. The public
request schema is `collector/schema/collector-request-v1alpha1.schema.json`,
and `collector.RequestJSONSchema()` generates the same artifact. Request
validation checks known and matching protocols plus the existing claim
structure. It does not validate configuration or fact declarations.

The host remains responsible for admission, grants, result validation, process
watchdogs, and output limits. This package does not guarantee prompt shutdown
of a blocking reader or a noncooperative handler, process tree cleanup, or a
hard memory bound. Provider byte budgets need representative measurement before
adoption. No runtime telemetry is added by this library boundary.
