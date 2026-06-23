# §3 gap-list status

## Done
- `current-coverage.txt` — the **141** `azurerm_*` resource types terraformer
  emits today (`grep -rhoE '"azurerm_[a-z0-9_]+"' providers/azure/*.go`). Real,
  reproducible, committed.

## Pending: provider schema dump (the (a) side of the §3 diff)
Not yet generated. `terraform providers schema -json` needs the provider plugin
to launch, and the command sandbox breaks the terraform↔provider go-plugin gRPC
handshake ("Unrecognized remote plugin message" — the plugin never receives the
magic cookie). Same plugin blocker recorded for the AWS full-support work.

What already works offline (no network):
- terraform v1.5.7 + jq present.
- azurerm provider **v4.78.0** cached locally at
  `/Users/puru/code/devops-playbooks/infra/.terraform/providers/.../darwin_arm64/`.
- A filesystem-mirror CLI config + `terraform init` succeed against that cache.

Only the final `providers schema` step is blocked, and only because it execs the
plugin under the sandbox.

## To finish (run OUTSIDE the sandbox, e.g. `!`-prefix in session)
```
bash <scratch>/tfschema/gen-gap.sh
```
(staged this session under the session scratchpad; mirror + tfrc + main.tf
alongside it). It writes `tf-azurerm-all-resources.txt`,
recomputes `current-coverage.txt`, and produces `missing-resources.txt`
(the gap), then prints counts. Re-stage if the scratchpad is gone:
- `main.tf`: `required_providers { azurerm = { source="hashicorp/azurerm", version="= 4.78.0" } }`
- `tfcli.tfrc`: `provider_installation { filesystem_mirror { path = "<cache>/providers" } direct { exclude=["registry.terraform.io/*/*"] } }`
- `TF_CLI_CONFIG_FILE=<tfrc> terraform init && terraform providers schema -json | jq ...`

## Notes
- Binary-strings fallback (`grep -aoE 'azurerm_[a-z0-9_]+' <provider-bin>`)
  rejected: yields 1024 names mixing **resources + data sources**, and 1024 <
  the ~1100 resources v4.x actually has (Go string table incomplete) → would
  pollute the gap. Schema dump is the only authoritative source.
- Version used (4.78.0) ≠ the 4.0.0 floor in `plan.md` §3. Reconcile the
  documented floor once the real list lands; 4.78.0 gives a superset.
