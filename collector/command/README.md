# Bounded command runner

`Run` reads exactly one JSON `collector.Request` from a reader and writes one
JSON `collector.Result` plus a newline to a writer. Callers supply explicit
supported protocols and positive input/output byte budgets. The request budget
rejects `math.MaxInt64` because the runner reads `limit + 1` bytes to detect
overflow. The output budget includes the newline.

The runner uses the earlier of the parent context deadline and the claim
deadline. It checks cancellation before invoking the handler and before
writing output. Cancellation does not forcibly interrupt an arbitrary blocked
reader or a handler that ignores its context. A supervising host must enforce
the process deadline and output cap.

For v1alpha3, the runner rejects invalid UTF-8 request bytes and invalid
native coverage strings before JSON encoding can replace them with U+FFFD.
This is an encoding guard; the host still owns admission, grants, claim
identity, fact validation, and process supervision. A handler may
return `retryable` or `terminal` as a normal result with a nil Go error. A
non-nil Go error fails the command. Returned errors expose only a safe stage
name, never input, config, or provider error text. A writer can still partially
write before returning an error; callers must treat a write error as failed.
