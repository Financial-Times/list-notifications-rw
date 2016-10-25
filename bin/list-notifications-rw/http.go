package main

import (
	"github.com/Financial-Times/list-notifications-rw/bin/list-notifications-rw/resources"
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"github.com/Sirupsen/logrus"
	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Financial-Times/list-notifications-rw/mapping"
)

func server(mapper mapping.NotificationsMapper, db db.DB){
	r := mux.NewRouter()
	r.HandleFunc("/lists/notifications", resources.ReadNotifications(mapper, db))
	r.HandleFunc("/lists/notifications/{uuid}", resources.FilterSyntheticTransactions(resources.WriteNotification(mapper, db))).Methods("PUT")

	server := &http.Server{
		Handler:      r,
		Addr:         ":8080",

		WriteTimeout: 60 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logrus.Info("Starting server on :8080")
	server.ListenAndServe()
}