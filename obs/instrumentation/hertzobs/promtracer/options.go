/*
 * Copyright 2022 CloudWeGo Authors
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

package promtracer

import (
	"github.com/cloudwego-contrib/cwgo-pkg/obs/meter/metric"
	prom "github.com/prometheus/client_golang/prometheus"
)

// Option opts for monitor prometheus
type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	registry *prom.Registry
	measure  metric.Measure
}

func defaultConfig() *config {
	return &config{
		registry: prom.NewRegistry(),
	}
}

// WithRegistry define your custom registry
func WithRegistry(registry *prom.Registry) Option {
	return option(func(cfg *config) {
		if registry != nil {
			cfg.registry = registry
		}
	})
}

// WithMeasure define your custom registry
func WithMeasure(measure metric.Measure) Option {
	return option(func(cfg *config) {
		cfg.measure = measure
	})
}