---
status: todo
---
# Write Trigger Groups

## Description
`Serialize` writes a controller's triggers from the flat `Triggers` list, renumbered `trigger1..N`. A `triggerall` therefore comes back as an OR group, and `SerializeEdited` ignores edits made to `TriggerAll` and `TriggerGroups` (item 080). Write `triggerall` and the numbered groups as parsed, so that a round trip keeps their meaning.

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO compatibility (roadmap decision `038`).

## Acceptance Criteria
- [ ] Parse → Serialize → Parse keeps `TriggerAll` and `TriggerGroups` identical on real files
- [ ] An edit of a group or a triggerall line is written back, byte-exact elsewhere

## Notes
Found by item 080 (2026-11-19).
- [ ] Fresh `SerializeCns` / `SerializeCmd` and a regenerated controller in `SerializeEdited` keep `triggerall` and multi-line groups (audit 2026-11-21 reproduced `triggerall=A / trigger1=B / trigger1=C` written as `trigger1=A, trigger2=B, trigger3=C`, an AND turned into an OR); `ctrlEqual` compares the groups
