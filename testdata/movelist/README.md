# Move list fixtures

Hand-written, minimal files (not taken from a game): the Ikemen GO repository
at commit 54d4148a728153c2097a0fb899a06812b2d3e193 ships no character with a
move list (its `data/` folder holds only the engine's common files), so none
could be reused.

- `movelist.txt`: UTF-8 with a byte order mark, tab-separated columns, `<`-free
  glyph tokens such as `_QCF` and `^A`, and CRLF line ends, the shape Ikemen
  GO's pause menu parses (`external/script/menu.lua`, `menu.f_commandlistParse`).
