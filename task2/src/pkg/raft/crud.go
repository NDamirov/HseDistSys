package raft

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (r *Raft) statusCheck(c echo.Context) error {
	r.metaInfo.Lock()
	defer r.metaInfo.Unlock()

	if r.metaInfo.Status != Leader {
		return c.JSON(http.StatusServiceUnavailable, "Not leader")
	}
	return nil
}

func (r *Raft) CreateRequestHandler(c echo.Context) error {
	if err := r.statusCheck(c); err != nil {
		return err
	}

	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := r.storage.ValidateCreate(req.Key); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	entry := LogEntry{
		Command:      OpCreate,
		Key:          req.Key,
		Value:        &req.Value,
		CompareValue: nil,
	}

	err := r.Replicate(entry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}

func (r *Raft) ReadRequestHandler(c echo.Context) error {
	var req struct {
		Key string `json:"key"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := r.storage.ValidateGet(req.Key); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	result, err := r.storage.Get(req.Key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, struct {
		Value string `json:"value"`
	}{
		Value: result,
	})
}

func (r *Raft) UpdateRequestHandler(c echo.Context) error {
	if err := r.statusCheck(c); err != nil {
		return err
	}

	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := r.storage.ValidateSet(req.Key); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	entry := LogEntry{
		Command: OpSet,
		Key:     req.Key,
		Value:   &req.Value,
	}
	err := r.Replicate(entry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}

func (r *Raft) DeleteRequestHandler(c echo.Context) error {
	if err := r.statusCheck(c); err != nil {
		return err
	}
	var req struct {
		Key string `json:"key"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := r.storage.ValidateDelete(req.Key); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	entry := LogEntry{
		Command: OpDelete,
		Key:     req.Key,
	}
	err := r.Replicate(entry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}

func (r *Raft) CASRequestHandler(c echo.Context) error {
	if err := r.statusCheck(c); err != nil {
		return err
	}

	var req struct {
		Key          string `json:"key"`
		Value        string `json:"value"`
		CompareValue string `json:"compare_value"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := r.storage.ValidateCAS(req.Key, req.CompareValue); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	entry := LogEntry{
		Command:      OpCAS,
		Key:          req.Key,
		Value:        &req.Value,
		CompareValue: &req.CompareValue,
	}
	err := r.Replicate(entry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}

func (r *Raft) GetReplicasRequestHandler(c echo.Context) error {
	if err := r.statusCheck(c); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct {
		Replicas []int `json:"replicas"`
	}{
		Replicas: r.config.OtherPorts,
	})
}
