# opentracing
[Opentracing](https://github.com/opentracing/opentracing-go) tracer for [Hertz](https://github.com/cloudwego/hertz)

## Server usage
```go
func main() {
    ...
    h := server.Default(server.WithTracer(hertz.NewTracer(ht, func(c *app.RequestContext) string {
        return "test.hertz.server" + "::" + c.FullPath()
    })))
    h.Use(hertz.ServerCtx())
    ...
}
```

## Client usage
```go
func main() {
    ...
    c, _ := client.NewClient()
    c.Use(hertz.ClientTraceMW, hertz.ClientCtx)
    ...
}
```

## Example
[Executable Example](https://github.com/cloudwego/hertz-examples/tree/main/tracer)
