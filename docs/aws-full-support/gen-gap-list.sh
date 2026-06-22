#!/usr/bin/env bash
# Generate the authoritative AWS coverage gap list (plan.md §3).
#
# Produces, under docs/aws-full-support/:
#   current-coverage.txt      what terraformer emits today (§3b)   [works anywhere]
#   tf-aws-all-resources.txt  every resource the aws provider has  (§3a)
#   missing-resources.txt     (a) - (b)                            (§3c)
#
# §3a needs the terraform-provider-aws plugin schema. This script also works in
# restricted environments where `terraform init` cannot reach the registry over
# TLS: it fetches the provider via curl into a packed filesystem mirror. The one
# thing it still needs is the ability to LAUNCH the provider plugin (go-plugin
# gRPC over a local socket) for `terraform providers schema`; a sandbox that
# blocks that local handshake can only produce current-coverage.txt.
set -euo pipefail

AWS_PROVIDER_VERSION="${AWS_PROVIDER_VERSION:-5.80.0}"   # supported floor; bump deliberately
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
OUT="$REPO_ROOT/docs/aws-full-support"

# (b) current terraformer coverage — no tooling needed.
grep -rhoE '"aws_[a-z0-9_]+"' "$REPO_ROOT/providers/aws/"*.go | tr -d '"' | sort -u \
  > "$OUT/current-coverage.txt"
echo "current-coverage.txt: $(wc -l < "$OUT/current-coverage.txt") resources"

# (a) every resource in the provider, via a curl-populated filesystem mirror.
OS="$(uname | tr '[:upper:]' '[:lower:]')"; ARCH="$(uname -m)"
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
WORK="$(mktemp -d)"; MIRROR="$WORK/mirror/registry.terraform.io/hashicorp/aws"; mkdir -p "$MIRROR" "$WORK/work"
ZIP="$MIRROR/terraform-provider-aws_${AWS_PROVIDER_VERSION}_${OS}_${ARCH}.zip"
curl -fsSL -o "$ZIP" \
  "https://releases.hashicorp.com/terraform-provider-aws/${AWS_PROVIDER_VERSION}/terraform-provider-aws_${AWS_PROVIDER_VERSION}_${OS}_${ARCH}.zip"
cat > "$WORK/cli.tfrc" <<EOF
provider_installation { filesystem_mirror { path = "$WORK/mirror" } }
EOF
cat > "$WORK/work/main.tf" <<EOF
terraform { required_providers { aws = { source = "hashicorp/aws", version = "= ${AWS_PROVIDER_VERSION}" } } }
EOF
export TF_CLI_CONFIG_FILE="$WORK/cli.tfrc"
( cd "$WORK/work"
  terraform init >/dev/null
  # macOS gatekeeper quarantines curl-downloaded binaries; clear it so the plugin can launch.
  find .terraform -name 'terraform-provider-aws*' -type f -exec xattr -c {} + 2>/dev/null || true
  terraform providers schema -json )  \
  | jq -r '.provider_schemas[] | .resource_schemas | keys[]' | sort \
  > "$OUT/tf-aws-all-resources.txt"
echo "tf-aws-all-resources.txt: $(wc -l < "$OUT/tf-aws-all-resources.txt") resources"

# (c) the gap.
comm -23 "$OUT/tf-aws-all-resources.txt" "$OUT/current-coverage.txt" > "$OUT/missing-resources.txt"
echo "missing-resources.txt: $(wc -l < "$OUT/missing-resources.txt") resources"
rm -rf "$WORK"
