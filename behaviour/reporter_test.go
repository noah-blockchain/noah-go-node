package behaviour_test

import (
	"sync"
	"testing"

	"github.com/noah-blockchain/noah-go-node/p2p"
)

// TestMockReporter tests the MockReporter's ability to store reported
// peer behaviour in memory indexed by the peerID.
func TestMockReporter(t *testing.T) {
	var peerID p2p.ID = "MockPeer"
	pr := NewMockReporter()

	behaviours := pr.GetBehaviours(peerID)
	if len(behaviours) != 0 {
		t.Error("Expected to have no behaviours reported")
	}

	badMessage := BadMessage(peerID, "bad message")
	pr.Report(badMessage)
	behaviours = pr.GetBehaviours(peerID)
	if len(behaviours) != 1 {
		t.Error("Expected the peer have one reported behaviour")
	}

	if behaviours[0] != badMessage {
		t.Error("Expected Bad Message to have been reported")
	}
}

type scriptItem struct {
	peerID    p2p.ID
	behaviour PeerBehaviour
}

// equalBehaviours returns true if a and b contain the same PeerBehaviours with
// the same freequencies and otherwise false.
func equalBehaviours(a []PeerBehaviour, b []PeerBehaviour) bool {
	aHistogram := map[PeerBehaviour]int{}
	bHistogram := map[PeerBehaviour]int{}

	for _, behaviour := range a {
		aHistogram[behaviour] += 1
	}

	for _, behaviour := range b {
		bHistogram[behaviour] += 1
	}

	if len(aHistogram) != len(bHistogram) {
		return false
	}

	for _, behaviour := range a {
		if aHistogram[behaviour] != bHistogram[behaviour] {
			return false
		}
	}

	for _, behaviour := range b {
		if bHistogram[behaviour] != aHistogram[behaviour] {
			return false
		}
	}

	return true
}

// TestEqualPeerBehaviours tests that equalBehaviours can tell that two slices
// of peer behaviours can be compared for the behaviours they contain and the
// freequencies that those behaviours occur.
func TestEqualPeerBehaviours(t *testing.T) {
	var (
		peerID        p2p.ID = "MockPeer"
		consensusVote        = ConsensusVote(peerID, "voted")
		blockPart            = BlockPart(peerID, "blocked")
		equals               = []struct {
			left  []PeerBehaviour
			right []PeerBehaviour
		}{
			// Empty sets
			{[]PeerBehaviour{}, []PeerBehaviour{}},
			// Single behaviours
			{[]PeerBehaviour{consensusVote}, []PeerBehaviour{consensusVote}},
			// Equal Frequencies
			{[]PeerBehaviour{consensusVote, consensusVote},
				[]PeerBehaviour{consensusVote, consensusVote}},
			// Equal frequencies different orders
			{[]PeerBehaviour{consensusVote, blockPart},
				[]PeerBehaviour{blockPart, consensusVote}},
		}
		unequals = []struct {
			left  []PeerBehaviour
			right []PeerBehaviour
		}{
			// Comparing empty sets to non empty sets
			{[]PeerBehaviour{}, []PeerBehaviour{consensusVote}},
			// Different behaviours
			{[]PeerBehaviour{consensusVote}, []PeerBehaviour{blockPart}},
			// Same behaviour with different frequencies
			{[]PeerBehaviour{consensusVote},
				[]PeerBehaviour{consensusVote, consensusVote}},
		}
	)

	for _, test := range equals {
		if !equalBehaviours(test.left, test.right) {
			t.Errorf("Expected %#v and %#v to be equal", test.left, test.right)
		}
	}

	for _, test := range unequals {
		if equalBehaviours(test.left, test.right) {
			t.Errorf("Expected %#v and %#v to be unequal", test.left, test.right)
		}
	}
}

// TestPeerBehaviourConcurrency constructs a scenario in which
// multiple goroutines are using the same MockReporter instance.
// This test reproduces the conditions in which MockReporter will
// be used within a Reactor `Receive` method tests to ensure thread safety.
func TestMockPeerBehaviourReporterConcurrency(t *testing.T) {
	var (
		behaviourScript = []struct {
			peerID     p2p.ID
			behaviours []PeerBehaviour
		}{
			{"1", []PeerBehaviour{ConsensusVote("1", "")}},
			{"2", []PeerBehaviour{ConsensusVote("2", ""), ConsensusVote("2", ""), ConsensusVote("2", "")}},
			{"3", []PeerBehaviour{BlockPart("3", ""), ConsensusVote("3", ""), BlockPart("3", ""), ConsensusVote("3", "")}},
			{"4", []PeerBehaviour{ConsensusVote("4", ""), ConsensusVote("4", ""), ConsensusVote("4", ""), ConsensusVote("4", "")}},
			{"5", []PeerBehaviour{BlockPart("5", ""), ConsensusVote("5", ""), BlockPart("5", ""), ConsensusVote("5", "")}},
		}
	)

	var receiveWg sync.WaitGroup
	pr := NewMockReporter()
	scriptItems := make(chan scriptItem)
	done := make(chan int)
	numConsumers := 3
	for i := 0; i < numConsumers; i++ {
		receiveWg.Add(1)
		go func() {
			defer receiveWg.Done()
			for {
				select {
				case pb := <-scriptItems:
					pr.Report(pb.behaviour)
				case <-done:
					return
				}
			}
		}()
	}

	var sendingWg sync.WaitGroup
	sendingWg.Add(1)
	go func() {
		defer sendingWg.Done()
		for _, item := range behaviourScript {
			for _, reason := range item.behaviours {
				scriptItems <- scriptItem{item.peerID, reason}
			}
		}
	}()

	sendingWg.Wait()

	for i := 0; i < numConsumers; i++ {
		done <- 1
	}

	receiveWg.Wait()

	for _, items := range behaviourScript {
		reported := pr.GetBehaviours(items.peerID)
		if !equalBehaviours(reported, items.behaviours) {
			t.Errorf("Expected peer %s to have behaved \nExpected: %#v \nGot %#v \n",
				items.peerID, items.behaviours, reported)
		}
	}
}
