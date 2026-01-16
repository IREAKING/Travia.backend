package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ImageURLs(urls []string) []string {
	return urls
}

// UploadImagesForTour uploads multiple images for tour creation and returns URLs
// @Summary Upload multiple images for tour creation
// @Description Upload multiple images for tour creation and return URLs for tour creation
// @Tags Storage
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Image files to upload (JPEG, PNG, GIF, WebP, BMP)"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /storage/upload-tour-images [post]
func (s *Server) UploadImagesForTour(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	supabaseURL := s.config.SupabaseConfig.URL
	supabaseKey := s.config.SupabaseConfig.Key
	bucketName := s.config.SupabaseConfig.Bucket

	// Get form data
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form: " + err.Error()})
		return
	}
	defer form.RemoveAll()

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	// Use tour-images folder by default
	folderPath := "tours/upload"
	if bucketName == "" {
		bucketName = "images"
	}

	if supabaseURL == "" || supabaseKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SUPABASE_URL and SUPABASE_KEY environment variables are required"})
		return
	}

	var uploadedImages []gin.H
	var errors []string

	// Process each file
	for i, file := range files {
		// Check file type - only allow images
		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/jpg":  true,
			"image/png":  true,
			"image/gif":  true,
			"image/webp": true,
			"image/bmp":  true,
		}

		if !allowedTypes[file.Header.Get("Content-Type")] {
			errors = append(errors, fmt.Sprintf("File %d (%s): invalid file type. Only image files (JPEG, PNG, GIF, WebP, BMP) are allowed", i+1, file.Filename))
			continue
		}

		// Open the file
		fileReader, err := file.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("File %d (%s): failed to open file: %s", i+1, file.Filename, err.Error()))
			continue
		}

		// Read the file content
		fileContent, err := io.ReadAll(fileReader)
		fileReader.Close()
		if err != nil {
			errors = append(errors, fmt.Sprintf("File %d (%s): failed to read file: %s", i+1, file.Filename, err.Error()))
			continue
		}

		// Generate unique filename with timestamp
		ext := filepath.Ext(file.Filename)
		nameWithoutExt := file.Filename[:len(file.Filename)-len(ext)]
		timestamp := time.Now().UnixNano() // Use Nanosecond for better uniqueness
		uniqueFileName := fmt.Sprintf("%s_%d_%d%s", nameWithoutExt, timestamp, i, ext)

		// Create full path
		fullPath := fmt.Sprintf("%s/%s", folderPath, uniqueFileName)

		// Prepare the upload URL
		uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, fullPath)

		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, bytes.NewReader(fileContent))
		if err != nil {
			errors = append(errors, fmt.Sprintf("File %d (%s): failed to create request: %s", i+1, file.Filename, err.Error()))
			continue
		}

		// Set headers
		req.Header.Set("Authorization", "Bearer "+supabaseKey)
		req.Header.Set("Content-Type", file.Header.Get("Content-Type"))
		req.Header.Set("Cache-Control", "max-age=3600")

		// Make the request
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("File %d (%s): failed to upload file: %s", i+1, file.Filename, err.Error()))
			continue
		}

		// Check response status
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			errors = append(errors, fmt.Sprintf("File %d (%s): upload failed with status %d: %s", i+1, file.Filename, resp.StatusCode, string(body)))
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Add to successful uploads
		publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, fullPath)

		// Create image object for tour creation
		imageObj := gin.H{
			"link":            publicURL,
			"mo_ta_alt":       fmt.Sprintf("Tour image %d", i+1),
			"la_anh_chinh":    i == 0, // First image is main image
			"thu_tu_hien_thi": i + 1,
		}

		uploadedImages = append(uploadedImages, imageObj)
	}

	// Prepare response
	response := gin.H{
		"successful_uploads": len(uploadedImages),
		"total_files":        len(files),
		"images":             uploadedImages,
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["message"] = fmt.Sprintf("Uploaded %d out of %d files successfully", len(uploadedImages), len(files))
		c.JSON(http.StatusPartialContent, response)
	} else {
		response["message"] = "All files uploaded successfully"
		c.JSON(http.StatusOK, response)
	}
}

