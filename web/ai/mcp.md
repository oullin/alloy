# MCP (Model Context Protocol)

<!-- ref: @alloy/code-0103 -->
<!-- ref: @alloy/code-0102 -->
<!-- ref: @alloy/code-0101 -->

Alloy provides a complete Go implementation of the MCP server specification.

## Building a Server

```go
srv := mcp.NewServer("my-app", "1.0.0")

srv.AddTool(mcp.NewTool("greet", "Greet a user",
    schema,
    func(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
        return mcp.Text("Hello!"), nil
    },
))
```
