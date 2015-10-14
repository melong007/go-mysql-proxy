package proxy

import (
	"fmt"
	"time"

	"github.com/melong007/go-mysql/query"
	"github.com/wangjild/go-mysql-proxy/client"
	"github.com/wangjild/go-mysql-proxy/hack"
	"github.com/wangjild/go-mysql-proxy/log"
	. "github.com/wangjild/go-mysql-proxy/mysql"
	"github.com/wangjild/go-mysql-proxy/sql"
)

//we just go the microsecond timestamp
func getTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (c *Conn) handleQuery(sqlstmt string) (err error) {
	/*defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("execute %s error %v", sql, e)
			return
		}
	}()*/

	var stmt sql.IStatement
	var excess int64

	fp := query.Fingerprint(sqlstmt)
	now := getTimestamp()

	c.server.mu.Lock()
	if lr, ok := c.server.fingerprints[fp]; ok {
		//how many microsecond elapsed since last query
		ms := now - lr.last
		//Default, we have 1 r/s
		excess = lr.excess - c.server.cfg.ReqRate*(ms/1000) + 1000

		//If we need caculate every second speed,
		//Should reset to zero;
		if excess < 0 {
			excess = 0
		}

		//the race out the max Burst?
		log.AppLog.Notice("the Query excess(%d), the reqBurst(%d)", excess, c.server.cfg.ReqBurst)
		if excess > c.server.cfg.ReqBurst {

			c.server.mu.Unlock()
			//Just close the client or
			return fmt.Errorf(`the query excess(%d) over the reqBurst(%d), sql: %s "`, excess, c.server.cfg.ReqBurst, sqlstmt)
			//TODO: more gracefully add a Timer and retry?
		}
		lr.excess = excess
		lr.last = now
		lr.count++

	} else {
		lr := &LimitReqNode{}
		lr.excess = 0
		lr.last = getTimestamp()
		lr.query = fp

		lr.count = 1
		c.server.fingerprints[fp] = lr
	}
	c.server.mu.Unlock()

	stmt, err = sql.Parse(sqlstmt)
	if err != nil {
		return fmt.Errorf(`parse sql "%s" error "%s"`, sqlstmt, err.Error())
	}

	switch v := stmt.(type) {
	case sql.ISelect:
		return c.handleSelect(v, sqlstmt)
	case *sql.Insert:
		return c.handleExec(stmt, sqlstmt, false)
	case *sql.Update:
		return c.handleExec(stmt, sqlstmt, false)
	case *sql.Delete:
		return c.handleExec(stmt, sqlstmt, false)
	case *sql.Replace:
		return c.handleExec(stmt, sqlstmt, false)
	case *sql.Set:
		return c.handleSet(v, sqlstmt)
	case *sql.Begin:
		return c.handleBegin()
	case *sql.Commit:
		return c.handleCommit()
	case *sql.Rollback:
		return c.handleRollback()
	case sql.IShow:
		return c.handleShow(sqlstmt, v)
	case sql.IDDLStatement:
		return c.handleExec(stmt, sqlstmt, false)
	case *sql.Do:
		return c.handleExec(stmt, sqlstmt, false)
	case *sql.Call:
		return c.handleExec(stmt, sqlstmt, false)
	case *sql.Use:
		if err := c.useDB(hack.String(stmt.(*sql.Use).DB)); err != nil {
			return err
		} else {
			return c.writeOK(nil)
		}

	default:
		return fmt.Errorf("statement %T[%s] not support now", stmt, sqlstmt)
	}

	return nil
}

func (c *Conn) getConn(n *Node, isSelect bool) (co *client.SqlConn, err error) {
	if !c.needBeginTx() {
		if isSelect {
			co, err = n.getSelectConn()
		} else {
			co, err = n.getMasterConn()
		}
		if err != nil {
			return
		}
	} else {
		var ok bool
		c.Lock()
		co, ok = c.txConns[n]
		c.Unlock()

		if !ok {
			if co, err = n.getMasterConn(); err != nil {
				return
			}

			if err = co.SetAutocommit(c.IsAutoCommit()); err != nil {
				return
			}

			if err = co.Begin(); err != nil {
				return
			}

			c.Lock()
			c.txConns[n] = co
			c.Unlock()
		}
	}

	//todo, set conn charset, etc...
	if err = co.UseDB(c.schema.db); err != nil {
		return
	}

	if err = co.SetCharset(c.charset); err != nil {
		return
	}

	return
}

func (c *Conn) closeDBConn(co *client.SqlConn, rollback bool) {
	// since we have DDL, and when server is not in autoCommit,
	// we do not release the connection and will reuse it later
	if c.isInTransaction() || !c.isAutoCommit() {
		return
	}

	if rollback {
		co.Rollback()
	}

	co.Close()
}

func makeBindVars(args []interface{}) map[string]interface{} {
	bindVars := make(map[string]interface{}, len(args))

	for i, v := range args {
		bindVars[fmt.Sprintf("v%d", i+1)] = v
	}

	return bindVars
}

func (c *Conn) handleExec(stmt sql.IStatement, sqlstmt string, isread bool) error {

	if err := c.checkDB(); err != nil {
		return err
	}

	conn, err := c.getConn(c.schema.node, isread)
	if err != nil {
		return err
	} else if conn == nil {
		return fmt.Errorf("no available connection")
	}

	var rs *Result
	rs, err = conn.Execute(sqlstmt)

	c.closeDBConn(conn, err != nil)

	if err == nil {
		err = c.writeOK(rs)
	}

	return err
}

func (c *Conn) mergeSelectResult(rs *Result) error {
	r := rs.Resultset
	status := c.status | rs.Status
	return c.writeResultset(status, r)
}
