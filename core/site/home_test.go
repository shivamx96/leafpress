package site

import (
	"testing"

	"github.com/shivamx96/leafpress/core/content"
)

func TestApplyHomeTitleFillsOnlyUntitledHome(t *testing.T) {
	home := &content.Page{Slug: ""}
	titledHome := &content.Page{Slug: "", Title: "Welcome"}
	untitledPage := &content.Page{Slug: "notes/untitled"}
	ApplyHomeTitle([]*content.Page{home, titledHome, untitledPage, nil}, "Q&A Garden")
	if home.Title != "Q&A Garden" {
		t.Errorf("home title = %q, want site title", home.Title)
	}
	if titledHome.Title != "Welcome" {
		t.Errorf("frontmatter title replaced: %q", titledHome.Title)
	}
	if untitledPage.Title != "" {
		t.Errorf("non-home page received the site title: %q", untitledPage.Title)
	}
}
