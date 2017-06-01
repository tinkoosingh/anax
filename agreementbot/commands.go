package agreementbot

import (
	"fmt"
	"github.com/open-horizon/anax/events"
)

// ==============================================================================================================
type AgreementTimeoutCommand struct {
	AgreementId string
	Protocol    string
	Reason      uint
}

func (a AgreementTimeoutCommand) ShortString() string {
	return fmt.Sprintf("%v", a)
}

func NewAgreementTimeoutCommand(agreementId string, protocol string, reason uint) *AgreementTimeoutCommand {
	return &AgreementTimeoutCommand{
		AgreementId: agreementId,
		Protocol:    protocol,
		Reason:      reason,
	}
}

// ==============================================================================================================
type PolicyChangedCommand struct {
	Msg events.PolicyChangedMessage
}

func (p PolicyChangedCommand) ShortString() string {
	return fmt.Sprintf("%v", p)
}

func NewPolicyChangedCommand(msg events.PolicyChangedMessage) *PolicyChangedCommand {
	return &PolicyChangedCommand{
		Msg: msg,
	}
}

// ==============================================================================================================
type PolicyDeletedCommand struct {
	Msg events.PolicyDeletedMessage
}

func (p PolicyDeletedCommand) ShortString() string {
	return fmt.Sprintf("%v", p)
}

func NewPolicyDeletedCommand(msg events.PolicyDeletedMessage) *PolicyDeletedCommand {
	return &PolicyDeletedCommand{
		Msg: msg,
	}
}

// ==============================================================================================================
type NewProtocolMessageCommand struct {
	Message   []byte
	MessageId int
	From      string
	PubKey    []byte
}

func (p NewProtocolMessageCommand) ShortString() string {
	return fmt.Sprintf("%v", p)
}

func NewNewProtocolMessageCommand(msg []byte, msgId int, deviceId string, pubkey []byte) *NewProtocolMessageCommand {
	return &NewProtocolMessageCommand{
		Message:   msg,
		MessageId: msgId,
		From:      deviceId,
		PubKey:    pubkey,
	}
}

// ==============================================================================================================
type BlockchainEventCommand struct {
	Msg events.EthBlockchainEventMessage
}

func (e BlockchainEventCommand) ShortString() string {
	return e.Msg.ShortString()
}

func NewBlockchainEventCommand(msg events.EthBlockchainEventMessage) *BlockchainEventCommand {
	return &BlockchainEventCommand{
		Msg: msg,
	}
}

// ==============================================================================================================
type WorkloadUpgradeCommand struct {
	Msg events.ABApiWorkloadUpgradeMessage
}

func (e WorkloadUpgradeCommand) ShortString() string {
	return e.Msg.ShortString()
}

func NewWorkloadUpgradeCommand(msg events.ABApiWorkloadUpgradeMessage) *WorkloadUpgradeCommand {
	return &WorkloadUpgradeCommand{
		Msg: msg,
	}
}
