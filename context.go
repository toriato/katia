package katia

type Context struct {
	values map[string]interface{}
}

func (context Context) Get(key string) (value interface{}) {
	return context.values[key]
}

func (context Context) Set(key string, value interface{}) {
	context.values[key] = value
}
