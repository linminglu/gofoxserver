package chanrpc

import (
	"errors"
	"fmt"
	"time"

	"github.com/lovelly/leaf/log"
)

const (
	FuncCommon = iota
	FuncRoute
	FuncThis
)

// one server per goroutine (goroutine not safe)
// one client per goroutine (goroutine not safe)
type Server struct {
	// id -> function
	//
	// function:
	// func(args []interface{})
	// func(args []interface{}) interface{}
	// func(args []interface{}) []interface{}
	functions     map[interface{}]*FuncInfo
	ChanCall      chan *CallInfo
	RemoteAsynRet chan *RetInfo
	CloseFlg      bool
}

type FuncInfo struct {
	id    interface{}
	f     interface{}
	fType int
	this  interface{}
}

type CallInfo struct {
	fInfo   *FuncInfo
	args    []interface{}
	chanRet chan *RetInfo
	cb      interface{}
}

func BuildGoCallInfo(f *FuncInfo, args ...interface{}) *CallInfo {
	return &CallInfo{
		fInfo: f,
		args:  args,
	}
}

type RetInfo struct {
	// nil
	// interface{}
	// []interface{}
	Ret interface{}
	Err error
	// callback:
	// func(Err error)
	// func(Ret interface{}, Err error)
	// func(Ret []interface{}, Err error)
	Cb interface{}
}

type ExtRetFunc func(ret interface{}, err error)

type Client struct {
	s               *Server
	ChanSyncRet     chan *RetInfo
	ChanAsynRet     chan *RetInfo
	pendingAsynCall int
}

func NewServer(l int) *Server {
	s := new(Server)
	s.functions = make(map[interface{}]*FuncInfo)
	s.ChanCall = make(chan *CallInfo, l)
	s.RemoteAsynRet = make(chan *RetInfo, l)
	return s
}

func Assert(i interface{}) []interface{} {
	if i == nil {
		return nil
	} else {
		return i.([]interface{})
	}
}

func (s *Server) HasFunc(id interface{}) (*FuncInfo, bool) {
	f, ok := s.functions[id]
	return f, ok
}

// you must call the function before calling Open and Go
func (s *Server) RegisterFromType(id interface{}, f interface{}, fType int, this_param ...interface{}) {
	switch f.(type) {
	case func([]interface{}):
	case func([]interface{}) error:
	case func([]interface{}) (interface{}, error):
	case func([]interface{}) ([]interface{}, error):
	default:
		panic(fmt.Sprintf("function id %v: definition of function is invalid", id))
	}

	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}

	if len(this_param) > 0 {
		if fType != FuncThis {
			panic(fmt.Sprintf("function type not FuncThis, type:%v", fType))
		}
		s.functions[id] = &FuncInfo{id: id, f: f, fType: fType, this: this_param[0]}
	} else {
		s.functions[id] = &FuncInfo{id: id, f: f, fType: fType}
	}
}

func (s *Server) Register(id interface{}, f interface{}) {
	s.RegisterFromType(id, f, FuncCommon)
}

func (s *Server) ret(ci *CallInfo, ri *RetInfo) (err error) {
	if ci.chanRet == nil {
		if ci.cb != nil {
			ci.cb.(func(*RetInfo))(ri)
		}
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Recover(r)
		}
	}()

	ri.Cb = ci.cb
	ci.chanRet <- ri
	return
}

func (s *Server) exec(ci *CallInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Recover(r)
			s.ret(ci, &RetInfo{Err: fmt.Errorf("%v", r)})
		}
	}()

	if ci.fInfo.fType == FuncRoute {
		ci.args = append(ci.args, ci.fInfo.id)
	}

	if ci.fInfo.fType == FuncThis {
		ci.args = append(ci.args, ci.fInfo.this)
	}

	// execute
	retInfo := &RetInfo{}
	switch ci.fInfo.f.(type) {
	case func([]interface{}):
		ci.fInfo.f.(func([]interface{}))(ci.args)
	case func([]interface{}) error:
		retInfo.Err = ci.fInfo.f.(func([]interface{}) error)(ci.args)
	case func([]interface{}) (interface{}, error):
		retInfo.Ret, retInfo.Err = ci.fInfo.f.(func([]interface{}) (interface{}, error))(ci.args)
	case func([]interface{}) ([]interface{}, error):
		retInfo.Ret, retInfo.Err = ci.fInfo.f.(func([]interface{}) ([]interface{}, error))(ci.args)
	default:
		panic("bug")
	}

	return s.ret(ci, retInfo)
}

