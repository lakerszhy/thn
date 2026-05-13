package hn

import (
	"context"
	"fmt"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"github.com/lakerszhy/thn/domain"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
)

type Client struct {
	db *db.Client

	mu    sync.RWMutex
	cache map[domain.Category][]int64

	semaphore chan struct{}
}

func New(ctx context.Context) (*Client, error) {
	config := &firebase.Config{
		DatabaseURL: "https://hacker-news.firebaseio.com/v0/",
	}

	app, err := firebase.NewApp(ctx, config, option.WithoutAuthentication())
	if err != nil {
		return nil, err
	}

	client, err := app.Database(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		db:        client,
		cache:     make(map[domain.Category][]int64),
		semaphore: make(chan struct{}, 10),
	}, nil
}

func (c *Client) FetchItems(ctx context.Context, cat domain.Category, p domain.Pagination) ([]domain.Item, error) {
	c.mu.RLock()
	ids, exisit := c.cache[cat]
	c.mu.RUnlock()

	if !exisit {
		var err error
		ids, err = c.fetchIDs(ctx, categoryFromDomain(cat))
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		c.cache[cat] = ids
		c.mu.Unlock()
	}

	start, end := p.Range(len(ids))
	pageIDs := ids[start:end]

	items := make([]domain.Item, len(pageIDs))

	g, ctx := errgroup.WithContext(ctx)
	for i, id := range pageIDs {
		g.Go(func() error {
			item, err := c.fetchItem(ctx, id)
			if err != nil {
				return err
			}
			items[i] = item.ToDomain()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return items, nil
}

func (c *Client) FetchComments(ctx context.Context, item domain.Item) ([]domain.Item, error) {
	comments := make([]domain.Item, len(item.KIDs))

	g, ctx := errgroup.WithContext(ctx)
	for i, id := range item.KIDs {
		g.Go(func() error {
			comment, err := c.fetchItem(ctx, id)
			if err != nil {
				return err
			}
			comments[i] = comment.ToDomain()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (c *Client) fetchIDs(ctx context.Context, cat category) ([]int64, error) {
	if err := c.acquire(ctx); err != nil {
		return nil, err
	}
	defer c.release()

	var ids []int64

	err := c.db.NewRef(string(cat)).Get(ctx, &ids)
	if err != nil {
		return nil, fmt.Errorf("fetch ids: %w", err)
	}

	return ids, nil
}

func (c *Client) fetchItem(ctx context.Context, id int64) (item, error) {
	if err := c.acquire(ctx); err != nil {
		return item{}, err
	}
	defer c.release()

	var i item
	err := c.db.NewRef(fmt.Sprintf("item/%d", id)).Get(ctx, &i)
	if err != nil {
		return item{}, fmt.Errorf("fetch item: %w", err)
	}

	return i, nil
}

func (c *Client) acquire(ctx context.Context) error {
	select {
	case c.semaphore <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) release() {
	<-c.semaphore
}
