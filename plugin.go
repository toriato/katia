package katia

import (
	"errors"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type Plugin struct {
	Logger *logrus.Entry `json:"-"`

	Name        string   `json:"name"`
	Author      string   `json:"-"`
	Description string   `json:"-"`
	Version     [3]int   `json:"version"`
	Depends     []string `json:"-"`

	OnEnable  func(bot *Bot, plugin *Plugin) error `json:"-"`
	OnDisable func(bot *Bot, plugin *Plugin) error `json:"-"`
}

var (
	ErrPluginConflict       = errors.New("플러그인의 이름은 중복될 수 없습니다")
	ErrPluginBroken         = errors.New("플러그인의 구조가 잘못됐습니다")
	ErrPluginDependMissing  = errors.New("플러그인이 요구하는 종속 플러그인이 존재하지 않습니다")
	ErrPluginDependCircular = errors.New("플러그인이 요구하는 종속 플러그인이 순환됩니다")
)

func (plugin Plugin) Base() string {
	return filepath.Join("plugins", plugin.Name)
}
