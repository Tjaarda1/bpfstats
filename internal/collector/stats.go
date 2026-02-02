// Copyright 2026 Alejandro de Cock Buning
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package collector

import "sync"

type Stats struct {
	mux               sync.Mutex
	count             uint64
	min, max, mean, s float64
}

func (s *Stats) Add(val float64) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.count == 0 {
		s.min, s.max = val, val
	} else {
		if val < s.min {
			s.min = val
		}
		if val > s.max {
			s.max = val
		}
	}

	s.count++
	oldMean := s.mean
	s.mean += (val - oldMean) / float64(s.count)
	s.s += (val - oldMean) * (val - s.mean)
}

func (s *Stats) Reset() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.count, s.min, s.max, s.mean, s.s = 0, 0, 0, 0, 0
}

func (s *Stats) Min() float64  { return s.min }
func (s *Stats) Max() float64  { return s.max }
func (s *Stats) Count() uint64 { return s.count }
func (s *Stats) Mean() float64 { return s.mean }
func (s *Stats) Variance() float64 {

	if s.count > 1 {
		return s.s / float64(s.count-1)
	}
	return 0
}
