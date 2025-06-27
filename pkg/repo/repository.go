package repo

type WhereCondition map[string]interface{}
type Respository interface {
	deviceRepo
}
