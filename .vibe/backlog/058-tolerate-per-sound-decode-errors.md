---
status: done
---
# Tolerate Per-Sound Decode Errors

## Description
`Load`/`LoadBytes` abort entirely when any single `.snd` entry fails to decode, so an editor (`character-editor#017`) cannot show the sounds that work nor tell the user which one is broken. Decode each entry independently: a failing entry stays in its group, flagged with a descriptive error; every other entry loads normally.

## Acceptance Criteria
- [ ] A `.snd` (v1 and v2) with one undecodable entry still loads: the other sounds are decoded and exposed
- [ ] The failing `Sound` stays in its group, keyed by `(group, sample)`, with an `error` string naming group, sample and cause (empty PCM); `json:"error"` omitted/empty for good sounds
- [ ] A `.snd` that is entirely invalid (bad signature, unparsable table) is still a hard error naming the file
- [ ] Native `Load` and WASM `load` expose the same behavior; existing consumers of good sounds are unaffected

## Notes
Cross-repo: unblocks `character-editor#017` (sound browser), which needs the fixed `character` version published and pinned.
