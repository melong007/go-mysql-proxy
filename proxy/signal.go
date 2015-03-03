// Copyright (c) 2014 Xiaomi.com, Inc. All Rights Reserved
// @file    signal.go
// @author  王靖 (wangjing1@xiaomi.com)
// @date    14-11-26 16:04:09
// @version $Revision: 1.0 $
// @brief

package proxy

import (
	. "github.com/wangjild/go-mysql-proxy/log"
	"github.com/wangjild/go-mysql-proxy/signal"
	"os"
	"os/signal"
	"syscall"
)

// ignore SIGPIPE
func sigPipeHandler(s os.Signal, arg interface{}) error {
	return nil
}

func processSignals() {
	ss := signal2.NewSignalSet()
	ss.Register(syscall.SIGPIPE, sigPipeHandler)

	for {
		c := make(chan os.Signal, 16)
		var sigs []os.Signal
		for sig := range ss.M {
			sigs = append(sigs, sig)
		}

		signal.Notify(c, sigs...)
		sig := <-c
		if err := ss.Handle(sig, nil); err != nil {
			SysLog.Warn("hanle signal %v error: %s", sig, err.Error())
		}

		if sig == syscall.SIGINT || sig == syscall.SIGTERM {
			return
		}
	}
}

/* vim: set expandtab ts=4 sw=4 */
