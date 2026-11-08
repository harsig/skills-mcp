# skills-mcp

**skills-mcp** is a small Go MCP server (stdio transport, official Go SDK) that serves skills as tools. A skill is a folder with a `SKILL.md`; calling its tool returns the file's content. The available skills are defined by one YAML file passed as the only argument.

## Features

- **One tool per skill.** The tool name is the skill name, the tool description is the skill description (both from the `SKILL.md` frontmatter, falling back to the folder name, or the repository name for a skill at the root of a git source, and an empty description), so the model can see when to load a skill. Tools take no arguments (input schema: an empty object with `additionalProperties: false`), are annotated read-only and idempotent, and return the unmodified `SKILL.md` as a single text block.
- **Sources:** a list of local paths or git repositories. A source lists `folders` relative to its root, each containing a `SKILL.md`. Without `folders`, the source root itself must contain `SKILL.md`.
- **Git sources:** `skills-mcp --setup` clones each repository into `.skills-mcp/repos/<name>-<hash>/` next to the configuration file (the hash covers URL and `ref`), or fetches it if it is already there, and checks out the `ref` (tag, branch or commit; default: the remote's default branch). A normal run uses these clones and never touches the network; it fails with a hint to run `--setup` if a clone is missing, which includes after a `ref` change. Requires the `git` binary; authentication is whatever your git setup provides (SSH for private repositories).
- **Dynamic content:** the tool list is built once at startup. `SKILL.md` is read from disk on every call, so local edits show up without a restart.
- **Config checks at startup:** unknown keys fail; a source needs exactly one of `path` or `git`; a missing `SKILL.md`, an invalid skill name (`^[a-zA-Z0-9_-]+$`) or a duplicate skill name fails. Relative paths are resolved from the config file's directory, `~` is expanded.
- **Errors:** usage errors print the usage and exit with code 2; configuration, setup and startup problems exit with code 1 and a clear message. A failing tool call (e.g. the file vanished) returns an MCP tool error and never crashes the server. Logs and errors go to stderr, prefixed with `skills-mcp: `.
- **CLI:** `skills-mcp <config_file>` runs the server; `skills-mcp --setup <config_file>` clones or updates git sources and exits; `skills-mcp --version` prints the version, which `build.sh` takes from `git describe`. Options come before the config file; any other usage is a usage error. In server mode stdout carries only the MCP protocol. The server stops on stdin close, SIGINT or SIGTERM.
- **Platform:** Linux and other Unix systems.
- **Code layout:** a single `main` package: `main.go` (wiring), `config.go`, `git.go`, `skill.go`, `server.go`, with table-driven tests per file. Git sources, strict YAML loading and path resolution come from the shared module `github.com/harsig/harness-kit`. Tools are registered with the SDK's low-level `Server.AddTool` and an empty object `InputSchema`. No interfaces or abstractions beyond what the tests need.
- **Out of scope:** config hot reload, `${VAR}` expansion, MCP resources or prompts, serving files other than `SKILL.md`, authentication settings, HTTP transport.

## Configuration Schema

```yaml
sources:                        # list of sources, each has exactly one of path | git
  - path: ~/my-skills           # local directory
    folders:                    # optional, relative to the source root, each with a SKILL.md
      - code-review
      - release-notes

  - path: ./single-skill        # no folders: ./single-skill/SKILL.md is the skill

  - git:                        # cloned into ./.skills-mcp/repos/ by --setup
      url: git@github.com:acme/skills.git
      ref: v1.2.0               # optional: tag | branch | commit (default: remote default branch)
    folders: [skills/pdf, skills/docx]

  - git:
      url: git@github.com:acme/one-skill.git   # no folders: the repository root is the skill
```
