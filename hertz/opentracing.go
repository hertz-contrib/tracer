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
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/tracer/stats"
	"github.com/cloudwego/hertz/pkg/common/tracer/traceinfo"
	"github.com/opentracing/opentracing-go"
)

type commonTracer struct {
	tracer            opentracing.Tracer
	formOperationName func(ctx *app.RequestContext) string
}

func (c *commonTracer) newEventSpan(operationName string, hi traceinfo.HTTPStats, start, end stats.Event, parentContext opentracing.SpanContext) opentracing.Span {
	var opts []opentracing.StartSpanOption
	event := hi.GetEvent(start)
	if event == nil {
		return nil
	}
	startTime := opentracing.StartTime(event.Time())
	opts = append(opts, startTime)
	if parentContext != nil {
		opts = append(opts, opentracing.ChildOf(parentContext))
	}
	span := c.tracer.StartSpan(operationName, opts...)
	span.FinishWithOptions(opentracing.FinishOptions{FinishTime: hi.GetEvent(end).Time()})
	return span
}
