# Commit Message Convention

This project follows the [Conventional Commits](https://www.conventionalcommits.org/) specification for commit messages.

## Format

Each commit message consists of a **header**, a **body**, and a **footer**:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Header

The header is mandatory and must conform to the format: `<type>(<scope>): <subject>`

#### Type

Must be one of the following:

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Changes that don't affect code meaning (formatting, whitespace, etc.)
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **perf**: Performance improvement
- **test**: Adding or updating tests
- **build**: Changes to build system or dependencies (e.g., go.mod, Dockerfile, Taskfile)
- **ci**: Changes to CI configuration files and scripts (GitHub Actions)
- **chore**: Other changes that don't modify src or test files
- **revert**: Reverts a previous commit

#### Scope

The scope is optional but recommended. It should be the name of the component affected:

- **api**: API/endpoint changes
- **handlers**: Handler functions
- **services**: Service layer (jwt, metrics, body, tls)
- **models**: Data models and structs
- **docker**: Docker-related changes
- **k8s**: Kubernetes configurations
- **metrics**: Prometheus metrics
- **jwt**: JWT token handling
- **tls**: TLS/HTTPS functionality
- **templates**: HTML templates
- **deps**: Dependency updates
- **config**: Configuration changes
- **docs**: Documentation
- **tests**: Testing
- **ci**: CI/CD
- **devex**: Developer experience

#### Subject

The subject contains a succinct description of the change:

- Use the imperative, present tense: "change" not "changed" nor "changes"
- Minimum 10 characters
- Maximum 72 characters
- Don't capitalize the first letter
- No period (.) at the end

### Body

The body is optional. Use it to explain the motivation for the change and contrast it with previous behavior.

- Wrap at 100 characters per line
- Use the imperative, present tense
- Include motivation for the change and contrast with previous behavior

### Footer

The footer is optional. Use it for:

- **Breaking Changes**: Start with `BREAKING CHANGE:` followed by a description
- **Issue References**: Reference GitHub issues (e.g., `Fixes #123`, `Closes #456`, `Relates to #789`)

## Examples

### Simple feature commit

```
feat(handlers): add custom status code support via header

Allow clients to control the HTTP response status code by sending
the x-set-response-status-code header with a value between 200-599.

Closes #42
```

### Bug fix with scope

```
fix(jwt): handle missing authorization header gracefully

Previously, the JWT decoder would panic if the Authorization header
was missing. Now it returns nil without error, allowing the handler
to continue processing the request.
```

### Documentation update

```
docs: update README with TLS configuration examples

Add detailed examples for enabling HTTPS with both self-signed and
custom certificates.
```

### Breaking change

```
feat(api): change response format to include request body

BREAKING CHANGE: The response JSON structure now includes a new
"body" field at the root level. Clients parsing the response must
update their JSON models to accommodate this change.

Previously:
{
  "request": {...},
  "server": {...}
}

Now:
{
  "request": {...},
  "server": {...},
  "body": {...}
}

Fixes #67
```

### Dependency update

```
build(deps): update fiber to v2.52.10

Update gofiber/fiber from v2.52.9 to v2.52.10 to include security
fixes and performance improvements.
```

### Chore commit

```
chore(ci): update Go version to 1.25.6 in workflows
```

### Revert commit

```
revert: feat(handlers): add custom status code support

This reverts commit 1234567890abcdef. The feature caused
compatibility issues with certain HTTP clients.
```

## Validation

Commit messages are automatically validated using:

- **commitlint**: Validates commit message format (runs via pre-commit hook)
- **conventional-pre-commit**: Enforces conventional commits standard

To validate manually:

```bash
# Install commitlint CLI (requires Node.js)
npm install -g @commitlint/cli @commitlint/config-conventional

# Validate a commit message
echo "feat(api): add new endpoint" | commitlint

# Validate last commit
commitlint --from HEAD~1
```

## Tips

1. **Keep commits atomic**: Each commit should represent a single logical change
2. **Write meaningful subjects**: The subject should clearly describe what changed
3. **Use the body for context**: Explain "why" in the body, not just "what"
4. **Reference issues**: Always link to related issues or pull requests
5. **Use breaking change footer**: Clearly mark breaking changes for semantic versioning
6. **Run pre-commit hooks**: Let the hooks validate your commits before pushing

## Resources

- [Conventional Commits Specification](https://www.conventionalcommits.org/)
- [Angular Commit Guidelines](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit)
- [Semantic Versioning](https://semver.org/)
- [How to Write a Git Commit Message](https://chris.beams.io/posts/git-commit/)
