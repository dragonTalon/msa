# 规格：self-update

## Purpose

系统 SHALL 提供自更新能力，包括版本检查、下载、替换二进制文件以及一键安装脚本。

## Requirements

### Requirement: Check for updates

The system SHALL provide a `msa update --check` command that checks for the latest version from GitHub Releases API without performing an update.

#### Scenario: Newer version available
- **WHEN** user runs `msa update --check`
- **AND** a newer version exists on GitHub Releases
- **THEN** system displays current version, latest version, and prompts that an update is available

#### Scenario: Already on latest version
- **WHEN** user runs `msa update --check`
- **AND** current version matches the latest release
- **THEN** system displays "Already on latest version" message

#### Scenario: Network error during check
- **WHEN** user runs `msa update --check`
- **AND** GitHub API is unreachable
- **THEN** system displays an error message and exits with non-zero status

### Requirement: Perform self-update

The system SHALL provide a `msa update` command that downloads and replaces the current binary with the latest version from GitHub Releases.

#### Scenario: Successful update on Linux/macOS
- **WHEN** user runs `msa update`
- **AND** a newer version exists
- **THEN** system downloads the appropriate binary for current OS/arch
- **AND** system verifies checksum from checksums.txt
- **AND** system backs up current binary
- **AND** system replaces the binary
- **AND** system cleans up backup and temporary files
- **AND** system displays success message with new version

#### Scenario: Windows platform
- **WHEN** user runs `msa update` on Windows
- **THEN** system downloads the new version to a temporary location
- **AND** system displays instructions for manual replacement

#### Scenario: Checksum verification failure
- **WHEN** user runs `msa update`
- **AND** downloaded file checksum does not match
- **THEN** system aborts the update
- **AND** system displays checksum mismatch error
- **AND** system preserves the original binary

#### Scenario: Insufficient permissions
- **WHEN** user runs `msa update`
- **AND** user lacks write permission to binary location
- **THEN** system displays permission error with suggested fix (e.g., run with sudo)

### Requirement: One-click installation script

The system SHALL provide an `install.sh` script for one-click installation of the latest MSA version.

#### Scenario: Fresh installation
- **WHEN** user runs `curl -fsSL https://raw.githubusercontent.com/dragonTalon/msa/main/install.sh | sh`
- **THEN** script detects OS and architecture
- **AND** script downloads the latest release binary
- **AND** script installs binary to ~/.local/bin or /usr/local/bin
- **AND** script adds install location to PATH if needed
- **AND** script displays success message

#### Scenario: Installation directory not in PATH
- **WHEN** install script completes
- **AND** installation directory is not in PATH
- **THEN** script displays instructions to add directory to PATH

#### Scenario: Existing installation
- **WHEN** user runs install script
- **AND** MSA is already installed
- **THEN** script upgrades to the latest version
