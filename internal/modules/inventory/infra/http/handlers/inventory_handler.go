package handlers

import (
	"encoding/json"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/application/usecases"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/infra/http/dto"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/infra/observability"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/utils"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/shared/validators"
	"net/http"
)

type InventoryHandler struct {
	createInventoryUC        *usecases.CreateInventoryUseCase
	getInventoryUC           *usecases.GetInventoryUseCase
	getInventoryByProductUC  *usecases.GetInventoryByProductUseCase
	deleteInventoryUC        *usecases.DeleteInventoryUseCase
	manualDecreaseStockUC    *usecases.ManualDecreaseStockUseCase
	reservedDecreaseStockUC  *usecases.ReservedDecreaseStockUseCase
	availableDecreaseStockUC *usecases.AvailableDecreaseStockUseCase
	reserveStockUC           *usecases.ReserveStockUseCase
	increaseStockUC          *usecases.IncreaseStockUseCase
	cancelReservedStockUC    *usecases.CancelReservedStockUseCase
	cancelConfirmedStockUC   *usecases.CancelConfirmedStockUseCase
}

func NewInventoryHandler(inventoryRepo inventory.Repository, productRepo product.Repository) *InventoryHandler {
	return &InventoryHandler{
		createInventoryUC:        usecases.NewCreateInventoryUseCase(inventoryRepo, productRepo),
		getInventoryUC:           usecases.NewGetInventoryUseCase(inventoryRepo),
		getInventoryByProductUC:  usecases.NewGetInventoryByProductUseCase(inventoryRepo),
		deleteInventoryUC:        usecases.NewDeleteInventoryUseCase(inventoryRepo),
		manualDecreaseStockUC:    usecases.NewManualDecreaseStockUseCase(inventoryRepo),
		reservedDecreaseStockUC:  usecases.NewReservedDecreaseStockUseCase(inventoryRepo),
		availableDecreaseStockUC: usecases.NewAvailableDecreaseStockUseCase(inventoryRepo),
		reserveStockUC:           usecases.NewReserveStockUseCase(inventoryRepo),
		increaseStockUC:          usecases.NewIncreaseStockUseCase(inventoryRepo),
		cancelReservedStockUC:    usecases.NewCancelReservedStockUseCase(inventoryRepo),
		cancelConfirmedStockUC:   usecases.NewCancelConfirmedStockUseCase(inventoryRepo),
	}
}

// CreateInventory godoc
// @Summary      Create inventory for product
// @Description  Initializes inventory tracking for a product
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Param        product_id  path      string  true  "Product ID (UUID)"
// @Success      201         {object}  utils.Envelope{data=dto.InventoryResponse}
// @Failure      400         {object}  utils.Envelope
// @Failure      401         {object}  utils.Envelope
// @Failure      404         {object}  utils.Envelope
// @Failure      409         {object}  utils.Envelope
// @Failure      500         {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /products/{product_id}/inventory [post]
func (h *InventoryHandler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "inventory.create")
	defer span.End()

	productID := r.PathValue("product_id")

	if err := utils.ValidateUUID(productID); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	input := usecases.CreateInventoryInput{
		ProductID: productID,
	}

	output, err := h.createInventoryUC.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		statusCode, errCode := mapInventoryError(err)
		utils.RespondErrorEnvelope(w, statusCode, errCode, err.Error())
		return
	}

	response := &dto.InventoryResponse{
		ID:                output.ID,
		ProductID:         output.ProductID,
		AvailableQuantity: output.AvailableQuantity,
		ReservedQuantity:  output.ReservedQuantity,
		PendingQuantity:   output.PendingQuantity,
		CreatedAt:         output.CreatedAt,
		UpdatedAt:         output.UpdatedAt,
	}
	utils.RespondSuccess(w, http.StatusCreated, response)
}

// GetInventory handles GET /api/products/:product_id/inventory (removed - use GetInventoryByProduct)

// GetInventoryByProduct godoc
// @Summary      Get inventory by product
// @Description  Retrieves inventory information for a specific product
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Param        product_id  path      string  true  "Product ID (UUID)"
// @Success      200         {object}  utils.Envelope{data=dto.InventoryResponse}
// @Failure      400         {object}  utils.Envelope
// @Failure      401         {object}  utils.Envelope
// @Failure      404         {object}  utils.Envelope
// @Failure      500         {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /products/{product_id}/inventory [get]
func (h *InventoryHandler) GetInventoryByProduct(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "inventory.get_by_product")
	defer span.End()

	productID := r.PathValue("product_id")

	if err := utils.ValidateUUID(productID); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	input := usecases.GetInventoryByProductInput{
		ProductID: productID,
	}

	output, err := h.getInventoryByProductUC.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		statusCode, errCode := mapInventoryError(err)
		utils.RespondErrorEnvelope(w, statusCode, errCode, err.Error())
		return
	}

	response := &dto.InventoryResponse{
		ID:                output.ID,
		ProductID:         output.ProductID,
		AvailableQuantity: output.AvailableQuantity,
		ReservedQuantity:  output.ReservedQuantity,
		PendingQuantity:   output.PendingQuantity,
		CreatedAt:         output.CreatedAt,
		UpdatedAt:         output.UpdatedAt,
	}
	utils.RespondSuccess(w, http.StatusOK, response)
}