// UploadImageWithPath uploads an image with a specific path in the bucket
// @Summary Upload image with a specific path in the bucket
// @Description Upload image with a specific path in the bucket
// @Tags Storage
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file to upload (JPEG, PNG, GIF, WebP, BMP)"
// @Param folder_path formData string true "Folder path"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /storage/upload [post]
func (s *Server) UploadImage(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	supabaseURL := s.config.SupabaseConfig.URL
	supabaseKey := s.config.SupabaseConfig.Key
	bucketName := s.config.SupabaseConfig.Bucket

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	fileName := file.Filename
	folderPath := c.PostForm("folder_path")

	// Check file type - only allow images
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"image/bmp":  true,
	}

	if !allowedTypes[file.Header.Get("Content-Type")] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only image files (JPEG, PNG, GIF, WebP, BMP) are allowed"})
		return
	}

	if bucketName == "" {
		bucketName = "images"
	}

	if supabaseURL == "" || supabaseKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SUPABASE_URL and SUPABASE_KEY environment variables are required"})
		return
	}

	// Open the file
	fileReader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file: " + err.Error()})
		return
	}
	defer fileReader.Close()

	// Read the file content
	fileContent, err := io.ReadAll(fileReader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file: " + err.Error()})
		return
	}

	// Generate unique filename with timestamp
	ext := filepath.Ext(fileName)
	nameWithoutExt := fileName[:len(fileName)-len(ext)]
	timestamp := time.Now().Unix()
	uniqueFileName := fmt.Sprintf("%s_%d%s", nameWithoutExt, timestamp, ext)

	// Create full path
	fullPath := fmt.Sprintf("%s/%s", folderPath, uniqueFileName)

	// Prepare the upload URL
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, fullPath)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, bytes.NewReader(fileContent))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request: " + err.Error()})
		return
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Cache-Control", "max-age=3600")

	// Make the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("upload failed with status %d: %s", resp.StatusCode, string(body))})
		return
	}

	// Return the public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, fullPath)
	c.JSON(http.StatusOK, gin.H{"url": publicURL})
}

// UploadMultipleImages uploads multiple images with a specific path in the bucket
// @Summary Upload multiple images with a specific path in the bucket
// @Description Upload multiple images with a specific path in the bucket. You can select multiple files in the Swagger UI by holding Ctrl/Cmd while clicking files.
// @Tags Storage
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Image files to upload (JPEG, PNG, GIF, WebP, BMP) - select multiple files by holding Ctrl/Cmd"
// @Param folder_path formData string true "Folder path"
// @Success 200 {object} gin.H
// @Success 206 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /storage/upload-multiple [post]
func (s *Server) UploadMultipleImages(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	supabaseURL := s.config.SupabaseConfig.URL
	supabaseKey := s.config.SupabaseConfig.Key
	bucketName := s.config.SupabaseConfig.Bucket

	// Get form data
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form: " + err.Error()})
		return
	}
	defer form.RemoveAll()

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	folderPath := c.PostForm("folder_path")
	if folderPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "folder_path is required"})
		return
	}

	if bucketName == "" {
		bucketName = "images"
	}

	if supabaseURL == "" || supabaseKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SUPABASE_URL and SUPABASE_KEY environment variables are required"})
		return
	}

	var uploadedURLs []string
	var errors []string

	// Process each file
	for i, file := range files {
		// Check file type - only allow images
		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/jpg":  true,
			"image/png":  true,
			"image/gif":  true,
			"image/webp": true,
			"image/bmp":  true,
		}

		if !allowedTypes[file.Header.Get("Content-Type")] {
			errors = append(errors, fmt.Sprintf("File %d (%s): invalid file type. Only image files (JPEG, PNG, GIF, WebP, BMP) are allowed", i+1, file.Filename))
			continue
		}

		// Open the file
		fileReader, err := file.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("File %d (%s): failed to open file: %s", i+1, file.Filename, err.Error()))
			continue
		}

		// Read the file content
		fileContent, err := io.ReadAll(fileReader)
		fileReader.Close()
		if err != nil {
			errors = append(errors, fmt.Sprintf("File %d (%s): failed to read file: %s", i+1, file.Filename, err.Error()))
			continue
		}

		// Generate unique filename with timestamp
		ext := filepath.Ext(file.Filename)
		nameWithoutExt := file.Filename[:len(file.Filename)-len(ext)]
		timestamp := time.Now().UnixNano() // Use Nanosecond for better uniqueness
		uniqueFileName := fmt.Sprintf("%s_%d_%d%s", nameWithoutExt, timestamp, i, ext)

		// Create full path
		fullPath := fmt.Sprintf("%s/%s", folderPath, uniqueFileName)

		// Prepare the upload URL
		uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, fullPath)

		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, bytes.NewReader(fileContent))
		if err != nil {
			errors = append(errors, fmt.Sprintf("File %d (%s): failed to create request: %s", i+1, file.Filename, err.Error()))
			continue
		}

		// Set headers
		req.Header.Set("Authorization", "Bearer "+supabaseKey)
		req.Header.Set("Content-Type", file.Header.Get("Content-Type"))
		req.Header.Set("Cache-Control", "max-age=3600")

		// Make the request
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("File %d (%s): failed to upload file: %s", i+1, file.Filename, err.Error()))
			continue
		}

		// Check response status
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			errors = append(errors, fmt.Sprintf("File %d (%s): upload failed with status %d: %s", i+1, file.Filename, resp.StatusCode, string(body)))
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Add to successful uploads
		publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, fullPath)
		uploadedURLs = append(uploadedURLs, publicURL)
	}
	ImageURLs(uploadedURLs)

	// Prepare response
	response := gin.H{
		"successful_uploads": len(uploadedURLs),
		"total_files":        len(files),
		"urls":               uploadedURLs,
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["message"] = fmt.Sprintf("Uploaded %d out of %d files successfully", len(uploadedURLs), len(files))
		c.JSON(http.StatusPartialContent, response)
	} else {
		response["message"] = "All files uploaded successfully"
		c.JSON(http.StatusOK, gin.H{
			"data":    ImageURLs(uploadedURLs),
			"message": "All files uploaded successfully",
		})
	}
}

