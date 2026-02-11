# Plan: Remove "Unavailable" Check from QuotaSense

## Problem
The user wants to remove the check for `Unavailable` in the quota display logic. Currently, if an account is marked as `Unavailable`, it is skipped. This behavior is no longer desired.

## Proposed Changes

### 1. Modify `cmd/root.go`
- Locate the `displayQuota` function.
- Find the loop iterating over `files`.
- Change the condition `if file.Disabled || file.Unavailable` to `if file.Disabled`.

### 2. Modify `internal/models/models.go`
- Locate the `AuthFile` struct definition.
- Remove the `Unavailable bool` field.

## Verification
- Run `go build` to ensure the project compiles successfully.
- Run the tool to verify that accounts previously marked as unavailable (if any) are now processed, or at least that the build doesn't break.
