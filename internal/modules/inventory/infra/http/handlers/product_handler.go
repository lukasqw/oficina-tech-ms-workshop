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

// ProductHandler gerencia as requisições HTTP relacionadas a produtos
type ProductHandler struct {
	createProductUC  *usecases.CreateProductUseCase
	getProductByIDUC *usecases.GetProductByIDUseCase
	getAllProductsUC *usecases.GetAllProductsUseCase
	updateProductUC  *usecases.UpdateProductUseCase
	deleteProductUC  *usecases.DeleteProductUseCase
}

// NewProductHandler cria uma nova instância do handler com todas as dependências
func NewProductHandler(productRepo product.Repository, inventoryRepo inventory.Repository) *ProductHandler {
	return &ProductHandler{
		createProductUC:  usecases.NewCreateProductUseCase(productRepo, inventoryRepo),
		getProductByIDUC: usecases.NewGetProductByIDUseCase(productRepo),
		getAllProductsUC: usecases.NewGetAllProductsUseCase(productRepo),
		updateProductUC:  usecases.NewUpdateProductUseCase(productRepo),
		deleteProductUC:  usecases.NewDeleteProductUseCase(productRepo, inventoryRepo),
	}
}

// CreateProduct godoc
// @Summary      Create a new product
// @Description  Creates a new product in the inventory system
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateProductRequest  true  "Product creation data"
// @Success      201      {object}  utils.Envelope{data=dto.ProductResponse}
// @Failure      400      {object}  utils.Envelope
// @Failure      401      {object}  utils.Envelope
// @Failure      409      {object}  utils.Envelope
// @Failure      500      {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /products [post]
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "product.create")
	defer span.End()

	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidRequest, "Invalid request body")
		return
	}

	if err := validators.ValidateStruct(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeValidationFailed, err.Error())
		return
	}

	input := usecases.CreateProductInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ProductType: req.ProductType,
	}

	output, err := h.createProductUC.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		mapping := utils.MapDomainError(err)
		utils.RespondErrorEnvelope(w, mapping.StatusCode, mapping.Code, err.Error())
		return
	}

	response := dto.ProductResponse{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		Price:       output.Price,
		ProductType: output.ProductType,
		CreatedAt:   output.CreatedAt,
		UpdatedAt:   output.UpdatedAt,
	}

	utils.RespondSuccess(w, http.StatusCreated, response)
}

// GetAllProducts manipula requisições GET para listar todos os produtos
func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "product.get_all")
	defer span.End()

	output, err := h.getAllProductsUC.Execute(ctx)
	if err != nil {
		span.RecordError(err)
		mapping := utils.MapDomainError(err)
		utils.RespondErrorEnvelope(w, mapping.StatusCode, mapping.Code, err.Error())
		return
	}

	response := make([]dto.ProductResponse, len(output.Products))
	for i, p := range output.Products {
		response[i] = dto.ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			ProductType: p.ProductType,
			CreatedAt:   utils.FormatTimeRFC3339(p.CreatedAt),
			UpdatedAt:   utils.FormatTimeRFC3339(p.UpdatedAt),
		}
	}

	utils.RespondSuccess(w, http.StatusOK, response)
}

// GetProductByID godoc
// @Summary      Get product by ID
// @Description  Retrieves a product by its unique identifier
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Product ID (UUID)"
// @Success      200  {object}  utils.Envelope{data=dto.ProductResponse}
// @Failure      400  {object}  utils.Envelope
// @Failure      401  {object}  utils.Envelope
// @Failure      404  {object}  utils.Envelope
// @Failure      500  {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /products/{id} [get]
func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "product.get_by_id")
	defer span.End()

	id := r.PathValue("id")

	// Validar UUID
	if err := utils.ValidateUUID(id); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	input := usecases.GetProductByIDInput{
		ID: id,
	}

	output, err := h.getProductByIDUC.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		mapping := utils.MapDomainError(err)
		utils.RespondErrorEnvelope(w, mapping.StatusCode, mapping.Code, err.Error())
		return
	}

	response := dto.ProductResponse{
		ID:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		Price:       output.Price,
		ProductType: output.ProductType,
		CreatedAt:   utils.FormatTimeRFC3339(output.CreatedAt),
		UpdatedAt:   utils.FormatTimeRFC3339(output.UpdatedAt),
	}

	utils.RespondSuccess(w, http.StatusOK, response)
}

// UpdateProduct godoc
// @Summary      Update product
// @Description  Updates an existing product's information
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      string                    true  "Product ID (UUID)"
// @Param        request  body      dto.UpdateProductRequest  true  "Product update data"
// @Success      200      {object}  utils.Envelope{data=object}
// @Failure      400      {object}  utils.Envelope
// @Failure      401      {object}  utils.Envelope
// @Failure      404      {object}  utils.Envelope
// @Failure      500      {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /products/{id} [put]
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "product.update")
	defer span.End()

	id := r.PathValue("id")

	// Validar UUID
	if err := utils.ValidateUUID(id); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidRequest, "Invalid request body")
		return
	}

	if err := validators.ValidateStruct(&req); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeValidationFailed, err.Error())
		return
	}

	input := usecases.UpdateProductInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ProductType: req.ProductType,
	}

	if err := h.updateProductUC.Execute(ctx, input); err != nil {
		span.RecordError(err)
		mapping := utils.MapDomainError(err)
		utils.RespondErrorEnvelope(w, mapping.StatusCode, mapping.Code, err.Error())
		return
	}

	utils.RespondSuccess(w, http.StatusOK, map[string]string{"message": "Product updated successfully"})
}

// DeleteProduct godoc
// @Summary      Delete product
// @Description  Soft deletes a product from the system
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Product ID (UUID)"
// @Success      200  {object}  utils.Envelope{data=object}
// @Failure      400  {object}  utils.Envelope
// @Failure      401  {object}  utils.Envelope
// @Failure      404  {object}  utils.Envelope
// @Failure      500  {object}  utils.Envelope
// @Security     BearerAuth
// @Router       /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.SpanHandler(r.Context(), "product.delete")
	defer span.End()

	id := r.PathValue("id")

	// Validar UUID
	if err := utils.ValidateUUID(id); err != nil {
		utils.RespondErrorEnvelope(w, http.StatusBadRequest, utils.ErrCodeInvalidUUID, "Invalid product ID format")
		return
	}

	input := usecases.DeleteProductInput{
		ID: id,
	}

	if err := h.deleteProductUC.Execute(ctx, input); err != nil {
		span.RecordError(err)
		mapping := utils.MapDomainError(err)
		utils.RespondErrorEnvelope(w, mapping.StatusCode, mapping.Code, err.Error())
		return
	}

	utils.RespondSuccess(w, http.StatusOK, map[string]string{"message": "Product deleted successfully"})
}
