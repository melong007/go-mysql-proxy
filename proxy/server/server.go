// Copyright (c) 2014 Xiaomi.com, Inc. All Rights Reserved
// @file    server.go
// @author  王靖 (wangjing1@xiaomi.com)
// @date    14-11-27 11:11:59
// @version $Revision: 1.0 $
// @brief

package server

import ()

type Proxy interface {
	Run() error
	Stop() error
	Close()
}

/* vim: set expandtab ts=4 sw=4 */
