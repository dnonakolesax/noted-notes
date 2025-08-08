package handlers

import "github.com/google/uuid"


type KernelService interface {
	Start() (uuid.UUID, error)
	Stop(initiator uuid.UUID, kernelId uuid.UUID) error
}