// DeleteInventory handles DELETE /api/products/:product_id/inventory
func (h *InventoryHandler) DeleteInventory(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "inventory.delete")
	defer span.End()

	productID := r.PathValue("product_id")

	if err := utils.ValidateUUID(productID); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	// First, get the inventory by product ID
	getInput := usecases.GetInventoryByProductInput{
		ProductID: productID,
	}

	getOutput, err := h.getInventoryByProductUC.Execute(ctx, getInput)
	if err != nil {
		span.RecordError(err)
		statusCode, errCode := mapInventoryError(err)
		utils.RespondErrorEnvelope(w, statusCode, errCode, err.Error())
		return
	}

	// Then delete using the inventory ID
	deleteInput := usecases.DeleteInventoryInput{
		ID: getOutput.ID,
	}

	if err := h.deleteInventoryUC.Execute(ctx, deleteInput); err != nil {
		span.RecordError(err)
		statusCode, errCode := mapInventoryError(err)
		utils.RespondErrorEnvelope(w, statusCode, errCode, err.Error())
		return
	}

	utils.RespondSuccess(w, http.StatusOK, map[string]string{"message": "Inventory deleted successfully"})
}

// ManualDecrease godoc
// @Summary      Manually decrease stock
// @Description  Manually decreases available stock (e.g., for damaged or lost items)
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Param        product_id  path      string                     true  "Product ID (UUID)"
// @Param        request     body      dto.StockMovementRequest   true  "Stock decrease data"
// @Success      200         {object}  utils.Envelope{data=dto.InventoryResponse}
// @Failure      400         {object}  utils.Envelope
// @Failure      401         {object}  utils.Envelope
// @Failure      404         {object}  utils.Envelope
// @Failure      422         {object}  utils.Envelope
// @Failure      500         {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /products/{product_id}/inventory/manual-decrease [post]
func (h *InventoryHandler) ManualDecrease(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "inventory.manual_decrease")
	defer span.End()

	productID := r.PathValue("product_id")

	if err := utils.ValidateUUID(productID); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	var req dto.StockMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidRequest, "Invalid request body")
		return
	}

	if err := validators.ValidateStruct(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeValidationFailed, err.Error())
		return
	}

	input := usecases.ManualDecreaseStockInput{
		ProductID: productID,
		Quantity:  req.Quantity,
	}

	output, err := h.manualDecreaseStockUC.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		statusCode, errCode := mapInventoryError(err)
		utils.RespondErrorEnvelope(w, statusCode, errCode, err.Error())
		return
	}

	response := &dto.InventoryResponse{
		ID:                output.ID,
		ProductID:         output.ProductID,
		AvailableQuantity: output.AvailableQuantity,
		ReservedQuantity:  output.ReservedQuantity,
		PendingQuantity:   output.PendingQuantity,
		CreatedAt:         "",
		UpdatedAt:         output.UpdatedAt,
	}
	utils.RespondSuccess(w, http.StatusOK, response)
}

// ReservedDecrease godoc
// @Summary      Confirm reserved stock
// @Description  Confirms reserved stock usage (decreases reserved quantity)
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Param        product_id  path      string                     true  "Product ID (UUID)"
// @Param        request     body      dto.StockMovementRequest   true  "Stock confirmation data"
// @Success      200         {object}  utils.Envelope{data=dto.InventoryResponse}
// @Failure      400         {object}  utils.Envelope
// @Failure      401         {object}  utils.Envelope
// @Failure      404         {object}  utils.Envelope
// @Failure      422         {object}  utils.Envelope
// @Failure      500         {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /products/{product_id}/inventory/reserved-decrease [post]
func (h *InventoryHandler) ReservedDecrease(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "inventory.reserved_decrease")
	defer span.End()

	productID := r.PathValue("product_id")

	if err := utils.ValidateUUID(productID); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	var req dto.StockMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidRequest, "Invalid request body")
		return
	}

	if err := validators.ValidateStruct(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeValidationFailed, err.Error())
		return
	}

	input := usecases.ReservedDecreaseStockInput{
		ProductID: productID,
		Quantity:  req.Quantity,
	}

	output, err := h.reservedDecreaseStockUC.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		statusCode, errCode := mapInventoryError(err)
		utils.RespondErrorEnvelope(w, statusCode, errCode, err.Error())
		return
	}

	response := &dto.InventoryResponse{
		ID:                output.ID,
		ProductID:         output.ProductID,
		AvailableQuantity: output.AvailableQuantity,
		ReservedQuantity:  output.ReservedQuantity,
		PendingQuantity:   output.PendingQuantity,
		CreatedAt:         "",
		UpdatedAt:         output.UpdatedAt,
	}
	utils.RespondSuccess(w, http.StatusOK, response)
}

