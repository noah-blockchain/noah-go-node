package mock

import (
	p2p2 "github.com/noah-blockchain/noah-go-node/p2p"
	conn2 "github.com/noah-blockchain/noah-go-node/p2p/conn"
	"net"

	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type Peer struct {
	*cmn.BaseService
	ip                   net.IP
	id                   p2p2.ID
	addr                 *p2p2.NetAddress
	kv                   map[string]interface{}
	Outbound, Persistent bool
}

// NewPeer creates and starts a new mock peer. If the ip
// is nil, random routable address is used.
func NewPeer(ip net.IP) *Peer {
	var netAddr *p2p2.NetAddress
	if ip == nil {
		_, netAddr = p2p2.CreateRoutableAddr()
	} else {
		netAddr = p2p2.NewNetAddressIPPort(ip, 26656)
	}
	nodeKey := p2p2.NodeKey{PrivKey: ed25519.GenPrivKey()}
	netAddr.ID = nodeKey.ID()
	mp := &Peer{
		ip:   ip,
		id:   nodeKey.ID(),
		addr: netAddr,
		kv:   make(map[string]interface{}),
	}
	mp.BaseService = cmn.NewBaseService(nil, "MockPeer", mp)
	mp.Start()
	return mp
}

func (mp *Peer) FlushStop()                              { mp.Stop() }
func (mp *Peer) TrySend(chID byte, msgBytes []byte) bool { return true }
func (mp *Peer) Send(chID byte, msgBytes []byte) bool    { return true }
func (mp *Peer) NodeInfo() p2p2.NodeInfo {
	return p2p2.DefaultNodeInfo{
		ID_:        mp.addr.ID,
		ListenAddr: mp.addr.DialString(),
	}
}
func (mp *Peer) Status() conn2.ConnectionStatus { return conn2.ConnectionStatus{} }
func (mp *Peer) ID() p2p2.ID                    { return mp.id }
func (mp *Peer) IsOutbound() bool               { return mp.Outbound }
func (mp *Peer) IsPersistent() bool             { return mp.Persistent }
func (mp *Peer) Get(key string) interface{} {
	if value, ok := mp.kv[key]; ok {
		return value
	}
	return nil
}
func (mp *Peer) Set(key string, value interface{}) {
	mp.kv[key] = value
}
func (mp *Peer) RemoteIP() net.IP             { return mp.ip }
func (mp *Peer) SocketAddr() *p2p2.NetAddress { return mp.addr }
func (mp *Peer) RemoteAddr() net.Addr         { return &net.TCPAddr{IP: mp.ip, Port: 8800} }
func (mp *Peer) CloseConn() error             { return nil }
