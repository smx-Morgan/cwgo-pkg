/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metric

// Option opts for Measure
type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	recoders map[string]Recorder
	counter  map[string]Counter
}

func defaultConfig() *config {
	return &config{
		counter:  map[string]Counter{},
		recoders: map[string]Recorder{},
	}
}

func newConfig(opts []Option) *config {
	cfg := defaultConfig()

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}

func WithCounter(name string, counter Counter) Option {
	return option(func(cfg *config) {
		if counter != nil {
			cfg.counter[name] = counter
		}
	})
}

func WithRecorder(name string, recorder Recorder) Option {
	return option(func(cfg *config) {
		if recorder != nil {
			cfg.recoders[name] = recorder
		}
	})
}
