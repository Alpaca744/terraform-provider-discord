# Security Policy

## Reporting a vulnerability

Please report suspected security vulnerabilities privately rather than opening a
public issue. Use GitHub's private vulnerability reporting for this repository,
and include:

- a description of the issue and its impact,
- steps to reproduce, and
- any relevant logs with secrets redacted.

You will receive an acknowledgement, and we will coordinate a fix and disclosure
timeline with you.

## Handling of secrets

- Bot tokens, bearer tokens, and client secrets are marked sensitive in the
  provider schema and are never written to diagnostics or logs.
- Do not paste real tokens into issues, pull requests, or test fixtures. The
  acceptance tests read credentials from environment variables only.
