package templates

// DefaultCSS is the embedded default stylesheet
const DefaultCSS = `/* LeafPress Default Styles */

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

.lp-body {
  font-family: var(--lp-font-body);
  font-size: 16px;
  line-height: 1.6;
  color: var(--lp-text);
  background-color: var(--lp-bg);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* Navigation */
.lp-nav {
  border-bottom: 1px solid var(--lp-border);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem 0;
}

.lp-nav-container {
  width: 100%;
  max-width: var(--lp-max-width);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding: 0 1rem;
}

.lp-nav-brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.lp-nav-title {
  font-family: var(--lp-font-heading);
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--lp-text);
  text-decoration: none;
}

.lp-nav-title:hover {
  color: var(--lp-accent);
}

.lp-nav-links {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  justify-content: center;
}

.lp-nav-link {
  color: var(--lp-text-muted);
  text-decoration: none;
  font-size: 0.9rem;
}

.lp-nav-link:hover {
  color: var(--lp-accent);
}

.lp-theme-toggle {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  color: var(--lp-text);
  transition: opacity 0.2s;
}

.lp-theme-toggle:hover {
  opacity: 0.7;
}

.lp-theme-icon {
  display: none;
  width: 20px;
  height: 20px;
}

.lp-theme-icon-light {
  display: block;
}

[data-theme="dark"] .lp-theme-icon-light {
  display: none;
}

[data-theme="dark"] .lp-theme-icon-dark {
  display: block;
}

/* Mobile navigation */
@media (max-width: 768px) {
  .lp-nav-container {
    gap: 0.5rem;
  }

  .lp-nav-brand {
    width: 100%;
    justify-content: space-between;
  }

  .lp-nav-links {
    width: 100%;
    justify-content: flex-start;
  }
}

/* Desktop navigation */
@media (min-width: 769px) {
  .lp-nav {
    padding: 0;
    min-height: var(--lp-nav-height);
  }

  .lp-nav-container {
    flex-direction: row;
    justify-content: space-between;
    padding: 0 2rem;
  }

  .lp-nav-links {
    gap: 1.5rem;
    justify-content: flex-end;
  }
}

/* Main content */
.lp-main {
  flex: 1;
  width: 100%;
  max-width: var(--lp-max-width);
  margin: 0 auto;
  padding: 2rem;
}

/* Page container with TOC */
.lp-page-container {
  display: flex;
  gap: 1.5rem;
  align-items: flex-start;
}

/* Table of Contents */
.lp-toc {
  display: none;
}

@media (min-width: 1280px) {
  .lp-toc {
    display: block;
    width: 220px;
    flex-shrink: 0;
    position: sticky;
    top: calc(var(--lp-nav-height) + 2rem);
    align-self: flex-start;
    max-height: calc(100vh - var(--lp-nav-height) - 4rem);
    overflow-y: auto;
  }

  .lp-main:has(.lp-toc) {
    max-width: 1200px;
  }
}

.lp-toc-nav {
  font-size: 0.875rem;
}

.lp-toc-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.lp-toc-item {
  margin-bottom: 0.1rem;
}

.lp-toc-level-2 {
  padding-left: 0;
}

.lp-toc-level-3 {
  padding-left: 1rem;
  font-size: 0.8rem;
}

.lp-toc-link {
  color: var(--lp-text-muted);
  text-decoration: none;
  display: block;
  padding: 0.1rem 0;
  transition: color 0.2s;
}

.lp-toc-link:hover {
  color: var(--lp-accent);
}

/* Article */
.lp-article {
  width: 100%;
  min-width: 0;
}

/* Scroll offset for anchor links (accounts for sticky nav) */
.lp-content h2[id],
.lp-content h3[id] {
  scroll-margin-top: calc(var(--lp-nav-height) + 2rem);
}

.lp-header {
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--lp-border);
}

.lp-title {
  font-family: var(--lp-font-heading);
  font-size: 2rem;
  font-weight: 700;
  line-height: 1.2;
  margin-bottom: 0.5rem;
}

.lp-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: var(--lp-text-muted);
  font-size: 0.9rem;
}

.lp-date {
  color: var(--lp-text-muted);
}

.lp-growth {
  font-size: 1rem;
}

.lp-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

.lp-tag {
  color: var(--lp-accent);
  text-decoration: none;
  font-size: 0.85rem;
}

.lp-tag:hover {
  text-decoration: underline;
}

/* Content */
.lp-content {
  line-height: 1.7;
}

.lp-content h1,
.lp-content h2,
.lp-content h3,
.lp-content h4,
.lp-content h5,
.lp-content h6 {
  font-family: var(--lp-font-heading);
  margin-top: 1rem;
  margin-bottom: 0.5rem;
  font-weight: 600;
  line-height: 1.3;
}

.lp-content h1 { font-size: 1.75rem; }
.lp-content h2 { font-size: 1.5rem; }
.lp-content h3 { font-size: 1.25rem; }
.lp-content h4 { font-size: 1.1rem; }

.lp-content p {
  margin-bottom: 1rem;
}

.lp-content a {
  color: var(--lp-accent);
  text-decoration: none;
}

.lp-content a:hover {
  text-decoration: underline;
}

.lp-content ul,
.lp-content ol {
  margin-bottom: 0.5rem;
  padding-left: 1.5rem;
}

.lp-content li {
  margin-bottom: 0.25rem;
}

.lp-content blockquote {
  border-left: 3px solid var(--lp-accent);
  padding-left: 1rem;
  margin: 1rem 0;
  color: var(--lp-text-muted);
  font-style: italic;
}

.lp-content pre {
  background-color: var(--lp-code-bg);
  border-radius: 4px;
  padding: 1rem;
  overflow-x: auto;
  margin: 1rem 0;
  position: relative;
  max-width: 100%;
}

.lp-content pre code {
  display: block;
  white-space: pre-wrap;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.lp-copy-button {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  background-color: var(--lp-bg);
  color: var(--lp-text);
  border: 1px solid var(--lp-border);
  border-radius: 3px;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}

.lp-content pre:hover .lp-copy-button {
  opacity: 1;
}

.lp-copy-button:hover {
  background-color: var(--lp-accent);
  color: white;
  border-color: var(--lp-accent);
}

.lp-copy-button:active {
  transform: scale(0.95);
}

.lp-content code {
  font-family: var(--lp-font-mono);
  font-size: 0.9em;
}

.lp-content p code,
.lp-content li code {
  background-color: var(--lp-code-bg);
  padding: 0.15em 0.4em;
  border-radius: 3px;
}

.lp-content img {
  max-width: 100%;
  height: auto;
  border-radius: 4px;
}

.lp-content hr {
  border: none;
  border-top: 1px solid var(--lp-border);
  margin: 2rem 0;
}

.lp-content table {
  width: 100%;
  border-collapse: collapse;
  margin: 1rem 0;
}

.lp-content th,
.lp-content td {
  border: 1px solid var(--lp-border);
  padding: 0.5rem;
  text-align: left;
}

.lp-content th {
  background-color: var(--lp-code-bg);
  font-weight: 600;
}

/* Wiki links */
.lp-wikilink {
  color: var(--lp-accent);
  text-decoration: none;
  border-bottom: 1px dashed var(--lp-accent);
}

.lp-wikilink:hover {
  border-bottom-style: solid;
}

.lp-broken-link {
  color: #dc3545;
  text-decoration: line-through;
}

.lp-external::after {
  content: "";
}

/* Knowledge Graph */
.lp-graph-container {
  margin-top: 3rem;
  padding-top: 2rem;
  border-top: 1px solid var(--lp-border);
}

.lp-graph-title {
  font-family: var(--lp-font-heading);
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--lp-text);
  margin-bottom: 1.5rem;
}

#lp-graph {
  width: 100%;
  height: 500px;
  border-radius: 6px;
  background: linear-gradient(to bottom, var(--lp-bg), var(--lp-code-bg));
  overflow: hidden;
  position: relative;
}

.lp-graph-node {
  transition: all 0.2s ease;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
}

.lp-graph-node:hover {
  filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.2));
}

.lp-graph-link {
  transition: all 0.2s ease;
}

.lp-graph-label {
  font-family: var(--lp-font-body);
  transition: opacity 0.2s ease;
}

[data-theme="dark"] .lp-graph-link {
  stroke: #444444 !important;
}

[data-theme="dark"] .lp-graph-node {
  filter: drop-shadow(0 2px 4px rgba(255, 255, 255, 0.1));
}

[data-theme="dark"] .lp-graph-node:hover {
  filter: drop-shadow(0 4px 8px rgba(255, 255, 255, 0.2));
}

/* Backlinks */
.lp-backlinks {
  margin-top: 3rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--lp-border);
}

.lp-backlinks-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--lp-text-muted);
  margin-bottom: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.lp-backlinks-list {
  list-style: none;
  padding: 0;
}

.lp-backlinks-list li {
  margin-bottom: 0.5rem;
}

.lp-backlink {
  color: var(--lp-accent);
  text-decoration: none;
}

.lp-backlink:hover {
  text-decoration: underline;
}

/* Section pages */
.lp-section {
  width: 100%;
}

.lp-section-title {
  font-family: var(--lp-font-heading);
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 1.5rem;
}

.lp-section-intro {
  margin-bottom: 2rem;
  color: var(--lp-text-muted);
}

.lp-index {
  list-style: none;
  padding: 0;
}

.lp-index-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--lp-border);
}

.lp-index-item:last-child {
  border-bottom: none;
}

.lp-index-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--lp-text);
  text-decoration: none;
}

.lp-index-link:hover .lp-index-title {
  color: var(--lp-accent);
}

.lp-index-title {
  font-weight: 500;
}

.lp-index-growth {
  font-size: 0.9rem;
}

.lp-index-date {
  color: var(--lp-text-muted);
  font-size: 0.85rem;
}

/* Tag cloud */
.lp-tag-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
}

.lp-tag-cloud-item {
  color: var(--lp-accent);
  text-decoration: none;
  font-size: 1rem;
}

.lp-tag-cloud-item:hover {
  text-decoration: underline;
}

.lp-tag-count {
  color: var(--lp-text-muted);
  font-size: 0.85rem;
}

/* Footer */
.lp-footer {
  border-top: 1px solid var(--lp-border);
  padding: 1.5rem 2rem;
  text-align: center;
  color: var(--lp-text-muted);
  font-size: 0.85rem;
}

.lp-footer a {
  color: var(--lp-accent);
  text-decoration: none;
}

.lp-footer a:hover {
  text-decoration: underline;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .lp-main {
    padding: 1.5rem 1rem;
  }

  .lp-title {
    font-size: 1.5rem;
  }

  .lp-content pre {
    padding: 0.75rem;
    border-radius: 0;
  }

  .lp-copy-button {
    opacity: 1;
  }
}

/* Syntax Highlighting (Chroma - GitHub theme) */
.chroma { background-color: var(--lp-code-bg); }
.chroma .err { color: #f6f8fa; background-color: #82071e }
.chroma .lnlinks { outline: none; text-decoration: none; color: inherit }
.chroma .lntd { vertical-align: top; padding: 0; margin: 0; border: 0; }
.chroma .lntable { border-spacing: 0; padding: 0; margin: 0; border: 0; }
.chroma .hl { background-color: #dedede }
.chroma .lnt { white-space: pre; -webkit-user-select: none; user-select: none; margin-right: 0.4em; padding: 0 0.4em 0 0.4em; color: #7f7f7f }
.chroma .ln { white-space: pre; -webkit-user-select: none; user-select: none; margin-right: 0.4em; padding: 0 0.4em 0 0.4em; color: #7f7f7f }
.chroma .line { display: flex; }
.chroma .k { color: #cf222e }
.chroma .kc { color: #cf222e }
.chroma .kd { color: #cf222e }
.chroma .kn { color: #cf222e }
.chroma .kp { color: #cf222e }
.chroma .kr { color: #cf222e }
.chroma .kt { color: #cf222e }
.chroma .na { color: #1f2328 }
.chroma .nc { color: #1f2328 }
.chroma .no { color: #0550ae }
.chroma .nd { color: #0550ae }
.chroma .ni { color: #6639ba }
.chroma .nl { color: #990000; font-weight: bold }
.chroma .nn { color: #24292e }
.chroma .nx { color: #1f2328 }
.chroma .nt { color: #0550ae }
.chroma .nb { color: #6639ba }
.chroma .bp { color: #6a737d }
.chroma .nv { color: #953800 }
.chroma .vc { color: #953800 }
.chroma .vg { color: #953800 }
.chroma .vi { color: #953800 }
.chroma .vm { color: #953800 }
.chroma .nf { color: #6639ba }
.chroma .fm { color: #6639ba }
.chroma .s { color: #0a3069 }
.chroma .sa { color: #0a3069 }
.chroma .sb { color: #0a3069 }
.chroma .sc { color: #0a3069 }
.chroma .dl { color: #0a3069 }
.chroma .sd { color: #0a3069 }
.chroma .s2 { color: #0a3069 }
.chroma .se { color: #0a3069 }
.chroma .sh { color: #0a3069 }
.chroma .si { color: #0a3069 }
.chroma .sx { color: #0a3069 }
.chroma .sr { color: #0a3069 }
.chroma .s1 { color: #0a3069 }
.chroma .ss { color: #032f62 }
.chroma .m { color: #0550ae }
.chroma .mb { color: #0550ae }
.chroma .mf { color: #0550ae }
.chroma .mh { color: #0550ae }
.chroma .mi { color: #0550ae }
.chroma .il { color: #0550ae }
.chroma .mo { color: #0550ae }
.chroma .o { color: #1f2328 }
.chroma .ow { color: #cf222e }
.chroma .p { color: #1f2328 }
.chroma .c { color: #6a737d; font-style: italic }
.chroma .ch { color: #6a737d; font-style: italic }
.chroma .cm { color: #6a737d; font-style: italic }
.chroma .c1 { color: #6a737d; font-style: italic }
.chroma .cs { color: #6a737d; font-weight: bold; font-style: italic }
.chroma .cp { color: #1f2328; font-weight: bold }
.chroma .cpf { color: #6a737d; font-style: italic }
.chroma .gd { color: #82071e; background-color: #ffebe9 }
.chroma .ge { font-style: italic }
.chroma .gr { color: #82071e }
.chroma .gh { color: #0550ae; font-weight: bold }
.chroma .gi { color: #116329; background-color: #dafbe1 }
.chroma .go { color: #1f2328 }
.chroma .gp { color: #6a737d }
.chroma .gs { font-weight: bold }
.chroma .gu { color: #0550ae }
.chroma .gt { color: #82071e }

/* Dark mode syntax highlighting */
[data-theme="dark"] .chroma { background-color: var(--lp-code-bg); }
[data-theme="dark"] .chroma .err { color: #960050; background-color: #1e0010 }
[data-theme="dark"] .chroma .hl { background-color: #3c3d38 }
[data-theme="dark"] .chroma .k { color: #66d9ef }
[data-theme="dark"] .chroma .kc { color: #66d9ef }
[data-theme="dark"] .chroma .kd { color: #66d9ef }
[data-theme="dark"] .chroma .kn { color: #f92672 }
[data-theme="dark"] .chroma .kp { color: #66d9ef }
[data-theme="dark"] .chroma .kr { color: #66d9ef }
[data-theme="dark"] .chroma .kt { color: #66d9ef }
[data-theme="dark"] .chroma .na { color: #a6e22e }
[data-theme="dark"] .chroma .nc { color: #a6e22e }
[data-theme="dark"] .chroma .no { color: #66d9ef }
[data-theme="dark"] .chroma .nd { color: #a6e22e }
[data-theme="dark"] .chroma .ne { color: #a6e22e }
[data-theme="dark"] .chroma .nx { color: #a6e22e }
[data-theme="dark"] .chroma .nt { color: #f92672 }
[data-theme="dark"] .chroma .nf { color: #a6e22e }
[data-theme="dark"] .chroma .fm { color: #a6e22e }
[data-theme="dark"] .chroma .s { color: #e6db74 }
[data-theme="dark"] .chroma .sa { color: #e6db74 }
[data-theme="dark"] .chroma .sb { color: #e6db74 }
[data-theme="dark"] .chroma .sc { color: #e6db74 }
[data-theme="dark"] .chroma .dl { color: #e6db74 }
[data-theme="dark"] .chroma .sd { color: #e6db74 }
[data-theme="dark"] .chroma .s2 { color: #e6db74 }
[data-theme="dark"] .chroma .se { color: #ae81ff }
[data-theme="dark"] .chroma .sh { color: #e6db74 }
[data-theme="dark"] .chroma .si { color: #e6db74 }
[data-theme="dark"] .chroma .sx { color: #e6db74 }
[data-theme="dark"] .chroma .sr { color: #e6db74 }
[data-theme="dark"] .chroma .s1 { color: #e6db74 }
[data-theme="dark"] .chroma .ss { color: #e6db74 }
[data-theme="dark"] .chroma .m { color: #ae81ff }
[data-theme="dark"] .chroma .mb { color: #ae81ff }
[data-theme="dark"] .chroma .mf { color: #ae81ff }
[data-theme="dark"] .chroma .mh { color: #ae81ff }
[data-theme="dark"] .chroma .mi { color: #ae81ff }
[data-theme="dark"] .chroma .il { color: #ae81ff }
[data-theme="dark"] .chroma .mo { color: #ae81ff }
[data-theme="dark"] .chroma .o { color: #f92672 }
[data-theme="dark"] .chroma .ow { color: #f92672 }
[data-theme="dark"] .chroma .p { color: #f8f8f2 }
[data-theme="dark"] .chroma .c { color: #75715e }
[data-theme="dark"] .chroma .ch { color: #75715e }
[data-theme="dark"] .chroma .cm { color: #75715e }
[data-theme="dark"] .chroma .c1 { color: #75715e }
[data-theme="dark"] .chroma .cs { color: #75715e }
[data-theme="dark"] .chroma .cp { color: #75715e }
[data-theme="dark"] .chroma .cpf { color: #75715e }
[data-theme="dark"] .chroma .gd { color: #f92672 }
[data-theme="dark"] .chroma .ge { font-style: italic }
[data-theme="dark"] .chroma .gi { color: #a6e22e }
[data-theme="dark"] .chroma .gs { font-weight: bold }
[data-theme="dark"] .chroma .gu { color: #75715e }
`
