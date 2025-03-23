# Docker Test Dovecot Setup

This document explains how the test-dovecot Docker setup was configured to create a testing environment for IMAP email applications like smailnail.

## Overview

The docker-test-dovecot project provides a simple, containerized Dovecot mail server that:

1. Creates multiple test users with the password "pass"
2. Automatically generates test emails for these users
3. Provides access via IMAP, POP3, LMTP, and Sieve protocols
4. Uses self-signed SSL certificates for secure connections

## How to Use the Setup

To start the test environment:

```bash
# Clone the repository if you haven't already
git clone https://github.com/spezifisch/docker-test-dovecot.git
cd docker-test-dovecot

# Build and start the containers
docker compose up --build
```

This will:
- Build the custom Dovecot image
- Start the container with all necessary ports exposed
- Create persistent volumes for email storage and SSL certificates

## How Test-Dovecot Works

### Docker Image Configuration

The `Dockerfile` sets up a Debian-based container with:
- Dovecot IMAP, POP3, LMTP, and ManageSieve servers
- Configuration adjustments for testing purposes
- Custom scripts for mail generation and management

### User Creation

When the container starts, the `entrypoint.sh` script:

1. Generates a self-signed SSL certificate if one doesn't exist
2. Creates eight test users with the following characteristics:
   - Four regular users: `a`, `b`, `c`, `d`
   - Four receiver-only users: `rxa`, `rxb`, `rxc`, `rxd`
   - All users have the password `pass`
   - Each user gets a properly structured Maildir in their home directory

### Email Generation

The `mailgen.sh` script handles automatic mail generation:

1. Periodically checks the mail status of each user
2. If a user has no new emails, generates a test email for them
3. Uses Dovecot's delivery command to place emails directly in users' mailboxes

The emails have:
- A subject line containing "test mail" followed by a timestamp and the recipient username
- A simple body containing "this is content"
- From address of `noreply.username@mailgen.example.com`
- To address of `username@testcot`

### Periodic Mail Generation

The `cheapcron.sh` script runs in the background to:
1. Wait for Dovecot to start up
2. Run `mailgen.sh` every 60 seconds
3. Log any failures in the mail generation process

This ensures there are always test emails available for accessing via IMAP.

### Mail Cleanup

The `nukemails.sh` script provides a way to clear all emails from all users when needed.

## Accessing the Server

The server exposes several protocols:

- IMAP: Port 143 (plain) and 993 (SSL)
- POP3: Port 110 (plain) and 995 (SSL)
- LMTP: Port 24
- ManageSieve: Port 4190

## Testing with smailnail

The Dovecot server was tested with smailnail using the following command:

```bash
go run ./cmd/smailnail fetch-mail --server localhost --password pass --username a --insecure --output yaml
```

This command:
- Connects to the local Dovecot server on the default IMAP port
- Authenticates as user "a" with password "pass"
- Uses the `--insecure` flag to accept the self-signed SSL certificate
- Outputs the fetched emails in YAML format

Example output:
```yaml
content: "this is content\r\n"
date: "2025-03-23T01:29:29Z"
flags: ""
from: ' <noreply.a@mailgen.example.com>'
size: 146
subject: test mail 1742693369 to a
to: ' <a@testcot>'
uid: 1
```

## How Emails Are Structured

The test emails follow a simple format:
- Subject lines contain a timestamp to make them unique
- Content is minimal ("this is content")
- The sender is always `noreply.username@mailgen.example.com`
- The recipient is `username@testcot`

## Docker Compose Configuration

The `docker-compose.yaml` file:
- Builds the custom Dovecot image from the `test-dovecot` directory
- Maps all necessary ports to the local interface (127.0.0.1)
- Creates persistent volumes for home directories and SSL certificates

This setup provides a consistent, reproducible environment for testing email-related applications without requiring an external mail server. 