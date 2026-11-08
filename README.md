# skills-mcp

A small MCP server that serves **skills** as tools. A skill is a folder containing a `SKILL.md`. Each skill becomes one tool. The tool's name and description come from the skill's frontmatter, and calling the tool returns the content of `SKILL.md`. This lets any MCP client discover skills and load one when it needs it.

- stdio transport, built on the official Go MCP SDK
- skills come from local directories or git repositories
- a single YAML file configures everything
- Linux and other Unix systems

## Build

Requires Go (see [src/go.mod](src/go.mod) for the version). Shared code comes from [harness-kit](https://github.com/harsig/harness-kit). Until it has tagged releases, it is tracked at its `main` branch; update it with `go get github.com/harsig/harness-kit@main` in `src/`.

```sh
./build.sh                # go vet + static binary at bin/skills-mcp
cd src && go test ./...   # tests
```

## Writing a skill

A skill is a directory with a `SKILL.md` file. Frontmatter is optional:

```markdown
---
name: code-review
description: Checklist for reviewing pull requests. Use when asked to review code.
---

# Code review

1. ...
```

- `name` becomes the tool name. If it is missing, the folder name is used (for a skill at the root of a git source: the repository name). It must match `^[a-zA-Z0-9_-]+$` and be unique across all sources.
- `description` becomes the tool description, which is what the model uses to decide when to load the skill. If it is missing, the description is empty.
- The tool takes no arguments and returns the whole `SKILL.md`, frontmatter included, as one text block. It is annotated read-only and idempotent, a hint clients can use to treat it as safe to call.

## Configuration

Pass one YAML file to the server. It lists the sources, and each source has exactly one of `path` or `git`:

```yaml
sources:
  # Local directory with several skills
  - path: ~/my-skills
    folders:              # relative to the source root, each must contain SKILL.md
      - code-review
      - release-notes

  # Local directory that is itself a skill (no folders: ./single-skill/SKILL.md)
  - path: ./single-skill

  # Git repository with several skills
  - git:
      url: git@github.com:acme/skills.git
      ref: v1.2.0         # optional tag, branch or commit; default is the remote's default branch
    folders: [skills/pdf, skills/docx]

  # Git repository whose root is a skill
  - git:
      url: git@github.com:acme/one-skill.git
```

Rules:

- Relative paths are resolved from the directory that contains the config file. `~` is expanded.
- Unknown keys are an error. So are a missing `SKILL.md`, an invalid skill name and a duplicate skill name. The server checks all of this at startup and exits with code 1 and a message if anything is wrong.

## Usage

```text
skills-mcp <config_file>          run the MCP server (stdio)
skills-mcp --setup <config_file>  clone or update git sources and exit
skills-mcp --version              print the version
```

Options go before the config file. A usage error exits with code 2, any other error with code 1.

### 1. Fetch git sources (only needed if you use them)

```sh
skills-mcp --setup ~/.config/skills-mcp/config.yaml
```

- Missing repositories are cloned into `.skills-mcp/repos/` next to the config file.
- Existing repositories are fetched, and the configured `ref` is checked out.
- Run `--setup` again to get the latest version of a branch. After changing a `ref`, the server refuses to start until you do, because the clone folder depends on URL and `ref`.
- Authentication is whatever your local `git` provides. SSH is the expected way to reach private repositories.

A normal server run never uses the network. If a git source has not been cloned yet, the server fails to start and tells you to run `--setup`.

### 2. Register the server with your MCP client

**Claude Code:**

```sh
claude mcp add skills -- /path/to/skills-mcp/bin/skills-mcp ~/.config/skills-mcp/config.yaml
```

**Clients configured with JSON** (Claude Desktop, Cursor, and others):

```json
{
  "mcpServers": {
    "skills": {
      "command": "/path/to/skills-mcp/bin/skills-mcp",
      "args": ["/home/me/.config/skills-mcp/config.yaml"]
    }
  }
}
```

Use absolute paths here, because clients may start the server from any working directory.

### 3. Use it

The client lists one tool per skill. When a task matches a skill's description, the model calls that tool and gets the skill's instructions back.

## Behavior notes

- **Tool list is fixed at startup.** If you add or remove skills, or change a skill's `name` or `description`, restart the server (usually by reconnecting the client).
- **Content is read on every call.** Edits to the body of a `SKILL.md` show up immediately, with no restart.
- **Errors during a call** (for example, a file was deleted) come back as an MCP tool error. The server keeps running.
- **Logging:** stdout carries only the MCP protocol. Logs and errors go to stderr.
- The server stops when stdin closes or on SIGINT/SIGTERM.

## Out of scope

Config hot reload, `${VAR}` expansion, MCP resources and prompts, serving files other than `SKILL.md`, authentication settings and HTTP transport.
