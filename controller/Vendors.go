package controller

import (
	utils "backend-project/Utils"
	"backend-project/models"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var (
	vendor_columns = []string{
		"id",
		"name",
		"img",
		"description",
		"created_at",
		"updated_at",
		fmt.Sprintf("CASE WHEN NULLIF(img, '') IS NOT NULL THEN FORMAT('%s/%%s', img) ELSE NULL END AS img", Domain),
	}
)

func IndexVendorHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	enableCors(&w)
	var vendors []models.Vendors

	query, args, err := QB.Select(strings.Join(vendor_columns, ", ")).
		From("vendors").
		ToSql()

	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := db.Select(&vendors, query, args...); err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JasonResponseHandler(w, http.StatusInternalServerError, vendors)
}

func ShowVendorHandler(w http.ResponseWriter, r *http.Request) {
	var vendor models.Vendors
	id := r.PathValue("id")
	query, args, err := QB.Select(strings.Join(vendor_columns, ", ")).
		From("vendors").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := db.Get(&vendor, query, args...); err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JasonResponseHandler(w, http.StatusOK, vendor)
}

func SaveVendorHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	enableCors(&w)

	if r.Method != "POST" {
		utils.HandleErrors(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	var vendor models.Vendors

	vendor.Name = r.FormValue("name")
	vendor.Description = r.FormValue("description")
	fmt.Println(vendor.Name, vendor.Description)
	vendor.ID = uuid.New()

	if vendor.Name == "" || vendor.Description == "" {
		utils.HandleErrors(w, http.StatusBadRequest, "Name and description are required")
		return
	}

	file, fileHeader, err := r.FormFile("img")
	if err != nil && err != http.ErrMissingFile {
		utils.HandleErrors(w, http.StatusBadRequest, "Invalid file")
		return
	} else if err == nil {
		defer file.Close()
		imageName, err := utils.SaveImageFile(file, "vendors", fileHeader.Filename)
		if err != nil {
			utils.HandleErrors(w, http.StatusInternalServerError, "Error saving image")
		}
		vendor.Img = &imageName
	}

	query, args, err := QB.
		Insert("vendors").
		Columns("id", "img", "name", "description").
		Values(vendor.ID, vendor.Img, vendor.Name, vendor.Description).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(vendor_columns, ", "))).
		ToSql()
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error generate query")
		return
	}

	if err := db.QueryRowx(query, args...).StructScan(&vendor); err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error creating vendor"+err.Error())
		return
	}
	utils.JasonResponseHandler(w, http.StatusCreated, vendor)

}

func UpdateVendorHandler(w http.ResponseWriter, r *http.Request) {
	var vendor models.Vendors
	id := r.PathValue("id")
	query, args, err := QB.Select(vendor_columns...).
		From("vendors").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := db.Get(&vendor, query, args...); err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, err.Error())
		return
	}

	// update data
	if r.FormValue("name") != "" {
		vendor.Name = r.FormValue("name")
	}
	if r.FormValue("description") != "" {
		vendor.Description = r.FormValue("description")
	}

	query, args, err = QB.
		Update("vendors").
		Set("name", vendor.Name).
		Set("description", vendor.Description).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": vendor.ID}).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(vendor_columns, ", "))).
		ToSql()
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error building query")
		return
	}

	if err := db.QueryRowx(query, args...).StructScan(&vendor); err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error creating user"+err.Error())
		return
	}

	utils.JasonResponseHandler(w, http.StatusOK, vendor)
}

func DeleteVendorHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// Use QB to build the delete query
	query, args, err := QB.Delete("vendors").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error building query: "+err.Error())
		return
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error deleting vendor: "+err.Error())
		return
	}

	//	w.WriteHeader(http.StatusNoContent)
	utils.JasonResponseHandler(w, http.StatusGone, "deleted successfully")

}
