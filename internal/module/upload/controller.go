package upload

import (
	"boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type FileHandler struct {
	useCase FileUseCase
}

// NewFileHandler creates a new file handler
func NewFileHandler(useCase FileUseCase) *FileHandler {
	return &FileHandler{useCase: useCase}
}

// Upload godoc
// @Summary Upload a file
// @Description Upload a file to the server
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param type query string false "Upload type: image, document, or default" Enums(image, document, default)
// @Success 201 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Security BearerAuth
// @Router /files/upload [post]
func (h *FileHandler) Upload(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)

	file, err := c.FormFile("file")
	if err != nil {
		return errors.New(errors.ValidationFailed)
	}

	// Get upload options based on type
	uploadType := c.Query("type", "default")
	var opts *UploadOptions
	switch uploadType {
	case "image":
		opts = ImageUploadOptions()
	case "document":
		opts = DocumentUploadOptions()
	default:
		opts = DefaultUploadOptions()
	}

	upload, err := h.useCase.Upload(c.Context(), &userID, file, opts)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.CreateSuccessResponse(
		c,
		"File berhasil diupload",
		"File uploaded successfully",
		upload,
		fiber.StatusCreated,
	))
}

// GetByID godoc
// @Summary Get file by ID
// @Description Get file details by ID
// @Tags Files
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Security BearerAuth
// @Router /files/{id} [get]
func (h *FileHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errors.New(errors.ValidationFailed)
	}

	file, err := h.useCase.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.CreateSuccessResponse(
		c,
		"File ditemukan",
		"File found",
		file,
	))
}

// GetMyFiles godoc
// @Summary Get my files
// @Description Get all files uploaded by the current user
// @Tags Files
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Security BearerAuth
// @Router /files [get]
func (h *FileHandler) GetMyFiles(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	files, total, err := h.useCase.GetByUserID(c.Context(), userID, page, pageSize)
	if err != nil {
		return err
	}

	totalPage := (total + int64(pageSize) - 1) / int64(pageSize)

	return c.Status(fiber.StatusOK).JSON(response.CreatePaginatedResponse(
		c,
		"Berhasil mengambil daftar file",
		"Successfully retrieved files",
		files,
		&response.MetaResponse{
			Page:      int64(page),
			PageSize:  int64(pageSize),
			Total:     total,
			TotalPage: totalPage,
			IsNext:    int64(page) < totalPage,
			IsBack:    page > 1,
		},
	))
}

// Delete godoc
// @Summary Delete a file
// @Description Delete a file by ID (soft delete by default)
// @Tags Files
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Param hard query bool false "Hard delete (permanently remove file)"
// @Success 200 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Security BearerAuth
// @Router /files/{id} [delete]
func (h *FileHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errors.New(errors.ValidationFailed)
	}

	hardDelete := c.QueryBool("hard", false)

	if err := h.useCase.Delete(c.Context(), id, hardDelete); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.CreateSuccessResponse(
		c,
		"File berhasil dihapus",
		"File deleted successfully",
		nil,
	))
}

// GetURL godoc
// @Summary Get file URL
// @Description Get the URL for a file
// @Tags Files
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Security BearerAuth
// @Router /files/{id}/url [get]
func (h *FileHandler) GetURL(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errors.New(errors.ValidationFailed)
	}

	url, err := h.useCase.GetURL(c.Context(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.CreateSuccessResponse(
		c,
		"URL file",
		"File URL",
		fiber.Map{
			"url": url,
		},
	))
}
