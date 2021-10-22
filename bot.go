package katia

import (
	"os"
	"path/filepath"
	"plugin"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	mapset "github.com/deckarep/golang-set"
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
		Context: &Context{values: make(map[string]interface{})},
		Logger: &logrus.Logger{
			Out:   os.Stdout,
			Level: logrus.InfoLevel,
			Formatter: &logrus.JSONFormatter{
				DataKey:     "data",
				PrettyPrint: true,
			},
		},
		plugins:    map[string]*Plugin{},
		commands:   map[string]*Command{},
		components: map[string]*Component{},
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

	if plugin.Logger == nil {
		plugin.Logger = bot.Logger.WithField("plugin", plugin)
	}

	bot.plugins[plugin.Name] = plugin

	bot.Logger.
		WithField("plugin", plugin).
		Infof("plugin '%s' registered", plugin.Name)
	return nil
}

// RegisterCommand
func (bot *Bot) RegisterCommand(command *Command) error {
	if bot.Command(command.Name) != nil {
		return ErrCommandConflict
	}

	appID := bot.Session.State.User.ID
	app := &discordgo.ApplicationCommand{
		Name:        command.Name,
		Description: command.Description,
	}

	if command.Options != nil {
		app.Options = command.parseOptions(command.Options)
	}

	log := bot.Logger.WithField("command", command)

	if _, err := bot.Session.ApplicationCommandCreate(appID, "882969983418785802", app); err != nil {
		log.Error(err)
	}

	bot.commands[command.Name] = command

	log.Infof("command '%s' registered", command.Name)
	return nil
}

// RegisterComponent
func (bot *Bot) RegisterComponent(component *Component) error {
	if bot.Component(component.Name) != nil {
		return ErrComponentConflict
	}

	bot.components[component.Name] = component

	bot.Logger.
		WithField("component", component).
		Infof("component '%s' registered", component.Name)
	return nil
}

func (bot Bot) ResolvePluginGraph() ([]*Plugin, error) {
	sets := map[string]mapset.Set{}
	resolved := []*Plugin{}

	for name, plugin := range bot.plugins {
		set := mapset.NewSet()

		for _, dep := range plugin.Depends {
			set.Add(dep)
		}

		sets[name] = set
	}

	for len(sets) != 0 {
		ready := mapset.NewSet()

		for name, deps := range sets {
			if deps.Cardinality() == 0 {
				ready.Add(name)
			}
		}

		if ready.Cardinality() == 0 {
			g := []*Plugin{}

			for name := range sets {
				g = append(g, bot.plugins[name])
			}

			return g, ErrPluginDependCircular
		}

		for name := range ready.Iter() {
			delete(sets, name.(string))
			resolved = append(resolved, bot.plugins[name.(string)])
		}

		for name, deps := range sets {
			sets[name] = deps.Difference(ready)
		}
	}

	return resolved, nil
}
