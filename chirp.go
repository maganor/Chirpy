package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maganor/Chirpy/internal/auth"
	"github.com/maganor/Chirpy/internal/database"
)

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func cleanBody(body string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	for _, word := range badWords {
		body = strings.ReplaceAll(body, word, "****")
		body = strings.ReplaceAll(body, strings.ToUpper(word), "****")
		body = strings.ReplaceAll(body, strings.ToUpper(string(word[0]))+word[1:], "****")
	}
	return body
}

func CreateChirp(res http.ResponseWriter, req *http.Request) {
	type retError struct {
		Err string `json:"error"`
	}

	type params struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	parameters := params{}
	respError := retError{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&parameters)
	if err != nil {
		respError.Err = "Something went wrong"
		dat, _ := json.Marshal(respError)
		res.WriteHeader(400)
		res.Write(dat)
		return
	}
	res.Header().Set("Content-Type", "application/json")

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respError.Err = "Unauthorized"
		dat, _ := json.Marshal(respError)
		res.WriteHeader(401)
		res.Write(dat)
		return
	}

	fmt.Println("Authorizing token", token)

	userID, err := auth.ValidateJWT(token, apiCfg.jwt_token)
	if err != nil {
		fmt.Println("Error validating token", err)
		respError.Err = "Unauthorized"
		dat, _ := json.Marshal(respError)
		res.WriteHeader(401)
		res.Write(dat)
		return
	}

	if userID.String() == "" {
		respError.Err = "Unauthorized"
		dat, _ := json.Marshal(respError)
		res.WriteHeader(401)
		res.Write(dat)
		return
	}

	if len(parameters.Body) <= 140 {
		parameters.Body = cleanBody(parameters.Body)
		chirpResp, err := apiCfg.queries.CreateChirp(req.Context(), database.CreateChirpParams{Body: parameters.Body, UserID: userID})
		if err != nil {
			respError.Err = "Error creating chirp"
			dat, _ := json.Marshal(respError)
			res.WriteHeader(400)
			res.Write(dat)
			return
		}
		chirpJSON := chirp{
			ID:        chirpResp.ID.UUID,
			CreatedAt: chirpResp.CreatedAt,
			UpdatedAt: chirpResp.UpdatedAt,
			Body:      chirpResp.Body,
			UserId:    chirpResp.UserID,
		}
		res.WriteHeader(201)
		dat, _ := json.Marshal(chirpJSON)
		res.Write(dat)
		return
	}

	respError.Err = "Chirp is too long"
	res.WriteHeader(400)
	dat, _ := json.Marshal(respError)
	res.Write(dat)
}

func GetChirps(res http.ResponseWriter, req *http.Request) {
	chirps, err := apiCfg.queries.GetChirps(req.Context())
	chirpsJSON := []chirp{}
	for _, dbChirp := range chirps {
		chirpsJSON = append(chirpsJSON, chirp{
			ID:        dbChirp.ID.UUID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserId:    dbChirp.UserID,
		})
	}
	if err != nil {
		fmt.Println(err)
		res.WriteHeader(400)
		res.Write([]byte("Something went wrong"))
		return
	}
	res.WriteHeader(200)
	dat, _ := json.Marshal(chirpsJSON)
	res.Write(dat)
}

func GetChirp(res http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		fmt.Println(err)
		res.WriteHeader(400)
		res.Write([]byte("Something went wrong"))
		return
	}
	chirpReturn, err := apiCfg.queries.GetChirp(req.Context(), uuid.NullUUID{UUID: parsedUUID, Valid: true})
	chirpJSON := chirp{
		ID:        chirpReturn.ID.UUID,
		CreatedAt: chirpReturn.CreatedAt,
		UpdatedAt: chirpReturn.UpdatedAt,
		Body:      chirpReturn.Body,
		UserId:    chirpReturn.UserID,
	}
	if err != nil {
		fmt.Println(err)
		res.WriteHeader(400)
		res.Write([]byte("Something went wrong"))
		return
	}
	res.WriteHeader(200)
	dat, _ := json.Marshal(chirpJSON)
	res.Write(dat)
}
