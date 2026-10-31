---
status: todo
---
# Preserve comments and ordering when a saved file was modified

## Description
README lists this as planned: Document round trips are byte-exact only for unmodified files. Regenerate output around edits (changed controller, renamed animation, added state) while keeping untouched comments, blank lines and section order, for .def/.air/.cns/.cmd/.zss.

## Acceptance Criteria
- [ ] Editing one value changes only its own lines in the saved output
- [ ] Comments and ordering around untouched sections are preserved for each of the five formats
- [ ] Added and removed sections land in a documented, deterministic position
- [ ] Property-style test: parse -> no-op edit -> serialize is identical for the whole real-file corpus
