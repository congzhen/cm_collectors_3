package core

import "github.com/rs/xid"

func getXid() string {
	return xid.New().String()
}
