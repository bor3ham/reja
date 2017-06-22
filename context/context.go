package context

import (
	"sync"
	"database/sql"
	"github.com/bor3ham/reja/database"
	"github.com/bor3ham/reja/instances"
	gorillaContext "github.com/gorilla/context"
	"net/http"
)

type Context interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)

	InitCache()
	CacheObjects([]instances.Instance)
	GetCachedObject(instanceType string, instanceId string) instances.Instance
}

type RequestContext struct {
	Request *http.Request

	InstanceCache struct {
		sync.Mutex
		Instances map[string]map[string]instances.Instance
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
	rc.InstanceCache.Instances = map[string]map[string]instances.Instance{}
	rc.InstanceCache.Unlock()
}
func (rc *RequestContext) CacheObjects(objects []instances.Instance) {
	rc.InstanceCache.Lock()
	for _, instance := range objects {
		model := instance.GetType()
		id := instance.GetID()
		_, exists := rc.InstanceCache.Instances[model]
		if !exists {
			rc.InstanceCache.Instances[model] = map[string]instances.Instance{}
		}
		rc.InstanceCache.Instances[model][id] = instance
	}
	rc.InstanceCache.Unlock()
}
func (rc *RequestContext) GetCachedObject(instanceType string, instanceId string) instances.Instance {
	rc.InstanceCache.Lock()
	defer rc.InstanceCache.Unlock()
	models, modelExists := rc.InstanceCache.Instances[instanceType]
	if modelExists {
		instance, exists := models[instanceId]
		if exists {
			return instance
		}
	}
	return nil
}

