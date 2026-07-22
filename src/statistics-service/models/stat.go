package models

import "time"

// Stat is one aggregated row per post (source_id) in a likes/comments/views
// ClickHouse table: the full list of user_ids that interacted with the post,
// plus when the first and last interaction happened.
type Stat struct {
	SourceID         int64
	UserIDs          []int64
	CreatedTimestamp time.Time
	LastModified     time.Time
}

type PostStatsSummary struct {
	PostID              int64 `json:"post_id"`
	ViewsCount          int   `json:"views_count"`
	UniqueViewsCount    int   `json:"unique_views_count"`
	LikesCount          int   `json:"likes_count"`
	UniqueLikesCount    int   `json:"unique_likes_count"`
	CommentsCount       int   `json:"comments_count"`
	UniqueCommentsCount int   `json:"unique_comments_count"`
}

// UniqueInt64 returns the distinct values of ids, preserving first-seen order.
func UniqueInt64(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	unique := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}
