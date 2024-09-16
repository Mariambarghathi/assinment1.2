package controller

import (
	utils "backend-project/Utils"
	"backend-project/models"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/Masterminds/squirrel"
)

var (
	Domain = os.Getenv("DOMAIN")

	user_columns = []string{
		"id",
		"name",
		"email",
		"phone",
		"created_at",
		"updated_at",
		fmt.Sprintf("CASE WHEN NULLIF(img, '') IS NOT NULL THEN FORMAT('%s/%%s', img) ELSE NULL END AS img", Domain),
	}
)

// database connection
var db *sqlx.DB
var QB = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func SetDB(database *sqlx.DB) {
	db = database
}

func IndexUserHandler(w http.ResponseWriter, r *http.Request) {

	var users []models.User

	query, args, err := QB.Select(strings.Join(user_columns, ", ")).
		From("users").
		ToSql()
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := db.Select(&users, query, args...); err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JasonResponseHandler(w, http.StatusInternalServerError, users)
}

func ShowUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	id := r.PathValue("id")
	query, args, err := QB.Select(strings.Join(user_columns, ", ")).
		From("users").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := db.Get(&user, query, args...); err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JasonResponseHandler(w, http.StatusOK, user)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	id := r.PathValue("id")
	query, args, err := QB.Select(user_columns...).
		From("users").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := db.Get(&user, query, args...); err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, err.Error())
		return
	}

	// update data
	if r.FormValue("name") != "" {
		user.Name = r.FormValue("name")
	}
	if r.FormValue("phone") != "" {
		user.Phone = r.FormValue("phone")
	}
	if r.FormValue("email") != "" {
		user.Email = r.FormValue("email")
	}
	if r.FormValue("password") != "" {
		hashedPassword, err := utils.HashPassword(user.Password)
		if err != nil {
			utils.HandleErrors(w, http.StatusInternalServerError, "Error hashing password")
			return
		}
		user.Password = hashedPassword
	}

	query, args, err = QB.
		Update("users").
		Set("name", user.Name).
		Set("email", user.Email).
		Set("phone", user.Phone).
		Set("password", user.Password).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": user.ID}).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(user_columns, ", "))).
		ToSql()
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error building query")
		return
	}

	if err := db.QueryRowx(query, args...).StructScan(&user); err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error creating user"+err.Error())
		return
	}

	utils.JasonResponseHandler(w, http.StatusOK, user)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// Use QB to build the delete query
	query, args, err := QB.Delete("users").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error building query: "+err.Error())
		return
	}

	if err := db.QueryRow(query, args...).Scan(&id); err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error deleting user: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
