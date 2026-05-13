package handlers

import (
	"encoding/json"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/application/usecases"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/infra/http/dto"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/validators"
	"net/http"
)

type ServiceHandler struct {
	createUseCase  *usecases.CreateServiceUseCase
	getByIDUseCase *usecases.GetServiceByIDUseCase
	getAllUseCase  *usecases.GetAllServicesUseCase
	updateUseCase  *usecases.UpdateServiceUseCase
	deleteUseCase  *usecases.DeleteServiceUseCase
}

func NewServiceHandler(repository service.Repository) *ServiceHandler {
	return &ServiceHandler{
		createUseCase:  usecases.NewCreateServiceUseCase(repository),
		getByIDUseCase: usecases.NewGetServiceByIDUseCase(repository),
		getAllUseCase:  usecases.NewGetAllServicesUseCase(repository),
		updateUseCase:  usecases.NewUpdateServiceUseCase(repository),
		deleteUseCase:  usecases.NewDeleteServiceUseCase(repository),
	}
}

// CreateService godoc
// @Summary      Create a new service
// @Description  Create a new service in the catalog with name, description, and price
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateServiceRequest  true  "Service creation data"
// @Success      201      {object}  utils.Envelope{data=dto.ServiceResponse}
// @Failure      400      {object}  utils.Envelope
// @Failure      401      {object}  utils.Envelope
// @Failure      403      {object}  utils.Envelope
// @Failure      409      {object}  utils.Envelope
// @Failure      500      {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /services [post]
func (h *ServiceHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "service.create")
	defer span.End()

	var req dto.CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidRequest, "Invalid request body")
		return
	}

	if err := validators.ValidateStruct(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeValidationFailed, err.Error())
		return
	}

	input := usecases.CreateServiceInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}

	output, err := h.createUseCase.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		mapping := utils.MapDomainError(err)
		utils.RespondErrorEnvelope(w, mapping.StatusCode, mapping.Code, err.Error())
		return
	}

	response := dto.ServiceResponse{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		Price:       output.Price,
		CreatedAt:   utils.FormatTimeRFC3339(output.CreatedAt),
		UpdatedAt:   utils.FormatTimeRFC3339(output.UpdatedAt),
	}

	utils.RespondSuccess(w, http.StatusCreated, response)
}

// GetAllServices godoc
// @Summary      List all services
// @Description  Retrieve all services available in the catalog
// @Tags         services
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Envelope{data=[]dto.ServiceResponse}
// @Failure      401  {object}  utils.Envelope
// @Failure      403  {object}  utils.Envelope
// @Failure      500  {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /services [get]
func (h *ServiceHandler) GetAllServices(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "service.get_all")
	defer span.End()

	output, err := h.getAllUseCase.Execute(ctx)
	if err != nil {
		span.RecordError(err)
		mapping := utils.MapDomainError(err)
		utils.RespondErrorEnvelope(w, mapping.StatusCode, mapping.Code, err.Error())
		return
	}

	response := make([]dto.ServiceResponse, len(output.Services))
	for i, svc := range output.Services {
		response[i] = dto.ServiceResponse{
			ID:          svc.ID,
			Name:        svc.Name,
			Description: svc.Description,
			Price:       svc.Price,
			CreatedAt:   utils.FormatTimeRFC3339(svc.CreatedAt),
			UpdatedAt:   utils.FormatTimeRFC3339(svc.UpdatedAt),
		}
	}

	utils.RespondSuccess(w, http.StatusOK, response)
}

// GetServiceByID godoc
// @Summary      Get service by ID
// @Description  Retrieve detailed information about a specific service
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Service ID (UUID)"  example(550e8400-e29b-41d4-a716-446655440000)
// @Success      200  {object}  utils.Envelope{data=dto.ServiceResponse}
// @Failure      400  {object}  utils.Envelope
// @Failure      401  {object}  utils.Envelope
// @Failure      403  {object}  utils.Envelope
// @Failure      404  {object}  utils.Envelope
// @Failure      500  {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /services/{id} [get]
func (h *ServiceHandler) GetServiceByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "service.get_by_id")
	defer span.End()

	id := r.PathValue("id")

	// Validar UUID
	if err := utils.ValidateUUID(id); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid service ID format")
		return
	}

	input := usecases.GetServiceByIDInput{
		ID: id,
	}

	output, err := h.getByIDUseCase.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		mapping := utils.MapDomainError(err)
		utils.RespondErrorEnvelope(w, mapping.StatusCode, mapping.Code, err.Error())
		return
	}

	response := dto.ServiceResponse{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		Price:       output.Price,
		CreatedAt:   utils.FormatTimeRFC3339(output.CreatedAt),
		UpdatedAt:   utils.FormatTimeRFC3339(output.UpdatedAt),
	}

	utils.RespondSuccess(w, http.StatusOK, response)
}

// UpdateService godoc
// @Summary      Update service
// @Description  Update an existing service's information (partial updates supported)
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        id       path      string                    true  "Service ID (UUID)"  example(550e8400-e29b-41d4-a716-446655440000)
// @Param        request  body      dto.UpdateServiceRequest  true  "Service update data"
// @Success      200      {object}  utils.Envelope{data=object}
// @Failure      400      {object}  utils.Envelope
// @Failure      401      {object}  utils.Envelope
// @Failure      403      {object}  utils.Envelope
// @Failure      404      {object}  utils.Envelope
// @Failure      500      {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /services/{id} [put]
func (h *ServiceHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "service.update")
	defer span.End()

	id := r.PathValue("id")

	// Validar UUID
	if err := utils.ValidateUUID(id); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid service ID format")
		return
	}

	var req dto.UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidRequest, "Invalid request body")
		return
	}

	if err := validators.ValidateStruct(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeValidationFailed, err.Error())
		return
	}

	input := usecases.UpdateServiceInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}

	if err := h.updateUseCase.Execute(ctx, input); err != nil {
		span.RecordError(err)
		mapping := utils.MapDomainError(err)
		utils.RespondErrorEnvelope(w, mapping.StatusCode, mapping.Code, err.Error())
		return
	}

	utils.RespondSuccess(w, http.StatusOK, map[string]string{"message": "Service updated successfully"})
}

// DeleteService godoc
// @Summary      Delete service
// @Description  Delete a service from the catalog (soft delete)
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Service ID (UUID)"  example(550e8400-e29b-41d4-a716-446655440000)
// @Success      200  {object}  utils.Envelope{data=object}
// @Failure      400  {object}  utils.Envelope
// @Failure      401  {object}  utils.Envelope
// @Failure      403  {object}  utils.Envelope
// @Failure      404  {object}  utils.Envelope
// @Failure      500  {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /services/{id} [delete]
func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "service.delete")
	defer span.End()

	id := r.PathValue("id")

	// Validar UUID
	if err := utils.ValidateUUID(id); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid service ID format")
		return
	}

	input := usecases.DeleteServiceInput{
		ID: id,
	}

	if err := h.deleteUseCase.Execute(ctx, input); err != nil {
		span.RecordError(err)
		mapping := utils.MapDomainError(err)
		utils.RespondErrorEnvelope(w, mapping.StatusCode, mapping.Code, err.Error())
		return
	}

	utils.RespondSuccess(w, http.StatusOK, map[string]string{"message": "Service deleted successfully"})
}
