## ADDED Requirements

### Requirement: Test file organization
The system SHALL organize test files following Go standard conventions with `*_test.go` files in the same package directory as the code under test.

#### Scenario: Test file location
- **WHEN** adding tests for a package
- **THEN** test files SHALL be placed in the same directory as source files
- **THEN** test files SHALL be named `<source>_test.go`

#### Scenario: Test file structure
- **WHEN** test files are created
- **THEN** they SHALL be in the same package as the code being tested
- **THEN** they MAY use `package <name>_test` for black-box testing

### Requirement: Core utilities test coverage
The system SHALL include unit tests for core utility functions in `pkg/utils/`.

#### Scenario: Format functions
- **WHEN** testing format utilities
- **THEN** tests SHALL cover PrettyJSON, CompactJSON, ValidateJSON
- **THEN** tests SHALL include valid and invalid input scenarios

#### Scenario: String utilities
- **WHEN** testing string utilities
- **THEN** tests SHALL cover TruncateString, FormatArray
- **THEN** edge cases (empty, nil, overflow) SHALL be tested

### Requirement: Command system test coverage
The system SHALL include unit tests for the command registration and lookup system.

#### Scenario: Command registration
- **WHEN** a command is registered
- **THEN** it SHALL be retrievable via GetCommand
- **THEN** duplicate registrations SHALL be handled

#### Scenario: Command prefix matching
- **WHEN** using GetLikeCommand
- **THEN** matching commands SHALL be returned
- **THEN** case sensitivity SHALL be handled
- **THEN** "/" prefix SHALL be stripped

### Requirement: Configuration test coverage
The system SHALL include unit tests for configuration management.

#### Scenario: Config file operations
- **WHEN** testing configuration
- **THEN** file read/write operations SHALL be tested
- **THEN** default config creation SHALL be tested
- **THEN** JSON parsing errors SHALL be handled

### Requirement: Test coverage threshold
The system SHALL maintain at least 60% code coverage across all tested modules.

#### Scenario: Coverage measurement
- **WHEN** running `go test -cover`
- **THEN** coverage SHALL be at least 60%
- **THEN** coverage report SHALL be available

### Requirement: Test isolation
Each test SHALL run independently without affecting other tests.

#### Scenario: State cleanup
- **WHEN** tests modify shared state (singletons, globals)
- **THEN** state SHALL be reset before/after each test
- **THEN** tests SHALL not depend on execution order

### Requirement: Mock support
The system SHALL provide mocks for external dependencies.

#### Scenario: HTTP mocking
- **WHEN** testing code that makes HTTP requests
- **THEN** HTTP clients SHALL be mockable
- **THEN** responses SHALL be controllable in tests

#### Scenario: File system mocking
- **WHEN** testing code that reads/writes files
- **THEN** temporary directories SHALL be used for test files
- **THEN** test files SHALL be cleaned up after tests
