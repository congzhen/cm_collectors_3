package core

import "github.com/rs/xid"

func getXid() string {
	guid := xid.New().String()
	return guid
}
