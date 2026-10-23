<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan:
[specs/010-direct-aws-file-upload/plan.md](specs/010-direct-aws-file-upload/plan.md)
<!-- SPECKIT END -->

## Subagent Delegation (`golang-tdd-dev`)

Whenever a task involves creating, modifying, fixing, or refactoring Go backend code (`*.go`) or Go tests (`*_test.go`), you MUST delegate the implementation to the `golang-tdd-dev` subagent via `invoke_subagent`.

