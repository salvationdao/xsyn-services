package rpcclient

import (
	"github.com/gofrs/uuid"
)

type MechsReq struct {
}
type MechsResp struct {
	MechContainers []*MechContainer
}

type MechReq struct {
	MechID uuid.UUID
}

type MechResp struct {
	MechContainer *MechContainer
}

type MechsByOwnerIDReq struct {
	OwnerID uuid.UUID
}
type MechsByOwnerIDResp struct {
	MechContainers []*MechContainer
}

type MechRegisterReq struct {
	TemplateID uuid.UUID
	OwnerID    string
}
type MechRegisterResp struct {
	MechContainer *MechContainer
}

type MechSetNameReq struct {
	MechID uuid.UUID
	Name   string
}
type MechSetNameResp struct {
	MechContainer *MechContainer
}

type MechSetOwnerReq struct {
	MechID  uuid.UUID
	OwnerID uuid.UUID
}
type MechSetOwnerResp struct {
	MechContainer *MechContainer
}

type TemplatesReq struct {
}
type TemplatesResp struct {
	TemplateContainers []*TemplateContainer
}

type TemplateReq struct {
	TemplateID uuid.UUID
}
type TemplateResp struct {
	TemplateContainer *TemplateContainer
}

type TemplatePurchasedCountReq struct {
	TemplateID uuid.UUID
}
type TemplatePurchasedCountResp struct {
	Count int
}

type TemplatesByFactionIDReq struct {
	FactionID uuid.UUID
}
type TemplatesByFactionIDResp struct {
	TemplateContainers []*TemplateContainer
}
