package textrules

import "context"

type LetterRule struct {
	From string // 1 rune
	To   string // 1 rune
}

type WordRule struct {
	From string
	To   string
}

type Repo interface {
	ListLetterRules(ctx context.Context) ([]LetterRule, error)
	ListWordRules(ctx context.Context) ([]WordRule, error)

	AddLetterRule(ctx context.Context, from, to string) error
	AddWordRule(ctx context.Context, from, to string) error

	DeleteLetterRule(ctx context.Context, from string) error
	DeleteWordRule(ctx context.Context, from string) error
}

type Service interface {
	Process(ctx context.Context, text string) (string, error)
}
