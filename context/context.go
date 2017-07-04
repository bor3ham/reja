package context

import (
	"database/sql"
	"github.com/bor3ham/reja/database"
	"github.com/bor3ham/reja/instances"
	gorillaContext "github.com/gorilla/context"
	"net/http"
	"sync"
)

type Transaction struct {
	tx *sql.Tx
	c Context
}
func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	t.c.IncrementQueryCount()
	return t.tx.QueryRow(query, args...)
}
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	t.c.IncrementQueryCount()
	return t.tx.Query(query, args...)
}
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	t.c.IncrementQueryCount()
	return t.tx.Exec(query, args...)
}
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

type CachedInstance struct {
	Instance    instances.Instance
	RelationMap map[string]map[string][]string
}
type Context interface {
	IncrementQueryCount()
	QueryRow(string, ...interface{}) *sql.Row
	Query(string, ...interface{}) (*sql.Rows, error)
	Exec(string, ...interface{}) (sql.Result, error)
	Begin() (*Transaction, error)

	InitCache()
	CacheObject(instances.Instance, map[string]map[string][]string)
	GetCachedObject(string, string) (instances.Instance, map[string]map[string][]string)
}

type RequestContext struct {
	Request      *http.Request
	gorillaMutex sync.Mutex

	InstanceCache struct {
		sync.Mutex
		Instances map[string]map[string]CachedInstance
	}
}

func (rc *RequestContext) IncrementQueryCount() {
	rc.gorillaMutex.Lock()
	queries := rc.GetQueryCount()
	queries += 1
	gorillaContext.Set(rc.Request, "queries", queries)
	rc.gorillaMutex.Unlock()
}
func (rc *RequestContext) GetQueryCount() int {
	current := gorillaContext.Get(rc.Request, "queries")
	if current != nil {
		currentInt, ok := current.(int)
		if !ok {
			panic("Unable to convert query count to integer")
		}
		return currentInt
	}
	return 0
}
func (rc *RequestContext) QueryRow(query string, args ...interface{}) *sql.Row {
	rc.IncrementQueryCount()
	return database.QueryRow(query, args...)
}
func (rc *RequestContext) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rc.IncrementQueryCount()
	return database.Query(query, args...)
}
func (rc *RequestContext) Exec(query string, args ...interface{}) (sql.Result, error) {
	rc.IncrementQueryCount()
	return database.Exec(query, args...)
}
func (rc *RequestContext) Begin() (*Transaction, error) {
	tx, err := database.Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{
		c: rc,
		tx: tx,
	}, nil
}

func (rc *RequestContext) InitCache() {
	rc.InstanceCache.Lock()
	rc.InstanceCache.Instances = map[string]map[string]CachedInstance{}
	rc.InstanceCache.Unlock()
}
func (rc *RequestContext) CacheObject(object instances.Instance, relationMap map[string]map[string][]string) {
	rc.InstanceCache.Lock()
	model := object.GetType()
	id := object.GetID()
	_, exists := rc.InstanceCache.Instances[model]
	if !exists {
		rc.InstanceCache.Instances[model] = map[string]CachedInstance{}
	}
	rc.InstanceCache.Instances[model][id] = CachedInstance{
		Instance:    object,
		RelationMap: relationMap,
	}
	rc.InstanceCache.Unlock()
}
func (rc *RequestContext) GetCachedObject(instanceType string, instanceId string) (instances.Instance, map[string]map[string][]string) {
	rc.InstanceCache.Lock()
	defer rc.InstanceCache.Unlock()
	models, modelExists := rc.InstanceCache.Instances[instanceType]
	if modelExists {
		instance, exists := models[instanceId]
		if exists {
			return instance.Instance, instance.RelationMap
		}
	}
	return nil, nil
}
