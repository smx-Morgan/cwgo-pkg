/*
 * Copyright 2021 CloudWeGo Authors
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

package prometheus

import (
	"github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	prom "github.com/prometheus/client_golang/prometheus"
)

var defaultBuckets = []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000}

// Option opts for monitor prometheus
type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	buckets  []float64
	registry *prom.Registry
	counter  metric.Counter
	recorder metric.Recorder
}

func defaultConfig() *config {
	return &config{
		buckets:  defaultBuckets,
		registry: prom.NewRegistry(),
	}
}

// WithHistogramBuckets define your custom histogram buckets base on your biz
func WithHistogramBuckets(buckets []float64) Option {
	return option(func(cfg *config) {
		if len(buckets) > 0 {
			cfg.buckets = buckets
		}
	})
}

// WithRegistry define your custom registry
func WithRegistry(registry *prom.Registry) Option {
	return option(func(cfg *config) {
		if registry != nil {
			cfg.registry = registry
		}
	})
}

func WithCounter(counter *prom.CounterVec) Option {
	return option(func(cfg *config) {
		cfg.registry.Register(counter)
		cfg.counter = metric.NewPromCounter(counter)
	})
}

func WithRecorder(recorder *prom.HistogramVec) Option {
	return option(func(cfg *config) {
		cfg.registry.Register(recorder)
		cfg.recorder = metric.NewPromRecorder(recorder)
	})
}
