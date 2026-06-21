# End-to-end plan-correctness harness

This harness drives the provider through **real Terraform core** against an
in-memory mock of the Discord API. It validates the things mocked Go unit tests
cannot: that the schemas are accepted by Terraform, and that resources produce a
**clean, stable plan** with no perpetual diffs.

It does **not** validate behavior against the live Discord API (field names,
required fields, real error codes) — that still requires the gated acceptance
tests with real credentials.

## What it checks

`run.sh` performs a full lifecycle and asserts each step is clean:

1. `terraform apply` — create resources
2. `terraform plan -detailed-exitcode` — must report **no changes** (the key
   plan-correctness assertion; a perpetual diff fails the run)
3. update an attribute, re-apply, re-plan — must be clean
4. `terraform destroy`

The mock (`mock/main.go`, a separate module so it never affects the provider
build) keeps objects in memory so reads round-trip with writes.

## Running

```bash
./test/e2e/run.sh        # uses `terraform`; set TF=tofu for OpenTofu
```

Requires `terraform` (or `tofu`) and `go` on `PATH`. The script builds the
provider and mock, wires a `dev_overrides` CLI config, and points the provider's
`api_base_url` at the local mock.

> Note: the mock binds a local port and runs in the background. Run the script
> from an interactive shell or normal CI runner. Some sandboxed command runners
> reap backgrounded listeners aggressively and may report a spurious non-zero
> exit even when the lifecycle assertions passed; check the printed `PASS:` line.
