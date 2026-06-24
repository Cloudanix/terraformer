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

package terraformutils

import (
	"context"
	"testing"
)

func TestServiceContextDefaultsToBackground(t *testing.T) {
	s := &Service{}
	if s.Context() != context.Background() {
		t.Fatal("unset Context() should return context.Background()")
	}
}

func TestServiceContextRoundTrip(t *testing.T) {
	s := &Service{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.SetContext(ctx)
	if s.Context() != ctx {
		t.Fatal("Context() should return the context set via SetContext")
	}
	// Cancellation propagates through the stored context.
	cancel()
	if s.Context().Err() == nil {
		t.Fatal("cancellation should be visible via the stored context")
	}
}
