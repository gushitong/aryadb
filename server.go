package main

import (
	"fmt"
	"github.com/gushitong/rekkles/engine"
	"github.com/gushitong/rekkles/stor"
	"github.com/gushitong/rekkles/ut"
	"github.com/pkg/errors"
	"github.com/tidwall/redcon"
)

type Options struct {
	Addr       string
	Auth       string
	Dir        string
	ValueDir   string
	SyncWrites bool
}

var DefaultOptions = Options{
	Addr:       ":6380",
	Auth:       "",
	Dir:        "/tmp/rekkles",
	ValueDir:   "/tmp/rekkles",
	SyncWrites: false,
}

type handler struct {
	Name string
	NArg int
	Func func(stor.DB, aryConnection, aryCommand)
}

type server struct {
	db       stor.DB
	Options  *Options
	Handlers map[string]*handler
}

func (s *server) Authenticate(conn aryConnection, auth string) error {
	if s.Options.Auth != auth {
		return errors.New("auth failed.")
	}
	ctx := conn.Context().(*Context)
	ctx.Authenticated = true
	return nil
}

func (s *server) Handle(redConn redcon.Conn, redCmd redcon.Command) {

	if len(redCmd.Args) == 0 {
		redConn.WriteString("ERR no arguments provided.")
		return
	}

	command := ut.LowerString(redCmd.Args[0])
	aryConn := aryConnection{redConn}
	aryCmd := aryCommand{Args: redCmd.Args[1:], Raw: redCmd.Raw}

	if command == "auth" {
		if len(aryCmd.Args) == 0 {
			aryConn.WriteError("wrong number of arguments for 'auth' command")
			return
		}
		if aryConn.Context() == nil {
			aryConn.SetContext(&Context{})
		}
		if err := s.Authenticate(aryConn, ut.LowerString(aryCmd.Args[0])); err != nil {
			aryConn.SetContext(&Context{})
			aryConn.WriteError("ERR auth failed")
			return
		}
		aryConn.SetContext(&Context{Authenticated: true})
		aryConn.WriteString("OK")
		return
	}

	if s.Options.Auth != "" && aryConn.Authenticated() == false {
		aryConn.WriteString("ERR auth required")
		return
	}

	f, err := s.GetHandler(command, aryCmd)
	if err != nil {
		aryConn.WriteError(err.Error())
		return
	}

	f(s.db, aryConn, aryCmd)
}

func (s *server) Register(cmd string, f func(stor.DB, aryConnection, aryCommand), narg int) {
	handler := &handler{
		Name: cmd, NArg: narg, Func: f,
	}
	s.Handlers[cmd] = handler
}

func (s *server) RegisterAll() {
	// string
	s.Register("append", _append, 2)
	s.Register("bitcount", bitcount, 1)
	s.Register("decr", decr, 1)
	s.Register("decrby", decrby, 2)
	s.Register("get", get, 1)
	s.Register("getbit", getbit, 2)
	s.Register("getrange", getrange, 3)
	s.Register("getset", getset, 2)
	s.Register("incr", incr, 1)
	s.Register("incrby", incrby, 2)
	s.Register("incrbyfloat", incrbyfloat, 2)
	s.Register("mget", mget, -1)
	s.Register("mset", mset, -1)
	s.Register("msetnx", msetnx, -1)
	s.Register("set", set, 2)
	s.Register("setbit", setbit, 3)
	s.Register("setex", setex, 2)
	s.Register("setnx", setnx, 2)
	s.Register("setrange", setrange, -1)
	s.Register("strlen", strlen, 1)

	// hash
	s.Register("hdel", hdel, 2)
	s.Register("hexists", hexists, 2)
	s.Register("hget", hget, 2)
	s.Register("hgetall", hgetall, 1)
	s.Register("hincrby", hincrby, 3)
	s.Register("hincrbyfloat", hincrbyfloat, 3)
	s.Register("hkeys", hkeys, 1)
	s.Register("hlen", hlen, 1)
	s.Register("hmget", hmget, -1)
	s.Register("hmset", hmset, -1)
	s.Register("hscan", hscan, -1)
	s.Register("hset", hset, 3)
	s.Register("hscan", hscan, -1)
	s.Register("hsetnx", hsetnx, 3)
	s.Register("hstrlen", hstrlen, 2)
	s.Register("hvals", hvals, 1)

	//list
	s.Register("lindex", lindex, 2)
	s.Register("llen", llen, 1)
	s.Register("lpop", lpop, 1)
	s.Register("lpush", lpush, 2)
	s.Register("lpushx", lpushx, 2)
	s.Register("lrange", lrange, 3)
	s.Register("lset", lset, 3)

	// set
	s.Register("sadd", sadd, -1)
	s.Register("scard", scard, 1)
	s.Register("sismember", sismember, 2)
	s.Register("smembers", smembers, 1)
	s.Register("spop", spop, -1)

	//zset
	s.Register("zadd", zadd, 3)
	s.Register("zcard", zcard, 1)
	s.Register("zcount", zcount, 3)
	s.Register("zincrby", zincrby, 3)
	s.Register("zpopmax", zpopmax, 1)
	s.Register("zpopmin", zpopmin, 1)
	s.Register("zrange", zrange, 3)
	s.Register("zrangebyscore", zrangebyscore, 3)
	s.Register("zrank", zrank, 2)
	s.Register("zrevrange", zrevrange, 3)
	s.Register("zrevrangebyscore", zrevrangebyscore, 3)
	s.Register("zrevrank", zrevrank, 2)
	s.Register("zscore", zscore, 2)

	//other
	s.Register("del", del, 1)
	s.Register("ping", ping, 0)
}

func (s *server) GetHandler(command string, aryCmd aryCommand) (func(stor.DB, aryConnection, aryCommand), error) {
	h, o := s.Handlers[command]
	if !o {
		return nil, fmt.Errorf("ERR command '%s' not supported", command)
	}
	if h.NArg >= 0 && len(aryCmd.Args) != h.NArg {
		return nil, ErrWrongNumOfArguments
	}
	return h.Func, nil
}

func (s *server) ListenAndSrv() error {
	return redcon.ListenAndServe(s.Options.Addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			s.Handle(conn, cmd)
		},
		func(conn redcon.Conn) bool {
			// use this function to accept or deny the connection.
			// log.Printf("accept: %stor", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// this is called when the connection has been closed
			// log.Printf("closed: %stor, err: %v", conn.RemoteAddr(), err)
		},
	)
}

func NewRekklesServer(opt *Options) (*server, error) {
	if opt.Dir == "" || opt.ValueDir == "" {
		return nil, errors.New("WorkingDir or ValueDir not provided.")
	}
	storage, err := engine.NewBadgerStorage(opt.Dir, opt.ValueDir, opt.SyncWrites)
	if err != nil {
		panic(err)
	}
	server := &server{
		db:       storage,
		Options:  opt,
		Handlers: make(map[string]*handler),
	}
	server.RegisterAll()
	return server, nil
}
