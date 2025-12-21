package templates

// DefaultCSS is the embedded default stylesheet
const DefaultCSS = `/* LeafPress Default Styles */

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

.lp-body {
  font-family: var(--lp-font);
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
  height: var(--lp-nav-height);
  border-bottom: 1px solid var(--lp-border);
  display: flex;
  align-items: center;
  justify-content: center;
}

.lp-nav-container {
  width: 100%;
  max-width: var(--lp-max-width);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2rem;
}

.lp-nav-title {
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
  gap: 1.5rem;
}

.lp-nav-link {
  color: var(--lp-text-muted);
  text-decoration: none;
  font-size: 0.9rem;
}

.lp-nav-link:hover {
  color: var(--lp-accent);
}

/* Main content */
.lp-main {
  flex: 1;
  width: 100%;
  max-width: var(--lp-max-width);
  margin: 0 auto;
  padding: 2rem;
}

/* Article */
.lp-article {
  width: 100%;
}

.lp-header {
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--lp-border);
}

.lp-title {
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
  margin-top: 2rem;
  margin-bottom: 1rem;
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
  margin-bottom: 1rem;
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

/* Responsive */
@media (max-width: 768px) {
  .lp-nav-container {
    padding: 0 1rem;
  }

  .lp-main {
    padding: 1.5rem 1rem;
  }

  .lp-title {
    font-size: 1.5rem;
  }

  .lp-nav-links {
    gap: 1rem;
  }
}
`
