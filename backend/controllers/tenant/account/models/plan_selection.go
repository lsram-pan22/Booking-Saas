package models

import "errors"

var ErrPlanSelection = errors.New("invalid subscription scope selection")

type PlanSelection struct {
	ShopIDs   []string `json:"shop_ids"`
	MemberIDs []string `json:"member_ids"`
}
type PlanMember struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	Enabled bool   `json:"enabled"`
}
type PlanShop struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Status  string       `json:"status"`
	Enabled bool         `json:"enabled"`
	Members []PlanMember `json:"members"`
}
type PlanOverview struct {
	History []Subscription `json:"history"`
	Shops   []PlanShop     `json:"shops"`
}
