# specta

<p align="center">
  <a href="https://james-w.github.io/specta/">Documentation</a> - <a href="https://github.com/james-w/specta">GitHub</a>
</p>

A Go testing library that emphasizes **composition** and **reuse** through matchers and test data factories.

Build reusable, composable test components instead of duplicating assertions. Matchers compose, factories generate test data, and error messages show exactly what failed.

```go
// Define once
var IsActiveAdmin = specta.AllOf(
    UserMatches().Active(specta.IsTrue()).Matcher(),
    UserMatches().Role(specta.Equal("admin")).Matcher(),
)

// Reuse everywhere
specta.AssertThat(t, user, IsActiveAdmin)
specta.AssertThat(t, anotherUser, IsActiveAdmin)
```

## Installation

```bash
go get github.com/james-w/specta
```