func (s *Server) Exec(ci *CallInfo) {
	if s.CloseFlg {
		log.Error("at call Exec chan is close %v", ci)
		return
	}
	err := s.exec(ci)
	if err != nil {
		log.Error("%v", err)
	}
}

// goroutine safe
func (s *Server) Go(id interface{}, args ...interface{}) {
	if s.CloseFlg {
		log.Error("at Go chan is close funcName : =====  %v ", id)
		return
	}
	f := s.functions[id]
	if f == nil {
		log.Error("function id %v: function not registered", id)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Recover(r)
		}
	}()

	s.ChanCall <- &CallInfo{
		fInfo: f,
		args:  args,
	}
}

// goroutine safe
func (s *Server) Call0(id interface{}, args ...interface{}) error {
	if s.CloseFlg {
		log.Error("at Call0 chan is close %v", id)
		return errors.New("] send on closed channel")
	}
	err := s.Open(0).Call0(id, args...)
	if err != nil {
		log.Error("call %s faild error:%s", id, err.Error())
	}
	return err
}

// goroutine safe
func (s *Server) Call1(id interface{}, args ...interface{}) (interface{}, error) {
	if s.CloseFlg {
		log.Error("at Call1 chan is close %v", id)
		return nil, errors.New("] send on closed channel")
	}
	return s.Open(0).Call1(id, args...)
}

func (s *Server) TimeOutCall1(id interface{}, t time.Duration, args ...interface{}) (interface{}, error) {
	if s.CloseFlg {
		log.Error("at TimeOutCall1 chan is close %v", id)
		return nil, errors.New("] send on closed channel")
	}
	return s.Open(0).TimeOutCall1(id, t, args...)
}

// goroutine safe
func (s *Server) CallN(id interface{}, args ...interface{}) ([]interface{}, error) {
	if s.CloseFlg {
		log.Error("at CallN chan is close %v", id)
		return nil, errors.New("] send on closed channel")
	}
	return s.Open(0).CallN(id, args...)
}

func (s *Server) Close() {
	if s.CloseFlg {
		log.Error(" double close Server chanAll")
		return
	}
	close(s.ChanCall)
	close(s.RemoteAsynRet)
	s.CloseFlg = true
	for ci := range s.ChanCall {
		s.ret(ci, &RetInfo{
			Err: errors.New("chanrpc server closed"),
		})
	}
}

// goroutine safe
func (s *Server) Open(l int) *Client {
	c := NewClient(l)
	c.Attach(s)
	return c
}

func NewClient(l int) *Client {
	c := new(Client)
	c.ChanSyncRet = make(chan *RetInfo, 1)
	c.ChanAsynRet = make(chan *RetInfo, l)
	return c
}

func (c *Client) Attach(s *Server) {
	c.s = s
}

func (c *Client) GetServer() *Server {
	return c.s
}

func (c *Client) call(ci *CallInfo, block bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Recover(r)
			err = fmt.Errorf("%v", r)
		}
	}()

	if block {
		c.s.ChanCall <- ci
	} else {
		select {
		case c.s.ChanCall <- ci:
		default:
			err = errors.New("chanrpc channel full")
		}
	}
	return
}

func (c *Client) f(id interface{}, n int) (fInfo *FuncInfo, err error) {
	if c.s == nil {
		err = errors.New("server not attached")
		return
	}

	fInfo = c.s.functions[id]
	if fInfo == nil {
		err = fmt.Errorf("function id %v: function not registered", id)
		return
	}

	var ok bool
	switch n {
	case 0:
		_, ok = fInfo.f.(func([]interface{}) error)
	case 1:
		_, ok = fInfo.f.(func([]interface{}) (interface{}, error))
	case 2:
		_, ok = fInfo.f.(func([]interface{}) ([]interface{}, error))
	case -1:
		ok = true
	default:
		panic("bug")
	}

	if !ok {
		err = fmt.Errorf("function id %v: return type mismatch", id)
	}
	return
}

func (c *Client) Call0(id interface{}, args ...interface{}) error {
	f, err := c.f(id, 0)
	if err != nil {
		return err
	}

	err = c.call(&CallInfo{
		fInfo:   f,
		args:    args,
		chanRet: c.ChanSyncRet,
	}, true)
	if err != nil {
		return err
	}

	ri := <-c.ChanSyncRet
	return ri.Err
}

func (c *Client) Call1(id interface{}, args ...interface{}) (interface{}, error) {
	f, err := c.f(id, 1)
	if err != nil {
		return nil, err
	}

	err = c.call(&CallInfo{
		fInfo:   f,
		args:    args,
		chanRet: c.ChanSyncRet,
	}, true)
	if err != nil {
		return nil, err
	}

	ri := <-c.ChanSyncRet
	return ri.Ret, ri.Err
}

