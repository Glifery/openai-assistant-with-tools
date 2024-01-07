package types

type FunctionCalling interface {
	GetFunctionName() string
	GetFunctionDescription() string
	GetFunctionProperties() map[string]map[string]interface{}
	GetFunctionRequiredProperties() []string
	Execute(input string) (output any, err error)
}