// AvailableDecrease handles POST /api/products/:product_id/inventory/available-decrease
func (h *InventoryHandler) AvailableDecrease(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "inventory.available_decrease")
	defer span.End()

	productID := r.PathValue("product_id")

	if err := utils.ValidateUUID(productID); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	var req dto.StockMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidRequest, "Invalid request body")
		return
	}

	if err := validators.ValidateStruct(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeValidationFailed, err.Error())
		return
	}

	input := usecases.AvailableDecreaseStockInput{
		ProductID: productID,
		Quantity:  req.Quantity,
	}

	output, err := h.availableDecreaseStockUC.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		statusCode, errCode := mapInventoryError(err)
		utils.RespondErrorEnvelope(w, statusCode, errCode, err.Error())
		return
	}

	response := &dto.InventoryResponse{
		ID:                output.ID,
		ProductID:         output.ProductID,
		AvailableQuantity: output.AvailableQuantity,
		ReservedQuantity:  output.ReservedQuantity,
		PendingQuantity:   output.PendingQuantity,
		CreatedAt:         "",
		UpdatedAt:         output.UpdatedAt,
	}
	utils.RespondSuccess(w, http.StatusOK, response)
}

// Reserve godoc
// @Summary      Reserve stock
// @Description  Reserves available stock for a service order (moves from available to reserved)
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Param        product_id  path      string                     true  "Product ID (UUID)"
// @Param        request     body      dto.StockMovementRequest   true  "Stock reservation data"
// @Success      200         {object}  utils.Envelope{data=dto.InventoryResponse}
// @Failure      400         {object}  utils.Envelope
// @Failure      401         {object}  utils.Envelope
// @Failure      404         {object}  utils.Envelope
// @Failure      422         {object}  utils.Envelope
// @Failure      500         {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /products/{product_id}/inventory/reserve [post]
func (h *InventoryHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "inventory.reserve")
	defer span.End()

	productID := r.PathValue("product_id")

	if err := utils.ValidateUUID(productID); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	var req dto.StockMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidRequest, "Invalid request body")
		return
	}

	if err := validators.ValidateStruct(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeValidationFailed, err.Error())
		return
	}

	input := usecases.ReserveStockInput{
		ProductID: productID,
		Quantity:  req.Quantity,
	}

	output, err := h.reserveStockUC.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		statusCode, errCode := mapInventoryError(err)
		utils.RespondErrorEnvelope(w, statusCode, errCode, err.Error())
		return
	}

	response := &dto.InventoryResponse{
		ID:                output.ID,
		ProductID:         output.ProductID,
		AvailableQuantity: output.AvailableQuantity,
		ReservedQuantity:  output.ReservedQuantity,
		PendingQuantity:   output.PendingQuantity,
		CreatedAt:         "",
		UpdatedAt:         output.UpdatedAt,
	}
	utils.RespondSuccess(w, http.StatusOK, response)
}

// Increase godoc
// @Summary      Increase stock
// @Description  Increases available stock quantity (e.g., after receiving new products)
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Param        product_id  path      string                     true  "Product ID (UUID)"
// @Param        request     body      dto.StockMovementRequest   true  "Stock increase data"
// @Success      200         {object}  utils.Envelope{data=dto.InventoryResponse}
// @Failure      400         {object}  utils.Envelope
// @Failure      401         {object}  utils.Envelope
// @Failure      404         {object}  utils.Envelope
// @Failure      500         {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /products/{product_id}/inventory/increase [post]
func (h *InventoryHandler) Increase(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "inventory.increase")
	defer span.End()

	productID := r.PathValue("product_id")

	if err := utils.ValidateUUID(productID); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	var req dto.StockMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidRequest, "Invalid request body")
		return
	}

	if err := validators.ValidateStruct(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeValidationFailed, err.Error())
		return
	}

	input := usecases.IncreaseStockInput{
		ProductID: productID,
		Quantity:  req.Quantity,
	}

	output, err := h.increaseStockUC.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		statusCode, errCode := mapInventoryError(err)
		utils.RespondErrorEnvelope(w, statusCode, errCode, err.Error())
		return
	}

	response := &dto.InventoryResponse{
		ID:                output.ID,
		ProductID:         output.ProductID,
		AvailableQuantity: output.AvailableQuantity,
		ReservedQuantity:  output.ReservedQuantity,
		PendingQuantity:   output.PendingQuantity,
		CreatedAt:         "",
		UpdatedAt:         output.UpdatedAt,
	}
	utils.RespondSuccess(w, http.StatusOK, response)
}

