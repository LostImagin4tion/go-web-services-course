package repositories

import (
	"net/url"
	"regexp"
	"stepikGoWebServices/handlers/entities"
	"strings"
	"sync"
	"time"
)

type ArticlesDbRepository struct {
	articles     map[string]*entities.Article
	articleMutex *sync.RWMutex
}

func NewArticlesDbRepository() *ArticlesDbRepository {
	return &ArticlesDbRepository{
		articles:     make(map[string]*entities.Article),
		articleMutex: &sync.RWMutex{},
	}
}

func (db *ArticlesDbRepository) AddArticle(
	title string,
	description string,
	body string,
	tagList []string,
	author entities.User,
) *entities.Article {
	var now = time.Now()
	var slug = db.generateSlug(title)

	var article = &entities.Article{
		Slug:        slug,
		Title:       title,
		Description: description,
		Body:        body,
		TagList:     tagList,
		CreatedAt:   now,
		UpdatedAt:   now,
		Author: entities.User{
			Bio:      author.Bio,
			Username: author.Username,
		},
	}

	db.articleMutex.Lock()
	defer db.articleMutex.Unlock()

	db.articles[slug] = article
	return article
}

func (db *ArticlesDbRepository) GetArticles(
	author string,
	tag string,
) []entities.Article {
	db.articleMutex.RLock()
	defer db.articleMutex.RUnlock()

	var articles = make([]entities.Article, 0)

	for _, article := range db.articles {
		if len(author) != 0 && article.Author.Username != author {
			continue
		}
		if len(tag) != 0 {
			found := false
			for _, t := range article.TagList {
				if t == tag {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		articles = append(articles, *article)
	}
	return articles
}

func (db *ArticlesDbRepository) generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return url.QueryEscape(slug)
}