// DeleteImage deletes an image from Supabase storage
func (s *Server) DeleteImage(ctx context.Context, fileName string) error {
	supabaseURL := s.config.SupabaseConfig.URL
	supabaseKey := s.config.SupabaseConfig.Key
	bucketName := s.config.SupabaseConfig.Bucket

	if bucketName == "" {
		bucketName = "images"
	}

	if supabaseURL == "" || supabaseKey == "" {
		return fmt.Errorf("SUPABASE_URL and SUPABASE_KEY environment variables are required")
	}

	// Prepare the delete URL
	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, fileName)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+supabaseKey)

	// Make the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// uploadFileToSupabase uploads a file to Supabase storage and returns the public URL
func (s *Server) uploadFileToSupabase(ctx context.Context, fileContent []byte, fileName, bucketName, folderPath, contentType string) (string, error) {
	supabaseURL := s.config.SupabaseConfig.URL
	supabaseKey := s.config.SupabaseConfig.Key

	if supabaseURL == "" || supabaseKey == "" {
		return "", fmt.Errorf("SUPABASE_URL and SUPABASE_KEY environment variables are required")
	}

	// Generate unique filename with timestamp
	ext := filepath.Ext(fileName)
	nameWithoutExt := fileName[:len(fileName)-len(ext)]
	timestamp := time.Now().UnixNano()
	uniqueFileName := fmt.Sprintf("%s_%d%s", nameWithoutExt, timestamp, ext)

	// Create full path
	var fullPath string
	if folderPath != "" {
		fullPath = fmt.Sprintf("%s/%s", folderPath, uniqueFileName)
	} else {
		fullPath = uniqueFileName
	}

	// Prepare the upload URL
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, fullPath)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, bytes.NewReader(fileContent))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Cache-Control", "max-age=3600")

	// Make the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Return the public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, fullPath)
	return publicURL, nil
}

