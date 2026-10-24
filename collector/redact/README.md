# Redaction markers

`redact` builds the legacy HMAC-SHA256 marker from a deployment-scoped key,
normalized reason and source labels, and raw bytes. `String` and `Bytes` have
the same digest for the same bytes. The key is copied on construction and may
be shared by concurrent calls. The returned `Value` contains only the marker
and classification labels; callers must never put raw secrets in the labels.
The package does not log or emit telemetry.

The zero `Key` preserves legacy marker behavior but is not suitable for
protecting sensitive inputs. Construct a key with `NewKey` and reject `IsZero`
before source acquisition in production collectors. `String` and `Bytes` retain
the legacy zero-key behavior for compatibility. `NewKey` rejects whitespace-only
material but preserves every byte of accepted material.

This helper does not acquire provider data, choose which values are sensitive,
or change the collector wire protocol. Collector implementations own those
policies and must avoid retaining or logging raw values.
