# Liman MYS Render Engine

This repository contains Liman's render engine. Render engine's purpose is running Liman extensions in a sandbox environment and providing high performance connection.

### Features

- Run PHP code sandboxed
- SSH connection pooling for higher performance
- High performance SSH tunneling for unix socket
- File uploading to remote servers
- SSH, SFTP, SMB, SSH Tunnel, WinRM connection options

#### Generate documents

```
golds -gen -dir=docs -wdpkgs-listing=solo --compact -nouses -source-code-reading=rich .
```