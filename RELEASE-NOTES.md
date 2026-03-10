# godynamo release notes

## 2026-03-10 - v2.0.0 (Grafana fork)

New module path: `github.com/grafana/godynamo/v2` (forked from `github.com/btnguyen2k/godynamo` v1.3.0).

### Changed

- BREAKING: Removed `RegisterAWSConfig` and `DeregisterAWSConfig` (global state)
- BREAKING: `Driver.Open` no longer reads global AWS config; it uses only DSN credentials
- Removed package-level `awsConfig` variable and associated mutex

### Added/Refactoring

- New `Connector` type implementing `database/sql/driver.Connector` for per-instance AWS configuration
- New `NewConnector(dsn, *aws.Config)` constructor for use with `sql.OpenDB`
- Extracted shared connection logic into internal `openConn` helper

### Migration from v1.3.0

Replace:
```go
godynamo.RegisterAWSConfig(cfg)
db, err := sql.Open("godynamo", dsn)
```

With:
```go
connector := godynamo.NewConnector(dsn, &cfg)
db := sql.OpenDB(connector)
```

If using DSN credentials only (no `aws.Config`), `sql.Open("godynamo", dsn)` continues to work unchanged.

## 2024-05-02 - v1.3.0

### Added/Refactoring

- New feature: sql.Open with aws.Config

## 2024-03-21 - v1.2.1

### Fixed/Improvement

- Fix #146: Columns are not sorted in the order of selection.

## 2024-01-06 - v1.2.0

### Added/Refactoring

- Refactor transaction support

### Fixed/Improvement

- Fix: empty transaction should be committed successfully

## 2024-01-02 - v1.1.1

### Fixed/Improvement

- Fix: result returned from a SELECT can be paged if too big

## 2023-12-31 - v1.1.0

### Added/Refactoring

- Add function WaitForTableStatus
- Add function WaitForGSIStatus
- Add method TransformInsertStmToPartiQL

### Fixed/Improvement

- Fix: empty LSI should be nil

## 2023-12-27 - v1.0.0

### Changed

- BREAKING: bump Go version to 1.18

### Added/Refactoring

- Refactor to follow go-module-template structure

### Fixed/Improvement

- Fix GoLint
- Fix CodeQL alerts

## 2023-07-27 - v0.4.0

- Support `ConsistentRead` option for `SELECT` query.

## 2023-07-25 - v0.3.1

- Fix: placeholder parsing.

## 2023-07-24 - v0.3.0

- `ColumnTypeDatabaseTypeName` returns DynamoDB's native data types (e.g. `B`, `N`, `S`, `SS`, `NS`, `BS`, `BOOL`, `L`, `M`, `NULL`).
- `RowsDescribeTable.ColumnTypeScanType` and `RowsDescribeIndex.ColumnTypeScanType` return correct Go types based on DynamoDB spec.
- Support `LIMIT` clause for `SELECT` query.

## 2023-05-31 - v0.2.0

- Add transaction support.

## 2023-05-27 - v0.1.0

- Driver for `database/sql`, supported statements:
  - Table: `CREATE TABLE`, `LIST TABLES`, `DESCRIBE TABLE`, `ALTER TABLE`, `DROP TABLE`.
  - Index: `DESCRIBE LSI`, `CREATE GSI`, `DESCRIBE GSI`, `ALTER GSI`, `DROP GSI`.
  - Document: `INSERT`, `SELECT`, `UPDATE`, `DELETE`.
