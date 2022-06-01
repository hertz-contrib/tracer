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
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/tracer/common"
	"github.com/opentracing/opentracing-go"
)

func ClientTraceMW(next client.Endpoint) client.Endpoint {
	return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {
		// ------------- Start -------------
		operationName := "test.hertz.client" + "::" + string(req.RequestURI())
		_, ctx = opentracing.StartSpanFromContextWithTracer(ctx, opentracing.GlobalTracer(), operationName, opentracing.StartTime(time.Now()))
		span := opentracing.SpanFromContext(ctx)
		var b bytes.Buffer
		span.Tracer().Inject(span.Context(), opentracing.Binary, &b)
		ctx = metainfo.WithValue(ctx, common.SpanContextLabel, b.String())

		// ------------- Handle -------------
		err = next(ctx, req, resp)

		// ------------- Finish -------------
		span.FinishWithOptions(opentracing.FinishOptions{FinishTime: time.Now()})
		return
	}
}
