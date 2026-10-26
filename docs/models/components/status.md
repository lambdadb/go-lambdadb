# Status

Legacy status constants retained for SDK compatibility. The current
`CollectionResponse` has no status property; do not poll these values for
Collection readiness. Verify the intended data through reads instead.

## Values

| Name             | Value            |
| ---------------- | ---------------- |
| `StatusCreating` | CREATING         |
| `StatusActive`   | ACTIVE           |
| `StatusDeleting` | DELETING         |
