package schema

type Manager interface {
	Create() Instance
}

type ManagerStub struct {
}
func (stub ManagerStub) GetFilterForUser(user User, nextArg int) (string, []interface{}) {
	return "", []interface{}{}
}
