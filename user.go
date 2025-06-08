package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/maganor/Chirpy/internal/auth"
	"github.com/maganor/Chirpy/internal/database"
)

type userResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func CreateUser(res http.ResponseWriter, req *http.Request) {
	type userReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	body := userReq{}
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		http.Error(res, "Something went wrong", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPasword(body.Password)
	if err != nil {
		http.Error(res, "Something went wrong", http.StatusBadRequest)
		return
	}
	userResp, err := apiCfg.queries.CreateUser(req.Context(), database.CreateUserParams{
		Email:          body.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		http.Error(res, "Something went wrong", http.StatusBadRequest)
		return
	}

	userJSON := userResponse{
		ID:        userResp.ID,
		CreatedAt: userResp.CreatedAt,
		UpdatedAt: userResp.UpdatedAt,
		Email:     userResp.Email,
	}

	res.WriteHeader(201)
	dat, _ := json.Marshal(userJSON)
	res.Write(dat)
}

func Login(res http.ResponseWriter, req *http.Request) {
	type userReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	//Get Request Body
	body := userReq{}
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		println("error decoding body")
		http.Error(res, "Something went wrong", http.StatusUnauthorized)
		return
	}

	//Get User By Email
	user, err := apiCfg.queries.GetUserByEmail(req.Context(), body.Email)

	if err != nil {
		println("error getting user by email")
		http.Error(res, "Something went wrong", http.StatusUnauthorized)
		return
	}

	//Check if password is correct
	err = auth.CheckPasswordHash(user.HashedPassword, body.Password)

	if err != nil {
		println("error checking password hash")
		http.Error(res, "Something went wrong", http.StatusUnauthorized)
		return
	}

	//Setup JWT
	token, err := auth.MakeJWT(user.ID, apiCfg.jwt_token, time.Hour)
	if err != nil {
		http.Error(res, "Something went wrong", http.StatusBadRequest)
		return
	}

	//Setup Refresh Token
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		http.Error(res, "Something went wrong", http.StatusUnauthorized)
		return
	}

	apiCfg.queries.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	})

	//Generate Response
	userJSON := userResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}
	dat, _ := json.Marshal(userJSON)
	res.WriteHeader(200)
	res.Write(dat)

}

func RefreshUser(res http.ResponseWriter, req *http.Request) {
	type Res struct {
		Token string `json:"token"`
	}
	resp := Res{}
	//Get Refresh Token
	token, _ := auth.GetBearerToken(req.Header)
	tokenDb, err := apiCfg.queries.GetRefreshToken(req.Context(), token)
	if err != nil {
		http.Error(res, "Something went wrong", http.StatusUnauthorized)
		return
	}
	//Setup JWT
	if tokenDb.ExpiresAt.After(time.Now()) && !tokenDb.RevokedAt.Valid {
		tokenJwt, err := auth.MakeJWT(tokenDb.UserID, apiCfg.jwt_token, time.Hour)
		if err != nil {
			http.Error(res, "Something went wrong", http.StatusBadRequest)
			return
		}
		resp.Token = tokenJwt
		dat, _ := json.Marshal(resp)
		res.WriteHeader(200)
		res.Write(dat)
		return
	}
	res.WriteHeader(401)
}

func RevokeToken(res http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		http.Error(res, "Error Revoking token", http.StatusUnauthorized)
		return
	}
	err = apiCfg.queries.RevokeRefreshToken(req.Context(), token)
	if err != nil {
		http.Error(res, "Error Revoking token", http.StatusUnauthorized)
		return
	}
	res.WriteHeader(204)
}
