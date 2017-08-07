package schema

type Manager interface {
	Create() Instance
	GetFilterForUser(User, int) ([]string, []interface{})
}

type ManagerStub struct {
}
func (stub ManagerStub) GetFilterForUser(user User, nextArg int) ([]string, []interface{}) {
	return []string{}, []interface{}{}
}
