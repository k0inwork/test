🧹 Improve `RegisterWithDiscovery` retry loop

🎯 **What:** The code health issue addressed
Replaced the hardcoded `time.Sleep(5 * time.Second)` in the service discovery registration loop with an exponential backoff mechanism. Also fixed a bug where the code considered the registration successful if there was no network error but the server returned a non-200 status code.

💡 **Why:** How this improves maintainability
The exponential backoff prevents the service from repeatedly hammering the registry when it's down, which is a better practice for network resilience and avoids thundering herd problems when multiple services start simultaneously while the registry is offline or restarting. Furthermore, correctly asserting on HTTP 200 prevents premature or false successes.

✅ **Verification:** How you confirmed the change is safe
- Preserved existing functionality: it continues to retry registering in the background until successful.
- Wrote an integration-style test in `logger_test.go` (`TestRegisterWithDiscovery_Retry`) that spins up a mock registry. The mock registry intentionally fails the first two attempts and succeeds on the third. The test verifies that `RegisterWithDiscovery` successfully registers after multiple attempts.

✨ **Result:** The improvement achieved
A more resilient service discovery registration phase with fewer tight loops and reliable success detection.
