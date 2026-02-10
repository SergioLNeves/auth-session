# Commit Messages

Follow [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/).

## Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

## Types

- `feat` - new feature
- `fix` - bug fix
- `docs` - documentation only
- `style` - formatting, no code change
- `refactor` - code change without fix/feature
- `perf` - performance improvement
- `test` - adding/updating tests
- `build` - build system or dependencies
- `ci` - CI configuration
- `chore` - other changes

## Rules

1. Use imperative mood: "add" not "added"
2. Keep description under 50 characters
3. No period at end of description
4. Lowercase description
5. Scope is optional but helpful
6. Breaking changes: add `!` before `:` or use `BREAKING CHANGE:` footer

## Examples

```
feat(auth): add login with OAuth
fix: resolve memory leak in cache
docs(api): update endpoint documentation
refactor(utils)!: change signature of parse function

BREAKING CHANGE: parse now returns Promise
```