// GetSignedPDF lấy URL signed của file PDF
// @Sumary lấy URL signed của file PDF
// @Description lấy URL signed của file PDF
// @Tags Storage
// @Accept json
// @Produce json
// @Param filename path string true "Tên file"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Security ApiKeyAuth
// @Router /storage/get-signed-pdf/{filename} [get]
func (s *Server) GetSignedPDF(c *gin.Context) {
	// 1️⃣ Lấy filename từ route
	fileName := c.Param("filename")
	log.Println("[GetSignedPDF] filename:", fileName)

	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "filename parameter is required",
		})
		return
	}

	// 2️⃣ Load config
	baseURL := strings.TrimRight(s.config.SupabaseConfig.URL, "/")
	serviceKey := s.config.SupabaseConfig.Key
	bucket := "giay_phep_kinh_doanh"

	if baseURL == "" || serviceKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "SUPABASE_URL or SUPABASE_SERVICE_ROLE_KEY is missing",
		})
		return
	}

	// 3️⃣ Gọi Supabase tạo signed URL
	signAPI := fmt.Sprintf(
		"%s/storage/v1/object/sign/%s/%s",
		baseURL,
		bucket,
		fileName,
	)

	payload := strings.NewReader(`{"expiresIn":300}`)

	req, err := http.NewRequestWithContext(
		c.Request.Context(),
		http.MethodPost,
		signAPI,
		payload,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create request: " + err.Error(),
		})
		return
	}

	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to request Supabase: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 4️⃣ Check status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("[GetSignedPDF] Supabase error:", string(body))
		c.JSON(resp.StatusCode, gin.H{
			"error": string(body),
		})
		return
	}

	// 5️⃣ Decode response
	var res struct {
		SignedURL string `json:"signedURL"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to decode Supabase response: " + err.Error(),
		})
		return
	}

	if res.SignedURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Supabase returned empty signedURL",
		})
		return
	}

	// 6️⃣ Normalize signed URL (FIX LỖI /object)
	signedPath := res.SignedURL
	if strings.HasPrefix(signedPath, "/object/") {
		signedPath = "/storage/v1" + signedPath
	}

	fullURL := baseURL + signedPath

	// 7️⃣ Trả về frontend
	c.JSON(http.StatusOK, gin.H{
		"url": fullURL,
	})
}

// UploadMultipleImagesTest creates a simple HTML form for testing multiple file uploads
// @Summary Test multiple file upload with HTML form
// @Description Returns an HTML form for testing multiple file uploads
// @Tags Storage
// @Produce html
// @Success 200 {string} string "HTML form"
// @Router /storage/upload-multiple-test [get]
func (s *Server) UploadMultipleImagesTest(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Multiple File Upload Test</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .form-group { margin: 20px 0; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input[type="file"] { margin-bottom: 10px; }
        input[type="text"] { width: 300px; padding: 8px; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #0056b3; }
        .result { margin-top: 20px; padding: 15px; background: #f8f9fa; border-radius: 4px; }
    </style>
</head>
<body>
    <h2>Multiple File Upload Test</h2>
    <form id="uploadForm" enctype="multipart/form-data">
        <div class="form-group">
            <label for="files">Select Multiple Files:</label>
            <input type="file" id="files" name="files" multiple accept="image/jpeg,image/jpg,image/png,image/gif,image/webp,image/bmp">
            <small>Hold Ctrl/Cmd to select multiple files</small>
        </div>
        <div class="form-group">
            <label for="folder_path">Folder Path:</label>
            <input type="text" id="folder_path" name="folder_path" value="test" placeholder="e.g., tours, destinations">
        </div>
        <button type="submit">Upload Files</button>
    </form>
    
    <div id="result" class="result" style="display: none;"></div>

    <script>
        document.getElementById('uploadForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const formData = new FormData();
            const files = document.getElementById('files').files;
            const folderPath = document.getElementById('folder_path').value;
            
            if (files.length === 0) {
                alert('Please select at least one file');
                return;
            }
            
            if (!folderPath) {
                alert('Please enter a folder path');
                return;
            }
            
            // Add files to form data
            for (let i = 0; i < files.length; i++) {
                formData.append('files', files[i]);
            }
            formData.append('folder_path', folderPath);
            
            const resultDiv = document.getElementById('result');
            resultDiv.style.display = 'block';
            resultDiv.innerHTML = 'Uploading...';
            
            try {
                const response = await fetch('/api/storage/upload-multiple', {
                    method: 'POST',
                    body: formData
                });
                
                const data = await response.json();
                
                let html = '<h3>Upload Result:</h3>';
                html += '<p><strong>Status:</strong> ' + response.status + '</p>';
                html += '<p><strong>Message:</strong> ' + data.message + '</p>';
                html += '<p><strong>Successful Uploads:</strong> ' + data.successful_uploads + '/' + data.total_files + '</p>';
                
                if (data.urls && data.urls.length > 0) {
                    html += '<h4>Uploaded URLs:</h4><ul>';
                    data.urls.forEach(url => {
                        html += '<li><a href="' + url + '" target="_blank">' + url + '</a></li>';
                    });
                    html += '</ul>';
                }
                
                if (data.errors && data.errors.length > 0) {
                    html += '<h4>Errors:</h4><ul>';
                    data.errors.forEach(error => {
                        html += '<li style="color: red;">' + error + '</li>';
                    });
                    html += '</ul>';
                }
                
                resultDiv.innerHTML = html;
            } catch (error) {
                resultDiv.innerHTML = '<h3>Error:</h3><p style="color: red;">' + error.message + '</p>';
            }
        });
    </script>
</body>
</html>`
	c.Header("Content-Type", "text/html")
	c.String(200, html)
}

