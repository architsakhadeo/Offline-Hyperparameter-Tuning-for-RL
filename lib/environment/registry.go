package environment

import (
	"errors"
	"github.com/stellentus/cartpoles/lib/logger"
	"github.com/stellentus/cartpoles/lib/rlglue"
)

// NewEnvironmentCreator is a function that can create an environment.
type NewEnvironmentCreator func(logger.Debug) (rlglue.Environment, error)

// Add is used to register a new environment type.
// This is most likely called by an init function in the Environment's go file.
// The function returns an error if an environment with that name already exists.
func Add(name string, creator NewEnvironmentCreator) error {
	if _, ok := environmentList[name]; ok {
		return errors.New("Environment '" + name + "' has already been registered")
	}
	environmentList[name] = creator
	return nil
}

func Create(name string, debug logger.Debug) (rlglue.Environment, error) {
	creator, ok := environmentList[name]
	if !ok {
		return nil, errors.New("Environment '" + name + "' has not been registered")
	}
	return creator(debug)
}

var environmentList = map[string]NewEnvironmentCreator{}
