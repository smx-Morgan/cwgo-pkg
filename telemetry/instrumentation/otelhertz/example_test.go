// Copyright 2022 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelhertz

import (
	"context"
	"testing"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/instrumentation/otelhertz/testutil"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestMetricsExample(t *testing.T) {
	// test util
	tracerProvider, registry, measureClient, measureServer := testutil.OtelTestProvider()
	defer func(tracerProvider *sdktrace.TracerProvider, ctx context.Context) {
		_ = tracerProvider.Shutdown(ctx)
	}(tracerProvider, context.Background())

	// server example
	tracer, cfg := NewServerOption(WithMeasure(measureServer))
	h := server.Default(tracer, server.WithHostPorts(":39888"))
	h.Use(ServerMiddleware(cfg))
	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		hlog.CtxDebugf(c, "message received successfully")
		ctx.JSON(consts.StatusOK, "pong")
	})
	go h.Spin()

	<-time.After(time.Millisecond * 500)

	// client example
	c, _ := client.NewClient()
	c.Use(ClientMiddleware(WithMeasure(measureClient)))
	_, body, err := c.Get(context.Background(), nil, "http://localhost:39888/ping?foo=bar")
	require.NoError(t, err)
	assert.NotNil(t, body)

	// test client returns error
	_, _, err = c.Get(context.Background(), nil, "http://localhost:39887/ping?foo=bar")
	assert.NotNil(t, err)

	testerror := testutil.GatherAndCompare(
		registry, "testdata/hertz_request_metrics.txt",
		"http_server_request_count_total", "http_client_request_count_total")
	// diff meter
	assert.NoError(t, testerror)
}
