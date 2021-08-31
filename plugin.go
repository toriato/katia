package katia

import "errors"

type Plugin struct {
	Name        string
	Author      string
	Description string
	Version     [3]int
	// Depends     []string
	// SoftDepends []string

	Components []*Component
	Commands   []*Command

	OnRegister func(bot *Bot) error
}

var (
	ErrPluginConflict       = errors.New("플러그인의 이름은 중복될 수 없습니다")
	ErrPluginBroken         = errors.New("플러그인의 구조가 잘못됐습니다")
	ErrPluginMissingDepends = errors.New("플러그인이 요구하는 종속 플러그인이 존재하지 않습니다")
)
