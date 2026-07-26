package handlers

import (
	"errors"
	"net/http"
	"strings"

	"arturgudiev/dashboard/services"

	"github.com/gin-gonic/gin"
)

// GetFilesConfig handles GET /files/config
// @Summary      Get files storage config
// @Description  Returns whether files are stored encrypted and the configured base directory
// @Tags         files
// @Produce      json
// @Success      200  {object}  services.FilesConfig
// @Failure      403  {object}  map[string]string
// @Router       /files/config [get]
func (h *Handler) GetFilesConfig(c *gin.Context) {
	c.JSON(http.StatusOK, h.App.FilesService.Config())
}

// ListFiles handles GET /files
// @Summary      List files
// @Description  Lists all files and directories under the configured files directory (logical relative paths; when FILES_ENCRYPTED=true the trailing .bin suffix is stripped)
// @Tags         files
// @Produce      json
// @Success      200  {array}   services.FileInfo
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /files [get]
func (h *Handler) ListFiles(c *gin.Context) {
	files, err := h.App.FilesService.ListFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, files)
}

// GetFile handles GET /files/content/*filepath
// @Summary      Get file by relative path
// @Description  Returns file contents for the given logical relative path (e.g. 1.txt or sub/dir/file.txt). When FILES_ENCRYPTED=true reads path.bin on disk and returns opaque ciphertext (no server-side decrypt).
// @Tags         files
// @Produce      application/octet-stream
// @Param        filepath  path  string  true  "Relative file path"  example(1.txt)
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /files/content/{filepath} [get]
func (h *Handler) GetFile(c *gin.Context) {
	relPath := strings.TrimPrefix(c.Param("filepath"), "/")
	data, name, err := h.App.FilesService.GetFile(relPath)
	if err != nil {
		writeFilesError(c, err)
		return
	}

	c.Header("X-Files-Encrypted", boolHeader(h.App.FilesService.IsEncrypted()))
	c.Header("Content-Disposition", `attachment; filename="`+name+`"`)
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// PutFile handles PUT /files/content/*filepath
// @Summary      Upload or replace a file
// @Description  Writes request body to the given logical relative path. When FILES_ENCRYPTED=true stores as path.bin; client should send already-encrypted bytes.
// @Tags         files
// @Accept       application/octet-stream
// @Produce      json
// @Param        filepath  path  string  true  "Relative file path"  example(1.txt)
// @Param        body      body  string  true  "Raw file bytes"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /files/content/{filepath} [put]
func (h *Handler) PutFile(c *gin.Context) {
	relPath := strings.TrimPrefix(c.Param("filepath"), "/")
	if err := h.App.FilesService.PutFile(relPath, c.Request.Body); err != nil {
		writeFilesError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"path":      relPath,
		"encrypted": h.App.FilesService.IsEncrypted(),
	})
}

// DeleteFile handles DELETE /files/content/*filepath
// @Summary      Delete a file
// @Description  Deletes a file by logical relative path (when FILES_ENCRYPTED=true deletes path.bin)
// @Tags         files
// @Produce      json
// @Param        filepath  path  string  true  "Relative file path"  example(1.txt)
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /files/content/{filepath} [delete]
func (h *Handler) DeleteFile(c *gin.Context) {
	relPath := strings.TrimPrefix(c.Param("filepath"), "/")
	if err := h.App.FilesService.DeleteFile(relPath); err != nil {
		writeFilesError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"path": relPath, "deleted": true})
}

// GetFileParentsPath handles GET /files/parents-path/*filepath
// @Summary      Get parents path for a file
// @Description  Resolves the owning container from a logical relative file path and returns its parent container descriptions (leaf→root), same shape as POST /parents-path
// @Tags         files
// @Produce      json
// @Param        filepath  path  string  true  "Relative file path"  example(tasks/12_foo/note.md)
// @Success      200  {array}   string
// @Failure      403  {object}  map[string]string
// @Router       /files/parents-path/{filepath} [get]
func (h *Handler) GetFileParentsPath(c *gin.Context) {
	relPath := strings.TrimPrefix(c.Param("filepath"), "/")
	containerType, id, ok := services.OwningContainerFromFilesRelPath(relPath)
	if !ok {
		c.JSON(http.StatusOK, []string{})
		return
	}

	parentsPath := h.App.ContainerService.GetParentsPathDescriptions(c.Request.Context(), containerType, id)
	if parentsPath == nil {
		parentsPath = []string{}
	}
	c.JSON(http.StatusOK, parentsPath)
}

func writeFilesError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidFilePath):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrFileNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrNotAFile):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func boolHeader(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
