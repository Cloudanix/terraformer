#!/usr/bin/env bash
# Generate the authoritative AWS coverage gap list (plan.md §3).
#
# Produces, under docs/aws-full-support/:
#   current-coverage.txt      what terraformer emits today          (§3b)
#   tf-aws-all-resources.txt  every resource the aws provider has   (§3a)
#   missing-resources.txt     (a) - (b)                             (§3c)
#
# The resource list (§3a) is derived from the provider's own resource doc files
# (website/docs/r/*.html.markdown — one per aws_* resource) via the GitHub git
# tree API. This needs no terraform and, crucially, no provider PLUGIN: the
# `terraform providers schema` path launches the plugin over a go-plugin socket,
# which a sandbox that denies bind(2) cannot do. curl-to-GitHub is enough.
set -euo pipefail

AWS_PROVIDER_VERSION="${AWS_PROVIDER_VERSION:-5.80.0}"   # supported floor; bump deliberately
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
OUT="$REPO_ROOT/docs/aws-full-support"
REPO="hashicorp/terraform-provider-aws"
TAG="v${AWS_PROVIDER_VERSION}"

# (b) current terraformer coverage — non-test sources only.
grep -rhoE '"aws_[a-z0-9_]+"' $(ls "$REPO_ROOT/providers/aws/"*.go | grep -v '_test.go') \
  | tr -d '"' | sort -u > "$OUT/current-coverage.txt"
echo "current-coverage.txt:    $(wc -l < "$OUT/current-coverage.txt") resources"

# (a) every resource = every file under website/docs/r at the pinned tag.
# Contents API caps at 1000/dir, so resolve the dir tree SHA and read the tree.
RSHA="$(curl -fsSL "https://api.github.com/repos/$REPO/contents/website/docs?ref=$TAG" \
  | jq -r '.[] | select(.name=="r") | .sha')"
curl -fsSL "https://api.github.com/repos/$REPO/git/trees/$RSHA" \
  | jq -r '.tree[].path' \
  | sed -E 's/\.html\.markdown$//; s/\.markdown$//; s/^/aws_/' | sort -u \
  > "$OUT/tf-aws-all-resources.txt"
echo "tf-aws-all-resources.txt: $(wc -l < "$OUT/tf-aws-all-resources.txt") resources"

# (c) the gap.
comm -23 "$OUT/tf-aws-all-resources.txt" "$OUT/current-coverage.txt" > "$OUT/missing-resources.txt"
echo "missing-resources.txt:    $(wc -l < "$OUT/missing-resources.txt") resources"