func (c *Client) TimeOutCall1(id interface{}, t time.Duration, args ...interface{}) (interface{}, error) {
	f, err := c.f(id, 1)
	if err != nil {
		return nil, err
	}

	err = c.call(&CallInfo{
		fInfo:   f,
		args:    args,
		chanRet: c.ChanSyncRet,
	}, false)
	if err != nil {
		return nil, err
	}
	select {
	case ri := <-c.ChanSyncRet:
		return ri.Ret, ri.Err
	case <-time.After(time.Second * t):
		return nil, errors.New(fmt.Sprintf("time out at TimeOutCall1 function: %v", id))
	}
}

func (c *Client) CallN(id interface{}, args ...interface{}) ([]interface{}, error) {
	f, err := c.f(id, 2)
	if err != nil {
		return nil, err
	}

	err = c.call(&CallInfo{
		fInfo:   f,
		args:    args,
		chanRet: c.ChanSyncRet,
	}, true)
	if err != nil {
		return nil, err
	}

	ri := <-c.ChanSyncRet
	return Assert(ri.Ret), ri.Err
}

func (c *Client) RpcCall(id interface{}, args ...interface{}) {
	if len(args) < 1 {
		panic("callback function not found")
	}

	lastIndex := len(args) - 1
	cb := args[lastIndex]
	args = args[:lastIndex]

	var cbFunc func(*RetInfo)
	if cb != nil {
		cbFunc = cb.(func(*RetInfo))
	}

	var err error
	f := c.s.functions[id]
	if f == nil {
		err = fmt.Errorf("function id %v: function not registered", id)
		if cbFunc != nil {
			cbFunc(&RetInfo{Ret: nil, Err: err})
		}
		return
	}

	err = c.call(&CallInfo{
		fInfo: f,
		args:  args,
		cb:    cb,
	}, false)

	if err != nil && cbFunc != nil {
		cbFunc(&RetInfo{Ret: nil, Err: err})
	}
}

func (c *Client) asynCall(id interface{}, args []interface{}, cb interface{}, n int) {
	f, err := c.f(id, n)
	if err != nil {
		c.ChanAsynRet <- &RetInfo{Err: err, Cb: cb}
		return
	}

	err = c.call(&CallInfo{
		fInfo:   f,
		args:    args,
		chanRet: c.ChanAsynRet,
		cb:      cb,
	}, false)
	if err != nil {
		c.ChanAsynRet <- &RetInfo{Err: err, Cb: cb}
		return
	}
}

func (c *Client) AsynCall(id interface{}, _args ...interface{}) {
	if len(_args) < 1 {
		panic("callback function not found")
	}

	args := _args[:len(_args)-1]
	cb := _args[len(_args)-1]

	var n int
	switch cb.(type) {
	case func(error):
		n = 0
	case func(interface{}, error):
		n = 1
	case func([]interface{}, error):
		n = 2
	case ExtRetFunc:
		n = -1
	default:
		panic("definition of callback function is invalid")
	}

	// too many calls
	if c.pendingAsynCall >= cap(c.ChanAsynRet) {
		execCb(&RetInfo{Err: errors.New("too many calls"), Cb: cb})
		return
	}

	c.asynCall(id, args, cb, n)
	c.pendingAsynCall++
}

func execCb(ri *RetInfo) {
	defer func() {
		if r := recover(); r != nil {
			log.Recover(r)
		}
	}()

	if ri == nil {
		return
	}

	// execute
	switch ri.Cb.(type) {
	case func(error):
		ri.Cb.(func(error))(ri.Err)
	case func(interface{}, error):
		ri.Cb.(func(interface{}, error))(ri.Ret, ri.Err)
	case func([]interface{}, error):
		ri.Cb.(func([]interface{}, error))(Assert(ri.Ret), ri.Err)
	case ExtRetFunc:
		ri.Cb.(ExtRetFunc)(ri.Ret, ri.Err)
	default:
		panic("bug")
	}
	return
}

func (c *Client) Cb(ri *RetInfo) {
	c.pendingAsynCall--
	execCb(ri)
}

func (c *Client) ExecRemoveCb(ri *RetInfo) {
	execCb(ri)
}

func (c *Client) Close() {
	for c.pendingAsynCall > 0 {
		c.Cb(<-c.ChanAsynRet)
	}
}

func (c *Client) GetpendingAsynCall() int {
	log.Debug("pendingAsynCall %v", c.pendingAsynCall)
	return c.pendingAsynCall
}

func (c *Client) Idle() bool {
	return c.pendingAsynCall == 0
}
