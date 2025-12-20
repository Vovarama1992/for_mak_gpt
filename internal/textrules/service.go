package textrules

import (
	"context"
	"strings"
	"unicode"
)

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) Process(ctx context.Context, text string) (string, error) {
	// 1) letters
	letterRules, err := s.repo.ListLetterRules(ctx)
	if err != nil {
		return "", err
	}

	if len(letterRules) > 0 {
		var b strings.Builder
		for _, r := range text {
			replaced := r
			for _, rule := range letterRules {
				if len([]rune(rule.From)) == 1 && r == []rune(rule.From)[0] {
					replaced = []rune(rule.To)[0]
					break
				}
			}
			b.WriteRune(replaced)
		}
		text = b.String()
	}

	// 2) words
	wordRules, err := s.repo.ListWordRules(ctx)
	if err != nil {
		return "", err
	}

	if len(wordRules) == 0 {
		return text, nil
	}

	tokens := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r)
	})

	for i, tok := range tokens {
		for _, rule := range wordRules {
			if tok == rule.From {
				tokens[i] = rule.To
				break
			}
		}
	}

	return strings.Join(tokens, " "), nil
}
