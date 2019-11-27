package packit

type Environment map[string]string

func NewEnvironment() Environment {
	return Environment{}
}

func (e Environment) Override(name, value string) {
	e[name+".override"] = value
}

func (e Environment) Prepend(name, value, delim string) {
	e[name+".prepend"] = value

	delete(e, name+".delim")
	if delim != "" {
		e[name+".delim"] = delim
	}
}

func (e Environment) Append(name, value, delim string) {
	e[name+".append"] = value

	delete(e, name+".delim")
	if delim != "" {
		e[name+".delim"] = delim
	}
}

func (e Environment) Default(name, value string) {
	e[name+".default"] = value
}
