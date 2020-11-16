package pairmint

import (
	"container/list"
	"log"
)

// Pairminter defines the attributes of a pairminter node.
type Pairminter struct {
	// The logger used to print pairmint log messages.
	logger *log.Logger

	// The pairminter's configuration.
	config *Config

	// The node's current rank in the queue.
	// All pairminter nodes are part of a ranking system: rank #1
	// is always the signer while the ranks below that serve as
	// backups that gradually move up the ranks if the current
	// signer misses too many blocks in a row.
	rank uint

	// The amount of consecutive blocks that are constantly
	// monitored for missed blocks in order to determine when a
	// rank update should happen and the current signer should
	// move to the back of the queue.
	frame *list.List
}

// UpdateRank updates the rank of a pairminter node.
func (p *Pairminter) UpdateRank() {
	if p.rank == 1 { // If node is currently rank #1, move it to the back of the queue.
		p.logger.Printf("Updating rank (%v -> %v)", p.rank, p.config.QueueSize)
		p.rank = p.config.QueueSize
	} else { // Else move the node up one rank.
		p.logger.Printf("Updating rank (%v -> %v)", p.rank, p.rank-1)
		p.rank--
	}
}

// UpdateFrame updates the block frame containing signed and
// missed blocks.
func (p *Pairminter) UpdateFrame(height uint) {
	if p.frame.Len() >= 10 {
		p.frame.Remove(p.frame.Back())
	}
	p.frame.PushFront(height)
}

// ClearFrame clears the block frame containing signed and
// missed blocks.
func (p *Pairminter) ClearFrame() {
	p.frame = list.New()
}
