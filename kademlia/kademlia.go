package kademlia

// Kademlia represents the Kademlia distributed hash table.
type Kademlia struct {
	Network *Network
}

// NewKademlia initializes a new Kademlia instance.
func NewKademlia(localAddr string) *Kademlia {
	network := &Network{
		LocalID:   NewRandomKademliaID(),
		LocalAddr: localAddr,
	}
	return &Kademlia{
		Network: network,
	}
}

func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO
}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}
