package katia

type Context struct {
	values map[string]interface{}
}

func (context Context) Get(key string) (value interface{}, ok bool) {
	value, ok = context.values[key]
	return value, ok
}

func (context Context) Set(key string, value interface{}) {
	context.values[key] = value
}