// CancelReserved godoc
// @Summary      Cancel reserved stock
// @Description  Cancels a stock reservation (moves from reserved back to available)
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Param        product_id  path      string                     true  "Product ID (UUID)"
// @Param        request     body      dto.StockMovementRequest   true  "Stock cancellation data"
// @Success      200         {object}  utils.Envelope{data=dto.InventoryResponse}
// @Failure      400         {object}  utils.Envelope
// @Failure      401         {object}  utils.Envelope
// @Failure      404         {object}  utils.Envelope
// @Failure      422         {object}  utils.Envelope
// @Failure      500         {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /products/{product_id}/inventory/cancel-reserved [post]
func (h *InventoryHandler) CancelReserved(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "inventory.cancel_reserved")
	defer span.End()

	productID := r.PathValue("product_id")

	if err := utils.ValidateUUID(productID); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	var req dto.StockMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidRequest, "Invalid request body")
		return
	}

	if err := validators.ValidateStruct(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeValidationFailed, err.Error())
		return
	}

	input := usecases.CancelReservedStockInput{
		ProductID: productID,
		Quantity:  req.Quantity,
	}

	output, err := h.cancelReservedStockUC.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		statusCode, errCode := mapInventoryError(err)
		utils.RespondErrorEnvelope(w, statusCode, errCode, err.Error())
		return
	}

	response := &dto.InventoryResponse{
		ID:                output.ID,
		ProductID:         output.ProductID,
		AvailableQuantity: output.AvailableQuantity,
		ReservedQuantity:  output.ReservedQuantity,
		PendingQuantity:   output.PendingQuantity,
		CreatedAt:         "",
		UpdatedAt:         output.UpdatedAt,
	}
	utils.RespondSuccess(w, http.StatusOK, response)
}

// CancelConfirmed handles POST /api/products/:product_id/inventory/cancel-confirmed
func (h *InventoryHandler) CancelConfirmed(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "inventory.cancel_confirmed")
	defer span.End()

	productID := r.PathValue("product_id")

	if err := utils.ValidateUUID(productID); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	var req dto.StockMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidRequest, "Invalid request body")
		return
	}

	if err := validators.ValidateStruct(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeValidationFailed, err.Error())
		return
	}

	input := usecases.CancelConfirmedStockInput{
		ProductID: productID,
		Quantity:  req.Quantity,
	}

	output, err := h.cancelConfirmedStockUC.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		statusCode, errCode := mapInventoryError(err)
		utils.RespondErrorEnvelope(w, statusCode, errCode, err.Error())
		return
	}

	response := &dto.InventoryResponse{
		ID:                output.ID,
		ProductID:         output.ProductID,
		AvailableQuantity: output.AvailableQuantity,
		ReservedQuantity:  output.ReservedQuantity,
		PendingQuantity:   output.PendingQuantity,
		CreatedAt:         "",
		UpdatedAt:         output.UpdatedAt,
	}
	utils.RespondSuccess(w, http.StatusOK, response)
}

// mapInventoryError maps inventory domain errors to HTTP status codes and error codes
func mapInventoryError(err error) (int, string) {
	switch err {
	case inventory.ErrInventoryNotFound:
		return http.StatusNotFound, utils.ErrCodeNotFound
	case inventory.ErrProductAlreadyHasInventory:
		return http.StatusConflict, utils.ErrCodeDuplicateResource
	case inventory.ErrInvalidInventoryID, inventory.ErrInvalidProductID, inventory.ErrInvalidQuantity:
		return http.StatusBadRequest, utils.ErrCodeValidationFailed
	case inventory.ErrInsufficientStock, inventory.ErrInsufficientAvailable, inventory.ErrInsufficientReserved:
		return http.StatusUnprocessableEntity, utils.ErrCodeValidationFailed
	default:
		return http.StatusInternalServerError, utils.ErrCodeInternalError
	}
}
