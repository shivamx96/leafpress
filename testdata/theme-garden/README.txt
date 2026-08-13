Leafpress theme garden
======================

This garden is the stable content fixture for developing and reviewing bundled
themes. It deliberately exercises the major visual surfaces in one small site:

- root, explicit section, and generated section indexes
- articles with metadata, tags, growth states, and a table of contents
- prose, tables, task lists, blockquotes, code, footnotes, media, and Mermaid diagrams
- every canonical callout type
- resolved wikilinks, backlinks, search, and a non-trivial knowledge graph

The fixture intentionally has no style.css. A build therefore shows the
selected bundled theme without user overrides.

To preview it with a locally built leafpress binary:

    cd testdata/theme-garden
    leafpress serve

Generated _site/ output is ignored and must not be committed.
