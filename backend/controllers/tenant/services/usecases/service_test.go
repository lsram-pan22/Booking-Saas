package usecases

import (
	"booking-saas/controllers/tenant/services/models"
	"context"
	"errors"
	"testing"
)

type fakeRepo struct {
	calls   int
	request models.ServiceRequest
}

func (r *fakeRepo) List(context.Context, string, string) (models.ServiceList, error) {
	r.calls++
	return models.ServiceList{}, nil
}
func (r *fakeRepo) Save(_ context.Context, _, _, _ string, req models.ServiceRequest) (models.Service, error) {
	r.calls++
	r.request = req
	return models.Service{}, nil
}
func TestServiceValidation(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	valid := models.ServiceRequest{Name: "  ตัดผม  ", Description: "  รายละเอียด  ", DurationMinutes: 30, PriceSatang: 0, Status: "active"}
	cases := []struct {
		name   string
		mutate func(*models.ServiceRequest)
	}{
		{"blank name", func(r *models.ServiceRequest) { r.Name = "  " }},
		{"zero duration", func(r *models.ServiceRequest) { r.DurationMinutes = 0 }},
		{"long duration", func(r *models.ServiceRequest) { r.DurationMinutes = 1441 }},
		{"negative price", func(r *models.ServiceRequest) { r.PriceSatang = -1 }},
		{"large price", func(r *models.ServiceRequest) { r.PriceSatang = 100000001 }},
		{"invalid status", func(r *models.ServiceRequest) { r.Status = "suspended" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &fakeRepo{}
			req := valid
			tc.mutate(&req)
			_, err := NewServiceUsecase(r).Save(context.Background(), id, id, "", req)
			if !errors.Is(err, ErrInvalidService) || r.calls != 0 {
				t.Fatal("invalid request reached repository", err)
			}
		})
	}
	r := &fakeRepo{}
	u := NewServiceUsecase(r)
	if _, err := u.Save(context.Background(), id, id, "", valid); err != nil || r.request.Name != "ตัดผม" || r.request.Description != "รายละเอียด" || r.request.PriceSatang != 0 {
		t.Fatal("valid free service rejected or not normalized", err)
	}
	if _, err := u.Save(context.Background(), id, id, "invalid", valid); !errors.Is(err, ErrInvalidService) {
		t.Fatal("invalid service ID accepted")
	}
	if _, err := u.List(context.Background(), id, "invalid"); !errors.Is(err, ErrInvalidService) {
		t.Fatal("invalid shop ID accepted")
	}
}
