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
//

package hertzobs

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/instrumentation/internal"
	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/cloudwego/hertz/pkg/common/tracer/stats"
	prom "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"strconv"
)

var _ label.LabelControl = OtelLabelControl{}

type OtelLabelControl struct {
	tracer                   trace.Tracer
	shouldIgnore             ConditionFunc
	serverHttpRouteFormatter func(c *app.RequestContext) string
}

func NewOtelLabelControl(tracer trace.Tracer, shouldIgnore ConditionFunc, serverHttpRouteFormatter func(c *app.RequestContext) string) OtelLabelControl {
	return OtelLabelControl{
		tracer:                   tracer,
		shouldIgnore:             shouldIgnore,
		serverHttpRouteFormatter: serverHttpRouteFormatter,
	}
}

func (o OtelLabelControl) InjectLabels(ctx context.Context) context.Context {
	c, ok := ctx.Value(requestContextKey).(*app.RequestContext)
	if ok == false {
		return ctx
	}
	if o.shouldIgnore(ctx, c) {
		return ctx
	}
	tc := &internal.TraceCarrier{}
	tc.SetTracer(o.tracer)

	return internal.WithTraceCarrier(ctx, tc)
}

func (o OtelLabelControl) ExtractLabels(ctx context.Context) []label.CwLabel {
	c, ok := ctx.Value(requestContextKey).(*app.RequestContext)
	if ok == false {
		return nil
	}
	if o.shouldIgnore(ctx, c) {
		return nil
	}
	// trace carrier from context
	tc := internal.TraceCarrierFromContext(ctx)
	if tc == nil {
		logging.Debugf("get tracer container failed")
		return nil
	}

	ti := c.GetTraceInfo()
	st := ti.Stats()

	if st.Level() == stats.LevelDisabled {
		return nil
	}

	httpStart := st.GetEvent(stats.HTTPStart)
	if httpStart == nil {
		return nil
	}

	// span
	span := tc.Span()
	if span == nil || !span.IsRecording() {
		return nil
	}

	// span attributes from original http request
	if httpReq, err := adaptor.GetCompatRequest(c.GetRequest()); err == nil {
		span.SetAttributes(semconv.NetAttributesFromHTTPRequest("tcp", httpReq)...)
		span.SetAttributes(semconv.EndUserAttributesFromHTTPRequest(httpReq)...)
		span.SetAttributes(semconv.HTTPServerAttributesFromHTTPRequest("", o.serverHttpRouteFormatter(c), httpReq)...)
		span.SetStatus(semconv.SpanStatusFromHTTPStatusCode(c.Response.StatusCode()))
	}

	// span attributes
	attrs := []attribute.KeyValue{
		semconv.HTTPURLKey.String(c.URI().String()),
		semconv.NetPeerIPKey.String(c.ClientIP()),
		semconv.HTTPStatusCodeKey.Int(c.Response.StatusCode()),
	}
	span.SetAttributes(attrs...)

	injectStatsEventsToSpan(span, st)

	if panicMsg, panicStack, httpErr := parseHTTPError(ti); httpErr != nil || len(panicMsg) > 0 {
		recordErrorSpanWithStack(span, httpErr, panicMsg, panicStack)
	}

	span.End(oteltrace.WithTimestamp(getEndTimeOrNow(ti)))

	metricsAttributes := semantic.ExtractMetricsAttributesFromSpan(span)
	return label.ToCwLabelsFromOtels(metricsAttributes)
}

const (
	labelMethod       = "method"
	labelStatusCode   = "statusCode"
	labelPath         = "path"
	unknownLabelValue = "unknown"
)

var _ label.LabelControl = PromLabelControl{}

type PromLabelControl struct {
}

func DefaultPromLabelControl() PromLabelControl {
	return PromLabelControl{}
}

func (p PromLabelControl) InjectLabels(ctx context.Context) context.Context {
	return ctx
}

func (p PromLabelControl) ExtractLabels(ctx context.Context) []label.CwLabel {
	c, ok := ctx.Value(requestContextKey).(*app.RequestContext)
	if ok == false {
		return nil
	}
	return genCwLabels(c)
}

// genLabels make labels values.
func genLabels(ctx *app.RequestContext) prom.Labels {
	labels := make(prom.Labels)
	labels[labelMethod] = defaultValIfEmpty(string(ctx.Request.Method()), unknownLabelValue)
	labels[labelStatusCode] = defaultValIfEmpty(strconv.Itoa(ctx.Response.Header.StatusCode()), unknownLabelValue)
	labels[labelPath] = defaultValIfEmpty(ctx.FullPath(), unknownLabelValue)

	return labels
}

func genCwLabels(ctx *app.RequestContext) []label.CwLabel {
	labels := genLabels(ctx)
	return label.ToCwLabelFromPromelabel(labels)
}

func defaultValIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
