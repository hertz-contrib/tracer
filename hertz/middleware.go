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
	"bytes"
	"context"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/tracer/stats"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/tracer/common"
	"github.com/opentracing/opentracing-go"
)

type StringHeader protocol.RequestHeader

// Visit implements the metainfo.HTTPHeaderCarrier interface.
func (sh *StringHeader) Visit(f func(k, v string)) {
	(*protocol.RequestHeader)(sh).VisitAll(
		func(key, value []byte) {
			f(string(key), string(value))
		})
}

func ServerCtx() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ctx = metainfo.FromHTTPHeader(ctx, (*StringHeader)(&c.Request.Header))
		ctx = metainfo.TransferForward(ctx)
		tc, ok := ctx.Value(traceKey).(*traceContainer)
		if !ok {
			hlog.Errorf("get tracer container failed")
			return
		}
		serverTracer := tc.serverTracer
		var operationName string
		if serverTracer.formOperationName != nil {
			operationName = serverTracer.formOperationName(c)
		}

		var opts []opentracing.StartSpanOption
		st := c.GetTraceInfo().Stats()

		// ----------------  Start ----------------------
		if st.GetEvent(stats.HTTPStart) == nil {
			return
		}

		startTime := st.GetEvent(stats.HTTPStart).Time()
		opts = append(opts, opentracing.StartTime(startTime))

		if sck, ok := metainfo.GetValue(ctx, common.SpanContextLabel); ok {
			parentContext, err := serverTracer.tracer.Extract(opentracing.Binary, bytes.NewBuffer([]byte(sck)))
			if err != nil {
				hlog.Errorf("extract SpanContext failed, %w", err)
				return
			}
			opts = append(opts, opentracing.ChildOf(parentContext))
		}

		span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, serverTracer.tracer, operationName, opts...)
		tc.span = span

		c.Next(ctx)
	}
}

func ClientCtx(next client.Endpoint) client.Endpoint {
	return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {
		if ctx == nil {
			ctx = context.Background()
		}

		metainfo.ToHTTPHeader(ctx, &req.Header)
		return next(ctx, req, resp)
	}
}
