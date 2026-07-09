// Copyright 2018 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package providerwrapper

import (
	"errors"
	"testing"
)

func TestIsExpiredCredsError(t *testing.T) {
	expired := []string{
		"operation error CloudWatch Logs: DescribeLogStreams, https response error StatusCode: 400, ExpiredToken: The security token included in the request is expired",
		"operation error EC2: DescribeInstances, ExpiredTokenException: The security token included in the request is expired",
		"api error RequestExpired: Request has expired.",
		"api error InvalidClientTokenId: The security token included in the request is invalid.",
		"The security token included in the request is invalid",
	}
	for _, msg := range expired {
		if !isExpiredCredsError(errors.New(msg)) {
			t.Errorf("expected expired-creds detection for: %s", msg)
		}
	}

	// Must NOT trigger a credential refresh — unrelated failures.
	other := []string{
		"operation error S3: ListBuckets, AccessDenied: not authorized",
		"ValidationException: Must provide at least one of: x, y",
		"dial tcp: i/o timeout",
		"",
	}
	for _, msg := range other {
		if isExpiredCredsError(errors.New(msg)) {
			t.Errorf("did not expect expired-creds detection for: %q", msg)
		}
	}
	if isExpiredCredsError(nil) {
		t.Error("nil error must not be treated as expired creds")
	}
}
