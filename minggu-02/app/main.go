package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Message    string `json:"message"`
	Host       string `json:"hostname"`
	Env        string `json:"environment"`
	DbUser     string `json:"db_user"`
	LogPath    string `json:"log_path"`
	Time       string `json:"current_time"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appEnv := os.Getenv("APP_ENV")
	dbUser := os.Getenv("DB_USER")
	forceUnready := os.Getenv("FORCE_UNREADY")

	// Endpoint Utama
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hostname, _ := os.Hostname()
		res := Response{
			Message: "Hello from Production-Ready Go App on K8s!",
			Host:    hostname,
			Env:     appEnv,
			DbUser:  dbUser,
			LogPath: "/data/app.log",
			Time:    time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	// Liveness Probe Endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Readiness Probe Endpoint
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if forceUnready == "true" {
			log.Println("Simulasi: Readiness check GAGAL karena FORCE_UNREADY=true")
			http.Error(w, "Service Unavailable (Simulated Failure)", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	// Endpoint Menulis ke PVC Storage
	http.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		logFile := "/data/app.log"
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			http.Error(w, fmt.Sprintf("Gagal menulis ke PVC: %v", err), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		entry := fmt.Sprintf("[%s] Log entry created by host %s\n", time.Now().Format(time.RFC3339), os.Getenv("HOSTNAME"))
		f.WriteString(entry)
		w.Write([]byte("Berhasil menulis ke PVC storage: " + entry))
	})

	log.Printf("Server berjalan di port %s ...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server crash: %v", err)
	}
}
