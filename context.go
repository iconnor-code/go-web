package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	Req            *http.Request
	Resp           http.ResponseWriter
	PathParams     map[string]string
	queryValues    url.Values
	RespData       []byte
	RespStatusCode int
	tplEngine      TemplateEngine
	MatchedRoute   string
}

func (c *Context) Render(tplName string, data any) error {
	var err error
	c.RespData, err = c.tplEngine.Render(c.Req.Context(), tplName, data)
	if err != nil {
		c.RespStatusCode = http.StatusInternalServerError
		return err
	}
	c.RespStatusCode = http.StatusOK
	return nil
}

func (c *Context) RespJSON(status int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.RespData = data
	c.RespStatusCode = status
	return nil
}

func (c *Context) BindJSON(val any) error {
	if c.Req.Body == nil {
		return errors.New("web: body is nil")
	}
	return json.NewDecoder(c.Req.Body).Decode(val)
}

func (c *Context) FormValue(key string) *StringValue {
	if err := c.Req.ParseForm(); err != nil {
		return &StringValue{
			err: err,
		}
	}
	return &StringValue{
		val: c.Req.FormValue(key),
	}
}

func (c *Context) QueryValue(key string) *StringValue {
	if c.queryValues == nil {
		c.queryValues = c.Req.URL.Query()
	}
	val, ok := c.queryValues[key]
	if !ok {
		return &StringValue{
			err: errors.New(fmt.Sprintf("query param:%s not exists.", key)),
		}
	}
	return &StringValue{
		val: val[1],
	}
}

func (c *Context) PathValue(key string) *StringValue {
	val, ok := c.PathParams[key]
	if !ok {
		return &StringValue{
			err: errors.New(fmt.Sprintf("path param:%s not exixts.", key)),
		}
	}
	return &StringValue{
		val: val,
	}
}

type StringValue struct {
	val string
	err error
}

func (s *StringValue) AsInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	return strconv.ParseInt(s.val, 10, 64)
}
