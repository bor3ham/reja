package context

import (
	"database/sql"
	"github.com/bor3ham/reja/instances"
	gorillaContext "github.com/gorilla/context"
	"net/http"
	"sync"
)

type RequestContext struct {
	Server Server
	Request      *http.Request
	gorillaMutex sync.Mutex

	InstanceCache struct {
		sync.Mutex
		Instances map[string]map[string]CachedInstance
	}
}

func (rc *RequestContext) GetServer() Server {
	return rc.Server
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
	return rc.Server.GetDatabase().QueryRow(query, args...)
}
func (rc *RequestContext) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rc.IncrementQueryCount()
	return rc.Server.GetDatabase().Query(query, args...)
}
func (rc *RequestContext) Exec(query string, args ...interface{}) (sql.Result, error) {
	rc.IncrementQueryCount()
	return rc.Server.GetDatabase().Exec(query, args...)
}
func (rc *RequestContext) Begin() (*Transaction, error) {
	tx, err := rc.Server.GetDatabase().Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{
		c:  rc,
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
