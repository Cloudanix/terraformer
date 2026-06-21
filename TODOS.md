# TODOS

Deferred work captured during reviews. Each entry: What / Why / Context / Depends on.

---

## T4 — Add a fixed-region scope and the globalaccelerator service

**What:** Support AWS services whose control plane lives in a single fixed
region that is neither us-east-1 nor "global" — starting with
`globalaccelerator` (us-west-2 only), and emit `aws_globalaccelerator_accelerator`
(+ listener/endpoint_group).

**Why:** terraformer's region model only has regional / global (aws-global pass)
/ eastOnly (us-east-1 pass) — see `serviceScope` (providers/aws/scope.go) and
`SupportedGlobalResources` / `SupportedEastOnlyResources` (aws_provider.go).
Global Accelerator's API only answers in **us-west-2**; marking it regional
fails in every other region, and eastOnly hits the wrong region. The same
pattern will recur for `route53-recovery-*` and `cur` (us-east-1, already an
eastOnly case) — so a general "fixed region" binding is worth doing once.

**Context:** The region passes are driven in `cmd/provider_cmd_aws.go`
(`parseAndGroupResources` → global/eastOnly/regional groups, each a separate
`Import()` pass). Adding a fixed-region binding means: (1) a new `scopeFixed`
(or a `serviceScope` variant carrying a region string), (2) a new group +
pass in provider_cmd_aws.go that pins the region, (3) the scope_test.go
assertion updated to cover it. Once the machinery exists, the generator itself
is trivial (ListAccelerators paginator, ARN import ID). The SDK module
`github.com/aws/aws-sdk-go-v2/service/globalaccelerator` is already fetched.

**Depends on / blocked by:** Touches the core region-pass machinery in
cmd/provider_cmd_aws.go + scope.go; do it as its own change, not folded into a
single-service PR. Deferred from the P1 batch (2026-06-21) for this reason.

---

## T5 — Implement servicequotas with a non-default filter

**What:** Add a `servicequotas` generator emitting `aws_servicequotas_service_quota`
— but only for quotas that have actually been changed, not the thousands of
untouched defaults.

**Why:** `ListServiceQuotas` returns every quota for every service. A naive
"comprehensive" import = ListServices → ListServiceQuotas per service (N+1 API
explosion) and thousands of `aws_servicequotas_service_quota` resources that are
just AWS defaults — noise, not managed infrastructure. Deferred from the P1
batch (2026-06-21) because it needs a filtering design, unlike the other
services.

**Context:** SDK module `github.com/aws/aws-sdk-go-v2/service/servicequotas` is
already fetched. The import ID for aws_servicequotas_service_quota is
"<service-code>/<quota-code>". Filtering approaches to evaluate: (a)
`ListRequestedServiceQuotaChangeHistory` to find quotas the account has
requested changes for; (b) compare each quota's `Value` against
`GetAWSDefaultServiceQuota` and emit only where they differ (extra API call per
quota); (c) only emit quotas where `Adjustable` && a change request exists.
Option (a) is likely the cheapest signal for "quotas the user manages". Follow
the standard recipe (generator + registry + serviceScope regional entry + docs)
once the filter is decided.

**Depends on / blocked by:** None technical — just the filtering-policy
decision. Independent of the region-machinery work in T4.

---

## T2 — Stream resources to disk per service (bound memory)

**What:** Free each service's resources from memory after its files are written,
instead of holding every resource's fully-refreshed state in memory until the
end of the run.

**Why:** Terraformer refreshes every discovered resource through the
`terraform-provider-aws` gRPC plugin and holds the full state in memory until
the output stage. On a large account imported with full service coverage
(`--resources=*`), this is reachable OOM territory. The
[AWS full-coverage plan](docs/aws-full-support/plan.md) roughly doubles the
number of services, making this likely rather than theoretical.

**Context:** The accumulate-everything-then-write pattern lives in the import
pipeline: `cmd/import.go` (`initAllServicesResources` → `ProcessResources`) holds
`ProvidersMapping.Resources`, and `terraformutils/terraformoutput/` writes at the
end. Today each service appends to `Service.Resources`
(`terraformutils/service.go`) and nothing is released between services. Starting
point: write + drop a service's resources right after its
`InitResources`/refresh completes, rather than retaining the whole set. Confirm
the refresh step (`providerwrapper`) doesn't need cross-service state first.
Interim mitigation already documented in the plan: scope `--resources` instead of
`*`.

**Depends on / blocked by:** Touches the core output pipeline; do after the
AWS-coverage generators land so the change is validated against a realistically
large resource set.

---

## T3 — Replace `context.TODO()` with a timeout/cancellable context

**What:** Sweep every generator's `context.TODO()` and give the import run a
real context with a timeout (and Ctrl-C cancellation), threaded from the CLI.

**Why:** Every AWS generator calls AWS APIs with `context.TODO()`
(e.g. `providers/aws/ecr.go`, and the §5 generator skeleton in the plan). A
single hung AWS call hangs the entire import with no deadline. The full-coverage
plan multiplies the number of such calls, widening the hang surface.

**Context:** Pattern is repo-wide (`grep -rn 'context.TODO()' providers/`), not
AWS-only, which is why it was deferred out of the AWS-coverage plan rather than
fixed piecemeal. Right scope: thread a single `context.Context` from
`cmd/root.go` / `cmd/import.go` down through `ProviderGenerator.InitService` /
`ServiceGenerator.InitResources` (signature change), with a CLI
`--timeout` flag defaulting to something generous. Do it as one mechanical sweep
so existing and new generators move together — avoids a half-converted codebase.

**Depends on / blocked by:** Signature change to `InitResources()` across all
providers (not just AWS); coordinate so it doesn't collide with the AWS-coverage
PRs adding new generators. Cheapest to do either before the coverage work starts
or after it fully lands, not interleaved.
