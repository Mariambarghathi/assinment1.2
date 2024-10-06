package controller

import (
	utils "backend-project/Utils"
	"backend-project/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Enable CORS
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")                                // Allows all origins; adjust if needed
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE") // Specify allowed methods
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")     // Specify allowed headers
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	user := models.User{
		ID:       uuid.New(),
		Name:     r.FormValue("name"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}
	if user.Password == "" {
		utils.HandleErrors(w, http.StatusBadRequest, "Password is required")
		return
	}

	file, fileHeader, err := r.FormFile("img")
	if err != nil && err != http.ErrMissingFile {
		utils.HandleErrors(w, http.StatusBadRequest, "Invalid file")
		return
	} else if err == nil {
		defer file.Close()
		imageName, err := utils.SaveImageFile(file, "users", fileHeader.Filename)
		if err != nil {
			utils.HandleErrors(w, http.StatusInternalServerError, "Error saving image")
		}
		user.Img = &imageName
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error hashing password")
		return
	}
	user.Password = hashedPassword

	query, args, err := QB.
		Insert("users").
		Columns("id", "img", "name", "phone", "email", "password").
		Values(user.ID, user.Img, user.Name, user.Phone, user.Email, user.Password).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(user_columns, ", "))).
		ToSql()
	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error generate query")
		return
	}

	if err := db.QueryRowx(query, args...).StructScan(&user); err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error creating user"+err.Error())
		return
	}
	utils.JasonResponseHandler(w, http.StatusCreated, user)

}

func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Header.Get("Content-Type") != "application/json" {
		utils.HandleErrors(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
		utils.HandleErrors(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	email := loginData.Email
	password := loginData.Password

	if email == "" || password == "" {
		utils.HandleErrors(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Query to find user by email
	var user models.User
	query, args, err := QB.Select("id", "name", "phone", "email", "password").
		From("users").
		Where("email = ?", email).
		ToSql()

	if err != nil {
		utils.HandleErrors(w, http.StatusInternalServerError, "Error generating query")
		return
	}

	if err := db.Get(&user, query, args...); err != nil {
		utils.HandleErrors(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		utils.HandleErrors(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	utils.JasonResponseHandler(w, http.StatusOK, user)
}
