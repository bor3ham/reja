package schema

type Manager interface {
	Create() Instance
	GetFilterForUser(User, int) ([]string, []interface{})

	BeforeCreate(Context, map[string]interface{}) error
	BeforeUpdate(Context, map[string]interface{}, map[string]interface{}) error
	BeforeSave(Context, map[string]interface{}) error
}

type ManagerStub struct {
}

func (stub ManagerStub) GetFilterForUser(user User, nextArg int) ([]string, []interface{}) {
	return []string{}, []interface{}{}
}

func (stub ManagerStub) BeforeCreate(c Context, values map[string]interface{}) error {
	return nil
}
func (stub ManagerStub) BeforeUpdate(
	c Context,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) (
	error,
) {
	return nil
}
func (stub ManagerStub) BeforeSave(c Context, values map[string]interface{}) error {
	return nil
}
