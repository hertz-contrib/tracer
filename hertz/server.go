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

package hertz

import (
	"context"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/tracer"
	"github.com/cloudwego/hertz/pkg/common/tracer/stats"
	"github.com/opentracing/opentracing-go"
)

var _ tracer.Tracer = &serverTracer{}

type serverTracer struct {
	commonTracer
}

type traceContainer struct {
	serverTracer *serverTracer
	span         opentracing.Span
}

type traceKeyType struct{}

var traceKey traceKeyType

func (s *serverTracer) Start(ctx context.Context, c *app.RequestContext) context.Context {
	ctx = context.WithValue(ctx, traceKey, &traceContainer{serverTracer: s})
	return ctx
}

func (s *serverTracer) Finish(ctx context.Context, c *app.RequestContext) {
	ctx = metainfo.FromHTTPHeader(ctx, (*StringHeader)(&c.Request.Header))
	ctx = metainfo.TransferForward(ctx)
	st := c.GetTraceInfo().Stats()

	tc, ok := ctx.Value(traceKey).(*traceContainer)
	if !ok {
		hlog.Errorf("get tracer container failed")
		return
	}
	// ----------------  Finish ----------------------
	// read span
	readSpan := s.newEventSpan("read", st, stats.ReadHeaderStart, stats.ReadBodyFinish, tc.span.Context())
	if readSpan != nil {
		readSpan.SetTag("recv_size", st.RecvSize())
	}

	// handler span
	s.newEventSpan("handler", st, stats.ServerHandleStart, stats.ServerHandleFinish, tc.span.Context())

	// write span
	writeSpan := s.newEventSpan("write", st, stats.WriteStart, stats.WriteFinish, tc.span.Context())
	if writeSpan != nil {
		writeSpan.SetTag("send_size", st.SendSize())
	}

	tc.span.FinishWithOptions(opentracing.FinishOptions{FinishTime: st.GetEvent(stats.HTTPFinish).Time()})
}

func NewDefaultTracer() tracer.Tracer {
	st := &serverTracer{}
	st.tracer = opentracing.GlobalTracer()
	st.formOperationName = func(ctx *app.RequestContext) string {
		return "test.hertz.server" + "::" + ctx.FullPath()
	}
	return st
}

func NewTracer(tracer opentracing.Tracer, formOperationName func(c *app.RequestContext) string) tracer.Tracer {
	st := &serverTracer{}
	st.tracer = tracer
	st.formOperationName = formOperationName
	return st
}
