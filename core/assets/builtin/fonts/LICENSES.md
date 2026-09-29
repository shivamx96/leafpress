# Bundled font licenses

The woff2 files in this directory are latin/latin-ext webfont subsets sourced
from Google Fonts. Families use variable-weight files where available; IBM
Plex Mono ships its regular and bold static weights. All families are licensed under the
SIL Open Font License 1.1 (https://openfontlicense.org), which permits
bundling and redistribution:

- **Bricolage Grotesque** — Copyright 2022 The Bricolage Grotesque Project Authors
  (https://github.com/ateliertriay/bricolage)
- **IBM Plex Mono** — Copyright IBM Corp.
- **Inter** — Copyright 2020 The Inter Project Authors
  (https://github.com/rsms/inter)
- **JetBrains Mono** — Copyright 2020 The JetBrains Mono Project Authors
  (https://github.com/JetBrains/JetBrainsMono)
- **Newsreader** — Copyright 2020 The Newsreader Project Authors
  (http://github.com/productiontype/Newsreader)
- **Source Serif 4** — Copyright Adobe / Google Inc.
- **Space Grotesk** — Copyright 2020 The Space Grotesk Project Authors

The fonts are used as-is; no glyphs were modified. The subsets and their
unicode-range declarations mirror Google Fonts' serving of these families.
The bundled set is the families the built-in themes use, deliberately
Latin-focused (latin + latin-ext subsets); other scripts fall back to the
system font stacks. The CLI downloads any other Google Fonts family on first
use.

The complete OFL license text for each family, including its copyright
notice, ships in this directory (OFL-*.txt) as registry assets and is
materialized into exported sites alongside the font files, as the OFL's
redistribution terms require.
