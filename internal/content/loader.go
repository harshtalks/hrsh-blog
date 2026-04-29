package content

import (
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Post struct {
	Title      string
	Date       string
	Body       string
	Slug       string
	ParsedDate time.Time
}

func (p Post) ReadingTime() string {
	words := len(strings.Fields(p.Body))
	minutes := words / 200
	if minutes < 1 {
		return "< 1 min"
	}
	return fmt.Sprintf("%d min", minutes)
}

var (
	jsxTagRe  = regexp.MustCompile(`</?[A-Z][A-Za-z0-9.]*[^>]*>`)
	importRe  = regexp.MustCompile(`(?m)^import\s+[^\n]+\n?`)
	exportRe  = regexp.MustCompile(`(?m)^export\s+[^\n]+\n?`)
)

func stripMDX(body string) string {
	body = importRe.ReplaceAllString(body, "")
	body = exportRe.ReplaceAllString(body, "")
	body = jsxTagRe.ReplaceAllString(body, "")
	return strings.TrimSpace(body)
}

func LoadPosts(fsys embed.FS, dir string) ([]Post, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}

	var posts []Post
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || (!strings.HasSuffix(name, ".md") && !strings.HasSuffix(name, ".mdx")) {
			continue
		}
		data, err := fsys.ReadFile(dir + "/" + name)
		if err != nil {
			continue
		}
		post, body, published := parseFrontmatter(string(data))
		if !published {
			continue
		}
		post.Slug = strings.TrimSuffix(strings.TrimSuffix(name, ".mdx"), ".md")
		post.Body = stripMDX(body)
		if post.Date != "" {
			if t, err := time.Parse("2006-01-02", post.Date); err == nil {
				post.ParsedDate = t
			}
		}
		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ParsedDate.After(posts[j].ParsedDate)
	})

	return posts, nil
}

func parseFrontmatter(raw string) (Post, string, bool) {
	if !strings.HasPrefix(raw, "---") {
		return Post{Title: "Untitled"}, raw, true
	}
	rest := raw[3:]
	idx := strings.Index(rest, "---")
	if idx == -1 {
		return Post{Title: "Untitled"}, raw, true
	}
	fm := rest[:idx]
	body := strings.TrimSpace(rest[idx+3:])

	post := Post{}
	published := true // default to visible if field absent

	for _, line := range strings.Split(fm, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "title:"); ok {
			post.Title = strings.Trim(strings.TrimSpace(after), `"'`)
		}
		if after, ok := strings.CutPrefix(line, "pubDate:"); ok {
			raw := strings.Trim(strings.TrimSpace(after), `"'`)
			// handle both "2024-10-21" and "2024-10-21T00:00:00.000Z"
			if len(raw) > 10 {
				raw = raw[:10]
			}
			post.Date = raw
		}
		if after, ok := strings.CutPrefix(line, "published:"); ok {
			val := strings.TrimSpace(after)
			published = val == "true"
		}
	}

	if post.Title == "" {
		post.Title = "Untitled"
	}
	return post, body, published
}
