package system

import "time"

// ReviewListResponse レビュー一覧APIのルートレスポンス
type ReviewListResponse struct {
	Summary    *ReviewSummary `json:"summary"`    // レビュー集計情報 (Nullable: 商品にレビューがない場合)
	Reviews    []ReviewInfo   `json:"reviews"`    // レビューリスト
	Pagination PaginationInfo `json:"pagination"` // ページネーション情報
}

// ReviewSummary レビュー集計情報
type ReviewSummary struct {
	AverageRating float64 `json:"average_rating"` // 平均評価
	ReviewCount   int     `json:"review_count"`   // 承認済みレビュー総数
	Rating1Count  int     `json:"rating_1_count"` // 星1の数
	Rating2Count  int     `json:"rating_2_count"` // 星2の数
	Rating3Count  int     `json:"rating_3_count"` // 星3の数
	Rating4Count  int     `json:"rating_4_count"` // 星4の数
	Rating5Count  int     `json:"rating_5_count"` // 星5の数
}

// ReviewInfo 個々のレビュー情報 (修正: image_urls, helpful_count 追加)
type ReviewInfo struct {
	ID                 int      `json:"id"`                   // レビューID
	Nickname           string   `json:"nickname"`             // ニックネーム
	Rating             int      `json:"rating"`               // 評価 (1-5)
	Title              *string  `json:"title,omitempty"`      // タイトル (Nullable)
	Comment            string   `json:"comment"`              // 本文
	CreatedAtFormatted string   `json:"created_at_formatted"` // 表示用投稿日時 (例: "2023年10月26日")
	ImageUrls          []string `json:"image_urls,omitempty"` // ★添付画像URLリスト (画像がない場合は空配列 or 省略)
	HelpfulCount       int      `json:"helpful_count"`        // ★参考になった数
	// IsHelpfulByUser   *bool    `json:"is_helpful_by_user,omitempty"` // ★(オプション) ログインユーザーが参考になったを押したか (Nullable)
}

// PaginationInfo ページネーション情報 (変更なし)
type PaginationInfo struct {
	CurrentPage int `json:"current_page"`
	Limit       int `json:"limit"`
	TotalCount  int `json:"total_count"`
	TotalPages  int `json:"total_pages"`
}

// ErrorResponse エラーレスポンス構造 (共通)
//与之前相同，应该不需要
// type ErrorResponse struct {
// 	Error ErrorDetail `json:"error"`
// }
// type ErrorDetail struct {
// 	Code    string `json:"code"`
// 	Message string `json:"message"`
// 	Target  string `json:"target,omitempty"`
// }

//以下是相关数据库构造体
type ProductReview struct {
	ID        int    `gorm:"primaryKey;column:id" json:"id"`      // 评论ID
	ProductID string `gorm:"column:product_id" json:"product_id"` // 商品ID
	// UserID   int64     `gorm:"column:user_id" json:"user_id"`     // 用户ID，如果有需要
	Nickname  string    `gorm:"column:nickname;default:匿名さん" json:"nickname"`                  // 显示昵称
	Rating    uint8     `gorm:"column:rating" json:"rating"`                                   // 评分 (1-5)
	Title     *string   `gorm:"column:title" json:"title,omitempty"`                           // 评论标题
	Comment   string    `gorm:"column:comment" json:"comment"`                                 // 评论内容
	Status    string    `gorm:"column:status;default:'pending'" json:"status"`                 // 评论状态
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
	// UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"` // 更新字段，如果有需要
}

func (ProductReview) TableName() string {
	return "product_reviews"
}

type ReviewSummaryDb struct {
	ProductID        string    `gorm:"primaryKey;column:product_id" json:"product_id"`           // 商品ID
	AverageRating    float64   `gorm:"column:average_rating;default:0.00" json:"average_rating"` // 平均评分
	ReviewCount      uint      `gorm:"column:review_count;default:0" json:"review_count"`        // 审核通过的评论数量
	Rating1Count     uint      `gorm:"column:rating_1_count;default:0" json:"rating_1_count"`    // 评分1的数量
	Rating2Count     uint      `gorm:"column:rating_2_count;default:0" json:"rating_2_count"`    // 评分2的数量
	Rating3Count     uint      `gorm:"column:rating_3_count;default:0" json:"rating_3_count"`    // 评分3的数量
	Rating4Count     uint      `gorm:"column:rating_4_count;default:0" json:"rating_4_count"`    // 评分4的数量
	Rating5Count     uint      `gorm:"column:rating_5_count;default:0" json:"rating_5_count"`    // 评分5的数量
	LastCalculatedAt time.Time `gorm:"column:last_calculated_at" json:"last_calculated_at"`      // 最后计算的时间
}

func (ReviewSummaryDb) TableName() string {
	return "review_summaries"
}

type ReviewImage struct {
	ID        int       `gorm:"primaryKey;column:id" json:"id"`                                // 评论图片ID
	ReviewID  int       `gorm:"column:review_id;not null" json:"review_id"`                    // 评论ID
	ImageURL  string    `gorm:"column:image_url;not null" json:"image_url"`                    // 图片URL
	SortOrder int       `gorm:"column:sort_order;default:0" json:"sort_order"`                 // 显示顺序
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
}

func (ReviewImage) TableName() string {
	return "review_images"
}

type UserReviewHelpfulVote struct {
	UserID    int       `gorm:"column:user_id;not null" json:"user_id"`                        // 投票的用户ID
	ReviewID  int       `gorm:"column:review_id;not null" json:"review_id"`                    // 投票的评论ID
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"` // 投票时间
}

func (UserReviewHelpfulVote) TableName() string {
	return "user_review_helpful_votes"
}

//这个用于接受评论页的参数
type ReviewsRequests struct {
	Page   int `form:"page"`
	Limit  int `form:"limit"`
	Sort   int `form:"sort"`
	Rating int `form:"rating"`
}
