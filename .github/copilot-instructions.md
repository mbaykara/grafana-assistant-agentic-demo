# Review rules for this repository

When reviewing a pull request:

- Every new or changed Go test function must have a doc comment directly above
  it that starts with the test name and says which bug it guards against, for
  example `// TestTotal_X guards against <bug>.` If a test function has no such
  comment, leave an inline review comment on that function asking for one.
- Fixes in internal/pricing must keep returning ErrUnknownCode for unknown
  discount codes. Comment if that behaviour changes.
- Ignore files under demo/: they are presenter material, not application code.
