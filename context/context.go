package context

import (
	"database/sql"
	"github.com/bor3ham/reja/database"
	"github.com/bor3ham/reja/instances"
	gorillaContext "github.com/gorilla/context"
	"net/http"
	"sync"
)

type CachedInstance struct {
	Instance    instances.Instance
	RelationMap map[string]map[string][]string
}
type Context interface {
	QueryRow(string, ...interface{}) *sql.Row
	Query(string, ...interface{}) (*sql.Rows, error)

	InitCache()
	CacheObject(instances.Instance, map[string]map[string][]string)
	GetCachedObject(string, string) (instances.Instance, map[string]map[string][]string)
}

type RequestContext struct {
	Request *http.Request

	InstanceCache struct {
		sync.Mutex
		Instances map[string]map[string]CachedInstance
	}
}

func (rc *RequestContext) incrementQueryCount() {
	queries := rc.GetQueryCount()
	queries += 1
	gorillaContext.Set(rc.Request, "queries", queries)
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
	rc.incrementQueryCount()
	return database.QueryRow(query, args...)
}
func (rc *RequestContext) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rc.incrementQueryCount()
	return database.Query(query, args...)
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
