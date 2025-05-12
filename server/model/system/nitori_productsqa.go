package system

// QAListResponse Q&A一覧APIのルートレスポンス
type QAListResponse struct {
	QAList     []QAInfo       `json:"qa_list"`    // Q&Aリスト
	Pagination PaginationInfo `json:"pagination"` // ページネーション情報
}

// QAInfo 質問と回答のペア情報
type QAInfo struct {
	Question *QuestionInfo `json:"question"`         // 質問情報
	Answer   *AnswerInfo   `json:"answer,omitempty"` // 回答情報 (回答がない場合は null)
}

// QuestionInfo 質問情報 (nickname削除)
type QuestionInfo struct {
	ID                 int64  `json:"id"`                   // 質問ID
	QuestionText       string `json:"question_text"`        // 質問本文
	CreatedAtFormatted string `json:"created_at_formatted"` // 表示用投稿日時
}

// AnswerInfo 回答情報 (answerer_type削除)
type AnswerInfo struct {
	ID                 int64  `json:"id"`                   // 回答ID
	AnswererName       string `json:"answerer_name"`        // 回答者表示名
	AnswerText         string `json:"answer_text"`          // 回答本文
	HelpfulCount       int    `json:"helpful_count"`        // 参考になった数
	CreatedAtFormatted string `json:"created_at_formatted"` // 表示用回答日時
}

// PaginationInfo ページネーション情報 (total_countは質問総件数)
// type PaginationInfo struct {
// 	CurrentPage int `json:"current_page"`
// 	Limit       int `json:"limit"`
// 	TotalCount  int `json:"total_count"`
// 	TotalPages  int `json:"total_pages"`
// }

// ErrorResponse エラーレスポンス構造 (共通)
// type ErrorResponse struct {
// 	Error ErrorDetail `json:"error"`
// }
// type ErrorDetail struct {
// 	Code    string `json:"code"`
// 	Message string `json:"message"`
// 	Target  string `json:"target,omitempty"`
// }
//用于绑定qa查询参数
type QARequests struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
	Sort  int `form:"sort"`
}
