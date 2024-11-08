/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package local

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/log"
)

// Handles locking references

type lock struct {
	since time.Time
}

func (s *store) tryLock(ctx context.Context, ref string) error {
	s.locksMu.Lock()
	defer s.locksMu.Unlock()

	log.G(ctx).WithField("ref", ref).
		WithField("NAMESPACE_DONT_LOCK_REFS", os.Getenv("NAMESPACE_DONT_LOCK_REFS")).
		Infof("niklas test")

	if dontLock := os.Getenv("NAMESPACE_DONT_LOCK_REFS"); dontLock != "" {
		refs := strings.Split(dontLock, ",")
		if slices.Contains(refs, ref) {
			return nil
		}
	}

	if strings.Contains(ref, "sha256:bd9ddc54bea929a22b334e73e026d4136e5b73f5cc29942896c72e4ece69b13d") {
		// layer-sha256:bd9ddc54bea929a22b334e73e026d4136e5b73f5cc29942896c72e4ece69b13d is problematic but should be excluded above

		log.G(ctx).WithField("ref", ref).
			WithField("NAMESPACE_DONT_LOCK_REFS", os.Getenv("NAMESPACE_DONT_LOCK_REFS")).
			Infof("should not lock this ref")
	}

	if v, ok := s.locks[ref]; ok {
		// Returning the duration may help developers distinguish dead locks (long duration) from
		// lock contentions (short duration).
		now := time.Now()
		return fmt.Errorf(
			"ref %q locked for %s (since %s): %w", ref, now.Sub(v.since), v.since,
			errdefs.ErrUnavailable,
		)
	}

	s.locks[ref] = &lock{time.Now()}
	return nil
}

func (s *store) unlock(ref string) {
	s.locksMu.Lock()
	defer s.locksMu.Unlock()

	delete(s.locks, ref)
}
