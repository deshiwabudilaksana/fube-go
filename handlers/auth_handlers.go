package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/db"
	"github.com/deshiwabudilaksana/fube-go/services/auth"
)

func getAuthService(ctx context.Context) (*auth.Service, error) {
	cfg := config.Load()
	database, err := db.GetDatabase(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return auth.NewService(database), nil
}

func ShowLogin(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"templates/auth_layout.html",
		"templates/login_fragment.html",
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template parse error: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "auth_layout.html", nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
	}
}

func PostLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	authSvc, err := getAuthService(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	token, err := authSvc.LoginUser(r.Context(), email, password)
	if err != nil {
		// Return error message for HTMX
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Invalid credentials")
		return
	}

	// Set HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to dashboard using HTMX header
	w.Header().Set("HX-Redirect", "/yield-planning")
	w.WriteHeader(http.StatusOK)
}

func ShowRegister(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"templates/auth_layout.html",
		"templates/register_fragment.html",
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template parse error: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "auth_layout.html", nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
	}
}

func PostRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	params := auth.RegisterParams{
		Username:    r.FormValue("username"),
		Email:       r.FormValue("email"),
		Password:    r.FormValue("password"),
		CompanyName: r.FormValue("company_name"),
	}

	authSvc, err := getAuthService(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = authSvc.RegisterUser(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Registration failed: %v", err)
		return
	}

	// After registration, log them in
	token, err := authSvc.LoginUser(r.Context(), params.Email, params.Password)
	if err != nil {
		// Should not happen after successful register, but handle it
		w.Header().Set("HX-Redirect", "/login")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("HX-Redirect", "/yield-planning")
	w.WriteHeader(http.StatusOK)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusOK)
}
