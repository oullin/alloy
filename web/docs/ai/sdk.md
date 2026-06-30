# AI SDK

<!-- ref: @alloy/code-0008 -->
<!-- ref: @alloy/code-0007 -->
<!-- ref: @alloy/code-0001 -->
<!-- ref: @alloy/code-0006 -->
<!-- ref: @alloy/code-0002 -->
<!-- ref: @alloy/code-0003 -->
<!-- ref: @alloy/code-0010 -->
<!-- ref: @alloy/code-0005 -->
<!-- ref: @alloy/code-0012 -->
<!-- ref: @alloy/code-0004 -->
<!-- ref: @alloy/code-0011 -->

Alloy's AI SDK provides a unified Go API for interacting with AI providers. It mirrors the upstream `ai` package with idiomatic Go patterns.

## Agents

Agents encapsulate instructions and tools. In Alloy, you can use anonymous agents for quick tasks:

```go
agent := ai.NewAnonymousAgent(manager, "You are a helpful assistant.")
response, err := agent.Prompt(ctx, "Hello!")
```

## Images

Generate images with a fluent API:

```go
image, err := ai.Image("A donut on a counter").Landscape().Generate(ctx)
```

## Audio & Transcriptions

```go
audio, err := ai.Audio("Hello world").Female().Generate(ctx)
transcript, err := ai.Transcribe(audioFile).Generate(ctx)
```