type SupabaseFile struct {
	Name           string `json:"name"`
	ID             string `json:"id,omitempty"`
	BucketID       string `json:"bucket_id,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
	LastAccessedAt string `json:"last_accessed_at,omitempty"`
	Metadata       struct {
		Size int64  `json:"size"`
		Mime string `json:"mimetype"`
	} `json:"metadata"`
}

// ListImages lists all images in the bucket
// @Summary List images in the bucket
// @Description List all images in the bucket
// @Tags Storage
// @Accept json
// @Produce json
// @Param folder_path query string false "Folder path"
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Success 200 {array} SupabaseFile
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /storage/list [post]
func (s *Server) ListImages(c *gin.Context) {
	folderPath := c.Query("folder_path")
	limit := c.Query("limit")
	if limit == "" {
		limit = "1000"
	}
	offset := c.Query("offset")
	if offset == "" {
		offset = "0"
	}
	supabaseURL := s.config.SupabaseConfig.URL
	supabaseKey := s.config.SupabaseConfig.Key
	bucketName := s.config.SupabaseConfig.Bucket

	if bucketName == "" {
		bucketName = "images"
	}

	if supabaseURL == "" || supabaseKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SUPABASE_URL and SUPABASE_KEY are required"})
		return
	}

	// body cho API list
	body := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}
	if folderPath != "" {
		// Supabase yêu cầu prefix kèm dấu `/` nếu muốn lấy trong folder
		if !strings.HasSuffix(folderPath, "/") {
			folderPath += "/"
		}
		body["prefix"] = folderPath
	}

	bodyBytes, _ := json.Marshal(body)

	// Tạo request POST
	listURL := fmt.Sprintf("%s/storage/v1/object/list/%s", supabaseURL, bucketName)
	req, err := http.NewRequestWithContext(c.Request.Context(), "POST", listURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request: " + err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("apikey", supabaseKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list files: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("list failed with status %d: %s", resp.StatusCode, string(body))})
		return
	}

	var _files []SupabaseFile
	if err := json.NewDecoder(resp.Body).Decode(&_files); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response: " + err.Error()})
		return
	}

	// Ghép link public nếu bucket public
	var files []SupabaseFile
	for _, f := range _files {
		files = append(files, SupabaseFile{
			Name: fmt.Sprintf("%s/storage/v1/object/public/%s/%s%s",
				supabaseURL, bucketName, folderPath, f.Name),
			ID:             f.ID,
			BucketID:       f.BucketID,
			CreatedAt:      f.CreatedAt,
			UpdatedAt:      f.UpdatedAt,
			LastAccessedAt: f.LastAccessedAt,
			Metadata:       f.Metadata,
		})
	}
	c.JSON(http.StatusOK, gin.H{"files": files})
}

// CreateBucket creates a new bucket in Supabase storage
func CreateBucket(ctx context.Context, bucketName string, isPublic bool) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY_ROLE")

	if supabaseURL == "" || supabaseKey == "" {
		return fmt.Errorf("SUPABASE_URL and SUPABASE_KEY_ROLE environment variables are required")
	}

	// Prepare the create bucket URL
	createURL := fmt.Sprintf("%s/storage/v1/bucket", supabaseURL)

	// Prepare request body
	requestBody := map[string]interface{}{
		"id":              bucketName,
		"name":            bucketName,
		"public":          isPublic,
		"file_size_limit": 52428800, // 50MB
		"allowed_mime_types": []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
			"image/bmp",
		},
	}

	// Convert to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", createURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create bucket failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// EnsureBucketExists checks if bucket exists and creates it if not
func EnsureBucketExists(ctx context.Context, bucketName string) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY_ROLE")

	if supabaseURL == "" || supabaseKey == "" {
		return fmt.Errorf("SUPABASE_URL and SUPABASE_KEY_ROLE environment variables are required")
	}

	// Check if bucket exists by trying to list it
	listURL := fmt.Sprintf("%s/storage/v1/bucket/%s", supabaseURL, bucketName)

	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+supabaseKey)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}
	defer resp.Body.Close()

	// If bucket exists, return nil
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// If bucket doesn't exist (404), create it
	if resp.StatusCode == http.StatusNotFound {
		return CreateBucket(ctx, bucketName, true) // Create as public bucket
	}

	// Other errors
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("check bucket failed with status %d: %s", resp.StatusCode, string(body))
}
