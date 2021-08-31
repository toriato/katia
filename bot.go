package katia

import (
	"path/filepath"
	"plugin"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	*Context

	Session *discordgo.Session
	Logger  *logrus.Logger

	plugins    map[string]*Plugin
	commands   map[string]*Command
	components map[string]*Component
}

func New(token string) (*Bot, error) {
	bot := &Bot{
		Context:    &Context{},
		Logger:     logrus.New(),
		plugins:    make(map[string]*Plugin),
		commands:   make(map[string]*Command),
		components: make(map[string]*Component),
	}

	{
		session, err := discordgo.New("Bot " + token)
		if err != nil {
			return nil, err
		}

		session.AddHandler(func(_ *discordgo.Session, e *discordgo.Ready) {
			handleReady(bot, e)
		})

		session.AddHandler(func(_ *discordgo.Session, e *discordgo.InteractionCreate) {
			handleInteractionCreate(bot, e)
		})

		bot.Session = session
	}

	// 플러그인 등록하기
	{
		paths, err := filepath.Glob("plugins/*.so")
		if err != nil {
			return nil, err
		}

		for _, path := range paths {
			pkg, err := plugin.Open(path)
			if err != nil {
				return nil, err
			}

			symbol, err := pkg.Lookup("Plugin")
			if err != nil {
				return nil, err
			}

			plugin, ok := symbol.(*Plugin)
			if !ok {
				return nil, ErrPluginBroken
			}

			if err := bot.RegisterPlugin(plugin); err != nil {
				return nil, err
			}
		}
	}

	return bot, nil
}

// Plugin 메소드는 이름으로 플러그인을 찾아 반환합니다
func (bot Bot) Plugin(name string) *Plugin {
	if plugin, ok := bot.plugins[name]; ok {
		return plugin
	}

	return nil
}

// Command 메소드는 이름으로 명령어 구조를 찾아 반환합니다
func (bot Bot) Command(name string) *Command {
	if command, ok := bot.commands[name]; ok {
		return command
	}

	return nil
}

// Component 메소드는 이름으로 명령어 구조를 찾아 반환합니다
func (bot Bot) Component(name string) *Component {
	if component, ok := bot.components[name]; ok {
		return component
	}

	return nil
}

// RegisterPlugin
func (bot *Bot) RegisterPlugin(plugin *Plugin) error {
	if bot.Plugin(plugin.Name) != nil {
		return ErrPluginConflict
	}

	if plugin.OnRegister != nil {
		if err := plugin.OnRegister(bot); err != nil {
			return err
		}
	}

	// 플러그인 속 컴포넌트 등록하기
	for _, component := range plugin.Components {
		component.Plugin = plugin
		if err := bot.RegisterComponent(component); err != nil {
			return err
		}
	}

	// 플러그인 속 명령어 등록하기
	for _, command := range plugin.Commands {
		command.Plugin = plugin
		if err := bot.RegisterCommand(command); err != nil {
			return err
		}
	}

	bot.Logger.Infof("Plugin %s v%+v loaded", plugin.Name, plugin.Version)

	bot.plugins[plugin.Name] = plugin
	return nil
}

// RegisterCommand
func (bot *Bot) RegisterCommand(command *Command) error {
	if bot.Command(command.Name) != nil {
		return ErrCommandConflict
	}

	bot.commands[command.Name] = command
	return nil
}

// RegisterComponent
func (bot *Bot) RegisterComponent(component *Component) error {
	if bot.Component(component.Name) != nil {
		return ErrComponentConflict
	}

	bot.components[component.Name] = component
	return nil
}
