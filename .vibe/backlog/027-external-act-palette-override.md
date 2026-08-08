---
status: in_progress
depends_on: [026]
---
# External .act Palette Override

## Description
Add a function parsing a 768-byte `.act` buffer (256×RGB) into a 256-entry RGBA palette, reproducing `convertExternalPaletteToRGBA.mjs`'s two quirks exactly: the byte order is reversed (the file's first RGB triplet becomes palette index 255, its last becomes index 0), and only the resulting index 0 gets forced alpha 0 (all others opaque). Wire an optional override palette through item 026's resolution path so a caller can supply an external palette instead of the sprite's own.

## Acceptance Criteria
- [ ] Parsing a 768-byte `.act` buffer produces a 256-entry RGBA palette with the reversed ordering and index-0 transparency described above
- [ ] A non-768-byte buffer returns a descriptive error instead of panicking or silently truncating
- [ ] The v1/v2 palette resolution path (item 026) accepts an optional override palette and uses it in place of the sprite's own when supplied

## Notes
Reference: `convertExternalPaletteToRGBA.mjs` in `ikemen-launcher/sff-extractor`.
