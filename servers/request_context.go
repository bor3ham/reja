package servers

import (
	"database/sql"
	"encoding/json"
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/utils"
	"github.com/gorilla/context"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jwriter"
	"log"
	"net/http"
	"sync"
	"time"
)

type CachedInstance struct {
	Instance    schema.Instance
	RelationMap map[string]map[string][]string
}

type RequestContext struct {
	Server         schema.Server
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	user           schema.User
	gorillaMutex   sync.Mutex
	began          time.Time

	InstanceCache struct {
		sync.Mutex
		Instances map[string]map[string]CachedInstance
	}
}

func NewRequestContext(s schema.Server, w http.ResponseWriter, r *http.Request) *RequestContext {
	rc := RequestContext{
		Server:         s,
		Request:        r,
		ResponseWriter: w,
		began:          time.Now(),
	}
	rc.InitCache()
	return &rc
}

func (rc *RequestContext) GetRequest() *http.Request {
	return rc.Request
}
func (rc *RequestContext) GetServer() schema.Server {
	return rc.Server
}
func (rc *RequestContext) Authenticate() error {
	user, err := rc.Server.Authenticate(rc.ResponseWriter, rc.Request)
	if err != nil {
		authError, ok := err.(utils.AuthError)
		if ok {
			rc.ResponseWriter.WriteHeader(authError.Status)
			rc.WriteToResponse(schema.ErrorSet{
				Errors: []map[string]interface{}{
					map[string]interface{}{
						"status": authError.Status,
						"title":  err.Error(),
					},
				},
			})
		} else {
			rc.ResponseWriter.WriteHeader(http.StatusUnauthorized)
			rc.WriteToResponse(schema.ErrorSet{
				Errors: []map[string]interface{}{
					map[string]interface{}{
						"status": http.StatusUnauthorized,
						"title":  err.Error(),
					},
				},
			})
		}
		return err
	}
	rc.SetUser(user)
	return nil
}
func (rc *RequestContext) WriteToResponse(blob interface{}) {
	var err error
	easyBlob, hasEasyJson := blob.(interface {
		MarshalEasyJSON(*jwriter.Writer)
	})
	if rc.Server.UseEasyJSON() && hasEasyJson {
		_, _, err = easyjson.MarshalToHTTPResponseWriter(easyBlob, rc.ResponseWriter)
	} else {
		encoder := json.NewEncoder(rc.ResponseWriter)
		if rc.Server.Whitespace() {
			encoder.SetIndent("", "    ")
		}
		encoder.SetEscapeHTML(false)
		err = encoder.Encode(blob)
	}
	if err != nil {
		panic(err)
	}
}

func (rc *RequestContext) GetUser() schema.User {
	return rc.user
}
func (rc *RequestContext) SetUser(user schema.User) {
	rc.user = user
}

func (rc *RequestContext) IncrementQueryCount() {
	rc.gorillaMutex.Lock()
	queries := rc.GetQueryCount()
	queries += 1
	context.Set(rc.Request, "queries", queries)
	rc.gorillaMutex.Unlock()
}
func (rc *RequestContext) GetQueryCount() int {
	current := context.Get(rc.Request, "queries")
	if current != nil {
		currentInt, ok := current.(int)
		if !ok {
			panic("Unable to convert query count to integer")
		}
		return currentInt
	}
	return 0
}

func (rc *RequestContext) LogQuery(query string) {
	if rc.GetServer().LogSQL() {
		log.Println(query)
	}
}
func (rc *RequestContext) LogStats() {
	log.Println("\t", rc.Request.Method, rc.Request.URL.String())
	log.Println("Database queries:", rc.GetQueryCount())
	log.Println("Request duration:", time.Since(rc.began))
	log.Println()
}

func (rc *RequestContext) QueryRow(query string, args ...interface{}) *sql.Row {
	rc.LogQuery(query)
	rc.IncrementQueryCount()
	return rc.Server.GetDatabase().QueryRow(query, args...)
}
func (rc *RequestContext) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rc.LogQuery(query)
	rc.IncrementQueryCount()
	return rc.Server.GetDatabase().Query(query, args...)
}
func (rc *RequestContext) Exec(query string, args ...interface{}) (sql.Result, error) {
	rc.LogQuery(query)
	rc.IncrementQueryCount()
	return rc.Server.GetDatabase().Exec(query, args...)
}
func (rc *RequestContext) Begin() (schema.Transaction, error) {
	tx, err := rc.Server.GetDatabase().Begin()
	if err != nil {
		return nil, err
	}
	return &ContextTransaction{
		rc: rc,
		tx: tx,
	}, nil
}

func (rc *RequestContext) InitCache() {
	rc.InstanceCache.Lock()
	rc.InstanceCache.Instances = map[string]map[string]CachedInstance{}
	rc.InstanceCache.Unlock()
}
func (rc *RequestContext) FlushCache() {
	rc.InitCache()
}
func (rc *RequestContext) CacheObject(object schema.Instance, relationMap map[string]map[string][]string) {
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
func (rc *RequestContext) GetCachedObject(instanceType string, instanceId string) (schema.Instance, map[string]map[string][]string) {
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
